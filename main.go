package main

import (
	"context"
	"fmt"
	"go/token"
	"log"
	"path/filepath"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func main() {
	path := "/tmp/hello/..."
	path, _ = filepath.Abs(path)
	dir := filepath.Dir(path)

	ctx := context.Background()
	fset := token.NewFileSet()

	cfg := &packages.Config{
		Mode:    ^0,
		Context: ctx,
		Fset:    fset,
		Dir:     dir,
	}
	initial, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatalf("load package fail: %v", err)
	}

	if packages.PrintErrors(initial) > 0 {
		log.Fatalf("find errors")
		return
	}

	prog, _ := ssautil.Packages(initial, ssa.GlobalDebug|ssa.BareInits)
	prog.Build()

	for _, pkg := range prog.AllPackages() {
		analyzePackage(pkg)
	}
}

func analyzePackage(pkg *ssa.Package) {
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
					fmt.Printf("Function %s calls %s\n", fn.Name(), callee.Name())
				}
			}
		}
	}
}
