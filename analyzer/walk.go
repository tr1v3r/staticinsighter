package analyzer

import (
	"time"

	"golang.org/x/tools/go/ssa"
)

func (a *Analyzer) walk(handlers ...*ssa.Function) ([]*Chain, error) {
	a.logger.CtxInfo(a.ctx, "walking risky handlers (%d)...", len(handlers))
	defer func(s time.Time) { a.logger.CtxInfo(a.ctx, "walk risky handlers cost: %s", time.Since(s)) }(time.Now())

	for _, fn := range handlers {
		a.logger.CtxInfo(a.ctx, "walking risky handler: (%s).%s", fn.Pkg.Pkg.Path(), fn.Name())
		if a.CheckMode(ModeDebug) {
			// a.printSSAFunc(fn)
			_ = fn
		}
	}
	return nil, nil
}
