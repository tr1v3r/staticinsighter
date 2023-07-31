package analyzer

import (
	"golang.org/x/tools/go/ssa"
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
