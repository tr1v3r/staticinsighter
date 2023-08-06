package analyzer

import (
	"runtime"
	"sync"
	"time"

	"github.com/riverchu/pkg/pools"
	"golang.org/x/tools/go/ssa"
)

// CollectRiskyHandlers collect risky handlers
func (a *Analyzer) CollectRiskyHandlers(prog *ssa.Program) ([]*ssa.Function, error) {
	a.logger.CtxInfo(a.ctx, "collecting risky handlers...")
	defer func(s time.Time) { a.logger.CtxInfo(a.ctx, "collect risky handlers cost: %s", time.Since(s)) }(time.Now())

	var dependencies = make(map[string]*ssa.Package, 1024)

	// collect main packages and dependencies
	var mainPkgs []*ssa.Package
	for _, pkg := range prog.AllPackages() {
		if pkg.Pkg.Name() != "main" {
			dependencies[pkg.Pkg.Path()] = pkg
			continue
		}

		a.logger.CtxInfo(a.ctx, "find main package path: %s", pkg.Pkg.Path())
		mainPkgs = append(mainPkgs, pkg)
	}

	// collect init functions
	a.logger.CtxInfo(a.ctx, "collecting main and init...")
	s := time.Now()
	var entries []*ssa.Function
	for _, main := range mainPkgs {
		entries = append(entries, a.collectMainAndInit(main, dependencies)...)
	}
	a.logger.CtxInfo(a.ctx, "collect main and init (%d) cost: %s", len(entries), time.Since(s))

	if a.CheckMode(ModeDebug) {
		for _, fn := range entries {
			if a.MatchMain(fn) {
				a.AddActiveMain(fn)
			} else {
				a.AddActiveInit(fn)
			}
		}
	}

	s = time.Now()
	a.logger.CtxInfo(a.ctx, "collecting active handlers...")
	handlers := a.collectActiveHandlers(entries...)
	a.logger.CtxInfo(a.ctx, "collect active handlers (%d) cost: %s", len(handlers), time.Since(s))

	riskyHandlers := a.filterRiskyHandlers(handlers...)
	if a.CheckMode(ModeDebug) {
		a.AddActiveHandler(handlers...)
		a.AddRiskyHandler(riskyHandlers...)
	}
	return riskyHandlers, nil
}

// collectMainAndInit collect main and init functions
func (a *Analyzer) collectMainAndInit(pkg *ssa.Package, dependencies map[string]*ssa.Package) (funcs []*ssa.Function) {
	a.logger.CtxTrace(a.ctx, "collecting main and init functions in package %s", pkg.Pkg.Path())
	for _, member := range pkg.Members {
		fn, ok := member.(*ssa.Function)
		if !ok || fn.Blocks == nil { // ignore not function or nil function
			continue
		}

		switch {
		case a.MatchInit(fn):
			funcs = append(funcs, fn)
		case a.MatchMain(fn):
			funcs = append([]*ssa.Function{fn}, funcs...)
		}
	}

	for _, pkg := range pkg.Pkg.Imports() {
		dep, ok := dependencies[pkg.Path()]
		if !ok {
			a.logger.CtxWarn(a.ctx, "dependence package %s not found", pkg.Path())
			continue
		}
		if dep == nil {
			continue
		}
		dependencies[pkg.Path()] = nil // mark as visited
		funcs = append(funcs, a.collectMainAndInit(dep, dependencies)...)
	}
	return
}

// collectActiveHandlers collect handlers
func (a *Analyzer) collectActiveHandlers(entries ...*ssa.Function) (handlers []*ssa.Function) {
	for _, entry := range entries {
		a.visitInstr(func(instr ssa.Instruction) (stop bool) {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					a.logger.CtxTrace(a.ctx, "collect handler from %s -> %s", instr.Parent().Name(), callee.Name())

					if a.MatchHandler(callee) {
						handlers = append(handlers, callee)
						return false
					}

					for _, arg := range i.Call.Args {
						handlers = append(handlers, a.matchFunc(arg, a.MatchHandler))
					}
					handlers = append(handlers, a.collectActiveHandlers(callee)...)
				}
			default:
				handlers = append(handlers, a.matchFunc(i, a.MatchHandler))
			}
			return false
		}, entry)
	}
	return a.uniq(handlers)
}

func (a *Analyzer) filterRiskyHandlers(handlers ...*ssa.Function) []*ssa.Function {
	a.logger.CtxInfo(a.ctx, "filtering risky handlers...")
	defer func(s time.Time) { a.logger.CtxInfo(a.ctx, "filter risky handlers cost: %s", time.Since(s)) }(time.Now())

	concurrency := runtime.NumCPU()
	pool := pools.NewPool(concurrency)

	a.logger.CtxInfo(a.ctx, "filter risky handlers in %d goroutines", concurrency)

	var mu sync.Mutex
	riskyHandlers := make([]*ssa.Function, 0, len(handlers))

	for _, handler := range handlers {
		pool.Wait()
		go func(handler *ssa.Function) {
			defer pool.Done()
			r := NewRisk(handler)
			a.visitInstr(a.cgVisitor(handler, r), handler)
			r.Finish()

			if !r.Risky() {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			riskyHandlers = append(riskyHandlers, handler)
		}(handler)
	}
	pool.WaitAll()

	return riskyHandlers
}

type Visitor func(ssa.Instruction) (stop bool)

func (a *Analyzer) cgVisitor(entry *ssa.Function, p *RiskInfo) Visitor {
	return func(instr ssa.Instruction) (stop bool) {
		switch i := instr.(type) {
		case *ssa.Call:
			if callee := i.Call.StaticCallee(); callee != nil {
				var r = a.GetRisk(callee)
				if r == nil { // has no record
					switch {
					case a.MatchSource(callee):
						a.AddRisk(NewRiskSource(callee)) // avoid data race
						r = a.GetRisk(callee)
					case a.MatchSink(callee):
						a.AddRisk(NewRiskSink(callee)) // avoid data race
						r = a.GetRisk(callee)
					default:
						a.AddRisk(NewRisk(callee))
						r = a.GetRisk(callee)

						a.visitInstr(a.cgVisitor(entry, r), callee)
						r.Finish()
					}
				} else if !r.Done() && !r.Collected(entry) { // has record and not finished and not collected
					r.RecordEntry(entry)

					a.visitInstr(a.cgVisitor(entry, r), callee)
					r.Finish()
				}

				if r.HasSource() || r.IsSource() {
					p.AddSoucre(r)
				}
				if r.HasSink() || r.IsSink() {
					p.AddSink(r)
				}
				// if detect risky, stop visit
				if p.Risky() {
					return true
				}
			}
		}
		return false
	}
}

func (*Analyzer) visitInstr(visit Visitor, fn *ssa.Function) {
	if fn == nil || fn.Blocks == nil {
		return
	}
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			if visit(instr) {
				return
			}
		}
	}
}

func (a *Analyzer) matchFunc(i any, match func(*ssa.Function) bool) *ssa.Function {
	switch i := i.(type) {
	case *ssa.Store:
		switch x := i.Val.(type) {
		case *ssa.ChangeType:
			return a.matchFunc(x, match)
		}
	case *ssa.ChangeType:
		if x, ok := i.X.(*ssa.Function); ok {
			if match(x) {
				return x
			}
		}
	}
	return nil
}
