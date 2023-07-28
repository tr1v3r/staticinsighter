package analyzer

import (
	"golang.org/x/tools/go/ssa"
)

func (a *Analyzer) findHandler(funcs *Functions, fn *ssa.Function) (handlers []*ssa.Function) {
	for _, b := range fn.Blocks {
		for _, instr := range b.Instrs {
			switch i := instr.(type) {
			case *ssa.Call:
				if callee := i.Call.StaticCallee(); callee != nil {
					a.logger.Debug("call %s -> %s", fn.Name(), callee.Name())

					if funcs.hasHandler(callee) {
						handlers = append(handlers, callee)
						continue
					}

					for _, arg := range i.Call.Args {
						handlers = append(handlers, a.matchFunc(arg, funcs.hasHandler)...)
					}
					handlers = append(handlers, a.findHandler(funcs, callee)...)
				}
			default:
				handlers = append(handlers, a.matchFunc(i, funcs.hasHandler)...)
			}
		}
	}

	return a.uniq(handlers)
}

func (a *Analyzer) matchFunc(i any, match func(*ssa.Function) bool) (funcs []*ssa.Function) {
	switch i := i.(type) {
	case *ssa.Store:
		switch x := i.Val.(type) {
		case *ssa.ChangeType:
			funcs = append(funcs, a.matchFunc(x, match)...)
		}
	case *ssa.ChangeType:
		if x, ok := i.X.(*ssa.Function); ok {
			if match(x) {
				funcs = append(funcs, x)
			}
		}
	}
	return
}

func (a *Analyzer) uniq(funcs []*ssa.Function) (result []*ssa.Function) {
	m := make(map[*ssa.Function]bool, len(funcs))
	for _, fn := range funcs {
		m[fn] = true
	}
	for fn := range m {
		result = append(result, fn)
	}
	return result
}
