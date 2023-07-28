package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// defaultAnalyzer default analyzer
var defaultAnalyzer = NewAnalyzer(context.Background())

// Analyzer static code analyzer
type Analyzer struct {
	ctx context.Context

	*Configure
}

// WithContext set analyzer context
func (a Analyzer) WithContext(ctx context.Context) *Analyzer {
	a.ctx = ctx
	return &a
}

// WithConfig set analyzer config
func (a Analyzer) WithConfig() *Analyzer { return &a }

// Analyze analyze project
func (a *Analyzer) Analyze(paths ...string) error {
	defer a.recover()

	path, _ := a.parsePaths(paths)

	a.logger.Info("analyzing path %s", path)
	defer func(s time.Time) { a.logger.Info("analyze program cost: %s", time.Since(s)) }(time.Now())

	prog, err := a.buildProgram(path)
	if err != nil {
		return fmt.Errorf("build program fail: %w", err)
	}

	funcs, err := a.MatchFunctions(prog)
	if err != nil {
		return fmt.Errorf("match functions fail: %w", err)
	}

	if a.CheckMode(ModeDebug) {
		a.PrintFuncs(funcs)
	}

	if len(funcs.handlerFuncs) == 0 {
		a.logger.Info("handlers not found")
		return nil
	}

	a.analyzeFunctions(funcs)

	// a.logger.Info("analyzing all packages...")
	// defer func(s time.Time) { a.logger.Info("analyze all packages cost: %s", time.Since(s)) }(time.Now())
	// for _, pkg := range prog.AllPackages() {
	// 	a.analyzePackage(pkg)
	// }
	return nil
}

func (a *Analyzer) buildProgram(path string) (*ssa.Program, error) {
	a.logger.Info("building ssa program...")
	defer func(s time.Time) { a.logger.Info("build ssa program cost: %s", time.Since(s)) }(time.Now())

	initial, err := a.loadAST(path)
	if err != nil {
		return nil, fmt.Errorf("load packages fail: %w", err)
	}

	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("find errors when load packages")
	}

	var packages func(initial []*packages.Package, mode ssa.BuilderMode) (*ssa.Program, []*ssa.Package)
	switch {
	case a.CheckMode(ModeUltimate):
		packages = ssautil.AllPackages
	default:
		packages = ssautil.Packages
	}
	prog, _ := packages(initial, ssa.GlobalDebug|ssa.BareInits)
	prog.Build()

	return prog, nil
}

func (a *Analyzer) loadAST(path string) ([]*packages.Package, error) {
	var err error
	path, err = filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if !strings.HasSuffix(path, "/...") {
		path = filepath.Join(path, "...")
	}

	return packages.Load(&packages.Config{
		Mode:    ^0,
		Context: a.ctx,
		Fset:    token.NewFileSet(),
		Dir:     filepath.Dir(path),
	}, path)
}

func (a *Analyzer) MatchFunctions(prog *ssa.Program) (*Functions, error) {
	a.logger.Info("matching functions...")
	defer func(s time.Time) { a.logger.Info("match all functions cost: %s", time.Since(s)) }(time.Now())
	return NewFuncitons().Match(prog), nil
}

// func (a *Analyzer) analyzePackage(pkg *ssa.Package) {
// 	a.logger.Trace("analyzing package (%s)...", pkg.Pkg.Path())
// 	defer func(s time.Time) { a.logger.Trace("analyze package (%s) cost: %s", pkg.Pkg.Path(), time.Since(s)) }(time.Now())

// 	path := pkg.Pkg.Path()
// 	if firstPath := strings.Split(path, "/")[0]; strings.Contains(firstPath, ".") {
// 		a.logger.Trace("find non-built-in package: %s", path)
// 	}

// 	for _, member := range pkg.Members {
// 		if fn, ok := member.(*ssa.Function); ok {
// 			a.analyzeFunction(fn)
// 		}
// 	}
// }

func (a *Analyzer) analyzeFunctions(funcs *Functions) {
	// analyze init and main functions
	// find all handlers
	a.analyzeInitAndMain(funcs)

	// deep visit handler
	// remove nested handlers
	// detect if handler cg has source and sink

	// find taint code flow
}

func (a *Analyzer) analyzeInitAndMain(funcs *Functions) (handlers []*ssa.Function) {
	for fn := range funcs.initFuncs {
		handlers = append(handlers, a.findHandler(funcs, fn)...)
	}
	for fn := range funcs.mainFuncs {
		handlers = append(handlers, a.findHandler(funcs, fn)...)
	}
	return
}

func (a *Analyzer) analyzeHandlers(funcs *Functions) {
	for fn := range funcs.mainFuncs {
		_ = fn
	}
}

func (a *Analyzer) findHandler(funcs *Functions, fn *ssa.Function) (handlers []*ssa.Function) {
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					// a.logger.Debug("call %s -> %s", fn.Name(), callee.Name())
					for _, arg := range i.Call.Args {
						// a.logger.Debug("call %s -> %s, %s, %s", callee.Name(), arg.Name(), arg.String(), arg.Type())
						if sl, ok := arg.(*ssa.Slice); ok {
							_ = sl
						}
						if ct, ok := arg.(*ssa.ChangeType); ok {
							if fn, ok := ct.X.(*ssa.Function); ok {
								if funcs.hasHandler(fn) {
									a.logger.Info("find active handler: %s", fn)
								}
							}
						}
						if fn, ok := arg.(*ssa.Function); ok {
							fmt.Printf("\tFound function: %v\n", fn.String())
						}
						if mi, ok := arg.(*ssa.MakeInterface); ok {
							if fn, ok := mi.X.(*ssa.Function); ok {
								fmt.Printf("\tFound function: %v\n", fn.String())
							}
						}
					}

					a.findHandler(funcs, callee)
				}
			case *ssa.Slice:
				i.Name()
			}
		}
	}
	return
}

func (a *Analyzer) analyzeFunction(fn *ssa.Function) {
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					a.logger.Debug("call %s -> %s", fn.Name(), callee.Name())
					a.analyzeFunction(callee)
				}
			}
		}
	}
}

func (*Analyzer) parsePaths(paths []string) (path, entry string) {
	switch len(paths) {
	case 1:
		return paths[0], ""
	case 2:
		return paths[0], paths[1]
	default:
		return "", ""
	}
}

func (a *Analyzer) recover() {
	if e := recover(); e != nil {
		a.logger.CtxPanic(a.ctx, "analyze panic: %s\n%s", e, CatchStack())
	}
}
