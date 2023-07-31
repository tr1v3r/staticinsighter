package analyzer

import (
	"golang.org/x/tools/go/ssa"
)

func (a *Analyzer) uniq(funcs []*ssa.Function) (result []*ssa.Function) {
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

func (a *Analyzer) walkChains(funcs ...*ssa.Function) ([]*Chain, error) {
	return nil, nil
}
