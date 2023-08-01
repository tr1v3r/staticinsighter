package analyzer

import (
	"time"

	"golang.org/x/tools/go/ssa"
)

// CollectRiskyHandlers collect risky handlers
func (a *Analyzer) CollectRiskyHandlers(prog *ssa.Program) ([]*ssa.Function, error) {
	a.logger.Info("collecting handlers...")
	defer func(s time.Time) { a.logger.Info("collect handlers cost: %s", time.Since(s)) }(time.Now())

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
	var entries []*ssa.Function
	for _, main := range mainPkgs {
		entries = append(entries, a.collectMainAndInit(main, dependencies)...)
	}

	if a.CheckMode(ModeDebug) {
		for _, fn := range entries {
			if a.MatchMain(fn) {
				a.AddActiveMain(fn)
			} else {
				a.AddActiveInit(fn)
			}
		}
	}

	handlers := a.collectHandlers(entries...)
	riskyHandlers := a.collectRiskyHandlers(handlers...)
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

// collectHandlers collect handlers
func (a *Analyzer) collectHandlers(entries ...*ssa.Function) (handlers []*ssa.Function) {
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
					handlers = append(handlers, a.collectHandlers(callee)...)
				}
			default:
				handlers = append(handlers, a.matchFunc(i, a.MatchHandler))
			}
			return false
		}, entry)
	}
	return a.uniq(handlers)
}

func (a *Analyzer) collectRiskyHandlers(handlers ...*ssa.Function) []*ssa.Function {
	riskyHandlers := make([]*ssa.Function, 0, len(handlers))
	for _, handler := range handlers {
		var hasSource, hasSink bool

		var visit func(instr ssa.Instruction) (stop bool)
		visit = func(instr ssa.Instruction) (stop bool) {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					if a.MatchSource(callee) {
						hasSource = true
					} else if a.MatchSink(callee) {
						hasSink = true
					}

					for _, arg := range i.Call.Args {
						if a.matchFunc(arg, a.MatchSource) != nil {
							hasSource = true
						} else if a.matchFunc(arg, a.MatchSink) != nil {
							hasSink = true
						}
					}
					if hasSource && hasSink {
						return true
					}

					a.visitInstr(visit, callee)
				}
			default:
				if a.matchFunc(i, a.MatchSource) != nil {
					hasSource = true
				} else if a.matchFunc(i, a.MatchSink) != nil {
					hasSink = true
				}
			}
			return hasSource && hasSink
		}

		// detect source and sink
		a.visitInstr(visit, handler)

		if hasSource && hasSink {
			riskyHandlers = append(riskyHandlers, handler)
		} else if a.CheckMode(ModeDebug) {
			a.logger.CtxDebug(a.ctx, "active no risk handler: %s", handler.Name())
		}
	}
	return riskyHandlers
}

func (*Analyzer) visitInstr(visit func(ssa.Instruction) (stop bool), fn *ssa.Function) {
	if fn.Blocks == nil {
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
