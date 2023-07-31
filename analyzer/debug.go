package analyzer

import (
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (a *Analyzer) PrintFuncs(funcs *Functions) {
	var printFuncInfo = func(typ string, funcs map[*ssa.Function]bool) {
		for fn := range funcs {
			a.logger.Debug("match %s: (%s).%s%s %s",
				typ, fn.Pkg.Pkg.Path(), fn.Name(),
				fn.Signature.Params().String(), fn.Signature.Results().String())
		}
	}

	printFuncInfo("init", funcs.initFuncs)
	printFuncInfo("main", funcs.mainFuncs)
	printFuncInfo("handler", funcs.handlerFuncs)
	printFuncInfo("source", funcs.sourceFuncs)
	printFuncInfo("sink", funcs.sinkFuncs)
}

func (a *Analyzer) printAllHandlers(prog *ssa.Program, handlers []*ssa.Function) {
	allHandlers := make(map[*ssa.Function]bool)
	for fn := range ssautil.AllFunctions(prog) {
		// if fn.Pkg == nil {
		// 	continue
		// }
		if a.MatchHandler(fn) {
			allHandlers[fn] = true
		}
	}

	for _, handler := range handlers {
		if !allHandlers[handler] {
			a.logger.CtxDebug(a.ctx, "active handler %s not found in all funcs", handler.Name())
			continue
		}

		a.logger.CtxDebug(a.ctx, "risky handler %s", handler.Name())
		delete(allHandlers, handler) // mark
	}
	for handler := range allHandlers {
		a.logger.CtxDebug(a.ctx, "unactive handler %s", handler.Name())
	}
}
