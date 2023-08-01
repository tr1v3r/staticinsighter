package analyzer

import (
	"golang.org/x/tools/go/ssa"
)

type funcStat struct {
	active bool
	risky  bool
}

func (f *funcStat) Active() bool { return f != nil && f.active }
func (f *funcStat) Risky() bool  { return f != nil && f.risky }

func NewFuncitons() *Functions {
	return &Functions{
		initFuncs:    make(map[*ssa.Function]*funcStat),
		mainFuncs:    make(map[*ssa.Function]*funcStat),
		handlerFuncs: make(map[*ssa.Function]*funcStat),
		sourceFuncs:  make(map[*ssa.Function]*funcStat),
		sinkFuncs:    make(map[*ssa.Function]*funcStat),
	}
}

type Functions struct {
	initFuncs map[*ssa.Function]*funcStat
	mainFuncs map[*ssa.Function]*funcStat

	handlerFuncs map[*ssa.Function]*funcStat

	sourceFuncs map[*ssa.Function]*funcStat
	sinkFuncs   map[*ssa.Function]*funcStat
}

func (f *Functions) AddActiveInit(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if stat, ok := f.initFuncs[fn]; !ok {
			f.initFuncs[fn] = &funcStat{active: true}
		} else {
			stat.active = true
		}
	}
}

func (f *Functions) AddActiveMain(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if stat, ok := f.mainFuncs[fn]; !ok {
			f.mainFuncs[fn] = &funcStat{active: true}
		} else {
			stat.active = true
		}
	}
}

func (f *Functions) AddActiveHandler(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if stat, ok := f.handlerFuncs[fn]; !ok {
			f.handlerFuncs[fn] = &funcStat{active: true}
		} else {
			stat.active = true
		}
	}
}

func (f *Functions) AddRiskyHandler(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if stat, ok := f.handlerFuncs[fn]; !ok {
			f.handlerFuncs[fn] = &funcStat{active: true, risky: true}
		} else {
			stat.active = true
			stat.risky = true
		}
	}
}

func (f *Functions) AddSource(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if f.sourceFuncs[fn] != nil {
			f.sourceFuncs[fn] = new(funcStat)
		}
	}
}

func (f *Functions) AddSink(funcs ...*ssa.Function) {
	for _, fn := range funcs {
		if f.sinkFuncs[fn] != nil {
			f.sinkFuncs[fn] = new(funcStat)
		}
	}
}

func (*Functions) uniq(funcs []*ssa.Function) (result []*ssa.Function) {
	m := make(map[*ssa.Function]bool, len(funcs))
	for _, fn := range funcs {
		m[fn] = true
	}
	delete(m, nil)
	for fn := range m {
		result = append(result, fn)
	}
	return result
}
