package analyzer

import (
	"golang.org/x/tools/go/ssa"
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
