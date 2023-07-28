package analyzer

import (
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func NewFuncitons() *Functions {
	return &Functions{
		initFuncs:    make(map[*ssa.Function]bool),
		mainFuncs:    make(map[*ssa.Function]bool),
		handlerFuncs: make(map[*ssa.Function]bool),
		sourceFuncs:  make(map[*ssa.Function]bool),
		sinkFuncs:    make(map[*ssa.Function]bool),
	}
}

type Functions struct {
	initFuncs    map[*ssa.Function]bool
	mainFuncs    map[*ssa.Function]bool
	handlerFuncs map[*ssa.Function]bool
	sourceFuncs  map[*ssa.Function]bool
	sinkFuncs    map[*ssa.Function]bool
}

// Match match and save all functions
func (f *Functions) Match(prog *ssa.Program) *Functions {
	for fn := range ssautil.AllFunctions(prog) {
		switch {
		case matcher.MatchMain(fn.Name()) && fn.Blocks != nil:
			f.mainFuncs[fn] = true
		case matcher.MatchInit(fn.Name()) && fn.Blocks != nil:
			f.initFuncs[fn] = true
		case matcher.MatchHandler(fn.Signature.Params().String(), fn.Signature.Results().String()):
			f.handlerFuncs[fn] = true
		case matcher.MatchSource(fn.Signature.Params().String(), fn.Signature.Results().String()):
			f.sourceFuncs[fn] = true
		case matcher.MatchSink(fn.Signature.Params().String(), fn.Signature.Results().String()):
			f.sinkFuncs[fn] = true
		}
	}
	return f
}

func (f *Functions) hasInit(fn *ssa.Function) bool    { return f.initFuncs[fn] }
func (f *Functions) hasMain(fn *ssa.Function) bool    { return f.mainFuncs[fn] }
func (f *Functions) hasHandler(fn *ssa.Function) bool { return f.handlerFuncs[fn] }
func (f *Functions) hasSource(fn *ssa.Function) bool  { return f.sourceFuncs[fn] }
func (f *Functions) hasSink(fn *ssa.Function) bool    { return f.sinkFuncs[fn] }
