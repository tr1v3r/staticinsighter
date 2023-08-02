package analyzer

import (
	"sync"

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

		riskMap: make(map[*ssa.Function]*RiskInfo),
	}
}

type Functions struct {
	initFuncs map[*ssa.Function]*funcStat
	mainFuncs map[*ssa.Function]*funcStat

	handlerFuncs map[*ssa.Function]*funcStat

	sourceFuncs map[*ssa.Function]*funcStat
	sinkFuncs   map[*ssa.Function]*funcStat

	mu      sync.RWMutex
	riskMap map[*ssa.Function]*RiskInfo
}

func (f *Functions) GetRisk(fn *ssa.Function) *RiskInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.riskMap[fn]
}
func (f *Functions) AddRisk(r *RiskInfo) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.riskMap[r.Func()] = r
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

var deadCh = func() chan struct{} {
	var ch = make(chan struct{})
	close(ch)
	return ch
}()

// NewRisk create new risk info for function node
func NewRisk(fn *ssa.Function) *RiskInfo { return &RiskInfo{fn: fn, done: make(chan struct{})} }

// NewRiskSource create source risk info, always done
func NewRiskSource(fn *ssa.Function) *RiskInfo {
	return &RiskInfo{fn: fn, isSource: true, done: deadCh}
}

// NewRiskSink create sink risk info, always done
func NewRiskSink(fn *ssa.Function) *RiskInfo {
	return &RiskInfo{fn: fn, isSink: true, done: deadCh}
}

type RiskInfo struct {
	done chan struct{} // mark analyze finish status

	fn *ssa.Function

	isSink   bool
	isSource bool

	mu      sync.RWMutex
	sinks   []*RiskInfo
	sources []*RiskInfo
}

func (r *RiskInfo) Func() *ssa.Function { return r.fn }
func (r *RiskInfo) Risky() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return (len(r.sinks) > 0 && len(r.sources) > 0)
}
func (r *RiskInfo) IsSource() bool { return r.isSource }
func (r *RiskInfo) HasSource() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sources) > 0
}
func (r *RiskInfo) AddSoucre(risk *RiskInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources = append(r.sources, risk)
}
func (r *RiskInfo) IsSink() bool { return r.isSink }
func (r *RiskInfo) HasSink() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sinks) > 0
}
func (r *RiskInfo) AddSink(risk *RiskInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sinks = append(r.sinks, risk)
}

func (r *RiskInfo) Done() bool {
	select {
	case <-r.done:
		return true
	default:
		return false
	}
}
func (r *RiskInfo) AsyncDone() <-chan struct{} { return r.done }
func (r *RiskInfo) Finish()                    { close(r.done) }
