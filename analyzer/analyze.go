// Package analyzer ...
package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"path/filepath"
	"time"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func Analyze(path string) error {
	ctx := context.Background()

	prog, err := buildProgram(ctx, path)
	if err != nil {
		return fmt.Errorf("build program fail: %w", err)
	}

	fmt.Println("analyzing all packages...")
	defer func(s time.Time) { fmt.Printf("analyze all packaegs cost: %s\n", time.Since(s)) }(time.Now())

	for _, pkg := range prog.AllPackages() {
		analyzePackage(pkg)
	}
	return nil
}

func buildProgram(ctx context.Context, path string) (*ssa.Program, error) {
	fmt.Println("building ssa program...")
	defer func(s time.Time) { fmt.Printf("build ssa program cost: %s\n", time.Since(s)) }(time.Now())

	path, _ = filepath.Abs(path)
	dir := filepath.Dir(path)

	fset := token.NewFileSet()

	cfg := &packages.Config{
		Mode:    ^0,
		Context: ctx,
		Fset:    fset,
		Dir:     dir,
	}
	initial, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("load package fail: %w", err)
	}

	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("find errors when load packages")
	}

	prog, _ := ssautil.Packages(initial, ssa.GlobalDebug|ssa.BareInits)
	prog.Build()

	return prog, nil
}

func analyzePackage(pkg *ssa.Package) {
	fmt.Printf("analyzing package (%s)...\n", pkg.Pkg.Path())
	defer func(s time.Time) { fmt.Printf("analyze package (%s) cost: %s\n", pkg.Pkg.Path(), time.Since(s)) }(time.Now())

	for _, member := range pkg.Members {
		if fn, ok := member.(*ssa.Function); ok {
			analyzeFunction(fn)
		}
	}
}

func analyzeFunction(fn *ssa.Function) {
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					fmt.Printf("call %s -> %s\n", fn.Name(), callee.Name())
				}
			}
		}
	}
}
