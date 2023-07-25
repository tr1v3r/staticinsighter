// Package analyzer ...
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

	funcs map[*ssa.Function]bool
}

// WithContext set analyzer context
func (a Analyzer) WithContext(ctx context.Context) *Analyzer {
	a.ctx = ctx
	return &a
}

// WithConfig set analyzer config
func (a Analyzer) WithConfig() *Analyzer {
	return &a
}

// Analyze analyze project
func (a *Analyzer) Analyze(path string) error {
	a.logger.Info("analyzing path %s", path)
	defer func(s time.Time) { a.logger.Info("analyze program cost: %s", time.Since(s)) }(time.Now())

	prog, err := a.buildProgram(path)
	if err != nil {
		return fmt.Errorf("build program fail: %w", err)
	}

	a.logger.Info("analyzing functions...")
	defer func(s time.Time) { a.logger.Info("analyze all functions cost: %s", time.Since(s)) }(time.Now())
	a.matchFunctions(prog)

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
	path, _ = filepath.Abs(path)
	dir := filepath.Dir(path)
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode:    ^0,
		Context: a.ctx,
		Fset:    fset,
		Dir:     dir,
	}
	return packages.Load(cfg, path)
}

func (a *Analyzer) matchFunctions(prog *ssa.Program) {
	funcs := ssautil.AllFunctions(prog)
	for fn := range funcs {
		if MatchHandler(fn.Signature.Params().String(), fn.Signature.Results().String()) {
			path, funcName := fn.Pkg.Pkg.Path(), fn.Name()
			a.logger.Debug("match handler: %s.%s%s %s", path, funcName, fn.Signature.Params().String(), fn.Signature.Results().String())
			a.analyzeFunction(fn)
		}
	}
}

func (a *Analyzer) analyzePackage(pkg *ssa.Package) {
	a.logger.Trace("analyzing package (%s)...", pkg.Pkg.Path())
	defer func(s time.Time) { a.logger.Trace("analyze package (%s) cost: %s", pkg.Pkg.Path(), time.Since(s)) }(time.Now())

	path := pkg.Pkg.Path()
	if firstPath := strings.Split(path, "/")[0]; strings.Contains(firstPath, ".") {
		a.logger.Trace("find non-built-in package: %s", path)
	}

	for _, member := range pkg.Members {
		if fn, ok := member.(*ssa.Function); ok {
			a.analyzeFunction(fn)
		}
	}
}

func (a *Analyzer) analyzeFunction(fn *ssa.Function) {
	if a.funcs[fn] {
		return
	}
	a.funcs[fn] = true

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
