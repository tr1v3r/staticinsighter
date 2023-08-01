package analyzer

import (
	"golang.org/x/tools/go/ssa"
)

func (a *Analyzer) walk(enstries ...*ssa.Function) ([]*Chain, error) {
	for _, fn := range enstries {
		if a.CheckMode(ModeDebug) {
			a.printSSAFunc(fn)
		}
	}
	return nil, nil
}
