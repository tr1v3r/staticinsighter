package analyzer

import (
	"bytes"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (a *Analyzer) printAllUsefulFunc(prog *ssa.Program) {
	var (
		inits    = make(map[*ssa.Function]bool)
		mains    = make(map[*ssa.Function]bool)
		handlers = make(map[*ssa.Function]bool)
		sources  = make(map[*ssa.Function]bool)
		sinks    = make(map[*ssa.Function]bool)
	)
	for fn := range ssautil.AllFunctions(prog) {
		// if fn.Pkg == nil {
		// 	continue
		// }
		switch {
		case a.MatchInit(fn):
			inits[fn] = true
		case a.MatchMain(fn):
			mains[fn] = true
		case a.MatchHandler(fn):
			handlers[fn] = true
		case a.MatchSource(fn):
			sources[fn] = true
		case a.MatchSink(fn):
			sinks[fn] = true
		}
	}

	for fn := range mains {
		if a.mainFuncs[fn].Active() {
			a.logger.CtxDebug(a.ctx, "find active entry: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		} else {
			a.logger.CtxDebug(a.ctx, "find unactive entry: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		}
	}
	for fn := range inits {
		if a.initFuncs[fn].Active() {
			a.logger.CtxDebug(a.ctx, "find active entry: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		} else {
			a.logger.CtxTrace(a.ctx, "find unactive entry: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		}
	}
	for fn := range handlers {
		if a.handlerFuncs[fn].Risky() {
			a.logger.CtxDebug(a.ctx, "find risky handler: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		} else if a.handlerFuncs[fn].Active() {
			a.logger.CtxDebug(a.ctx, "find active handler: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		} else {
			a.logger.CtxDebug(a.ctx, "find unactive handler: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		}
	}
	for fn := range sources {
		a.logger.CtxDebug(a.ctx, "find source: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
	}
	for fn := range sinks {
		a.logger.CtxDebug(a.ctx, "find sink: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
	}
}

func (a *Analyzer) printSSAFunc(fn *ssa.Function) {
	buf := bytes.NewBuffer(nil)
	_, _ = fn.WriteTo(buf)
	a.logger.CtxDebug(a.ctx, "ssa func: %s", buf.String())
}
