package analyzer

import (
	"time"

	"golang.org/x/tools/go/ssa"
)

func (a *Analyzer) walk(handlers ...*ssa.Function) ([]*Chain, error) {
	a.logger.Info("walk risky handlers...")
	defer func(s time.Time) { a.logger.Info("walk risky handlers cost: %s", time.Since(s)) }(time.Now())

	for _, fn := range handlers {
		if a.CheckMode(ModeDebug) {
			a.printSSAFunc(fn)
		}
	}
	return nil, nil
}
