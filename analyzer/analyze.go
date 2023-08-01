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
var defaultAnalyzer, _ = NewAnalyzer(context.Background(), NewConfigure().Init())

// NewAnalyzer build a new analyzer
func NewAnalyzer(ctx context.Context, cfg *Configure) (*Analyzer, error) {
	matcher := NewMatcher()
	if err := matcher.LoadRules(cfg.HandlerSigRules, cfg.SourceSigRules, cfg.SinkSigRules); err != nil {
		return nil, fmt.Errorf("load rule fail: %w", err)
	}

	return &Analyzer{
		ctx: ctx,

		Configure: cfg,
		Matcher:   matcher,

		Functions: NewFuncitons(),
		chains:    nil,
	}, nil
}

// Analyzer static code analyzer
type Analyzer struct {
	ctx context.Context

	*Configure
	*Matcher

	*Functions
	chains []*Chain
}

// WithContext set analyzer context
func (a Analyzer) WithContext(ctx context.Context) *Analyzer {
	a.ctx = ctx
	return &a
}

// Analyze analyze project
func (a *Analyzer) Analyze(paths ...string) error {
	defer a.recover()

	path, _ := a.parsePaths(paths)

	a.logger.Info("analyzing path %s", path)
	defer func(s time.Time) { a.logger.Info("analyze program cost: %s", time.Since(s)) }(time.Now())

	// build SSA program
	prog, err := a.buildProgram(path)
	if err != nil {
		return fmt.Errorf("build program fail: %w", err)
	}

	// collect info
	handlers, err := a.CollectRiskyHandlers(prog)
	if err != nil {
		return fmt.Errorf("analyze packages fail: %w", err)
	}

	if a.CheckMode(ModeDebug) {
		a.printAllUsefulFunc(prog)
	}

	// analyze taint chains
	chains, err := a.walk(handlers...)
	if err != nil {
		return fmt.Errorf("walk chains fail: %w", err)
	}
	if a.CheckMode(ModeDebug) {
		for _, chain := range chains {
			a.logger.CtxDebug(a.ctx, chain.Output())
		}
	}

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
	prog, _ := packages(initial, ssa.GlobalDebug|ssa.BareInits /*|ssa.PrintFunctions*/)
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
