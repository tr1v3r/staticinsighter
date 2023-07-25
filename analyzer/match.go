package analyzer

import (
	"fmt"
	"regexp"
)

var (
	handlerSigReg []*FuncSigInfo
	sourceSigReg  []*FuncSigInfo
	sinkSigReg    []*FuncSigInfo

	// (c context.Context, ctx *app.RequestContext)
	handlerSigRules = map[string][][3]string{
		"hertz": {{"", `^\([\w\s]*context\.Context,[\w\s]*\*[\w./]*app.RequestContext\)$`, `^\(\)$`}},
	}
	sourceSigRules = map[string][][3]string{}
	sinkSigRules   = map[string][][3]string{}
)

func init() {
	loadRule := func(sigReg []*FuncSigInfo, rules map[string][][3]string) {
		for frame, regs := range rules {
			for _, reg := range regs {
				sig := new(FuncSigInfo)
				if err := sig.Load(frame, reg[0], reg[1], reg[2]); err != nil {
					panic(fmt.Errorf("load rules fail: %w", err))
				}
				handlerSigReg = append(handlerSigReg, sig)
			}
		}
	}
	loadRule(handlerSigReg, handlerSigRules)
	loadRule(sourceSigReg, sourceSigRules)
	loadRule(sinkSigReg, sinkSigRules)
}

// FuncSigInfo ...
type FuncSigInfo struct {
	// Frame framework name
	Frame string

	Recv   *regexp.Regexp
	Param  *regexp.Regexp
	Result *regexp.Regexp
}

func (i *FuncSigInfo) Load(frame, recvReg, paramReg, resultReg string) (err error) {
	i.Frame = frame
	i.Recv, err = regexp.Compile(recvReg)
	if err != nil {
		return fmt.Errorf("compile recv reg fail: %w", err)
	}
	i.Param, err = regexp.Compile(paramReg)
	if err != nil {
		return fmt.Errorf("compile param reg fail: %w", err)
	}
	i.Result, err = regexp.Compile(resultReg)
	if err != nil {
		return fmt.Errorf("compile result reg fail: %w", err)
	}
	return nil
}

func (i *FuncSigInfo) Match(param, result string) bool {
	return i.Param.Match([]byte(param)) && i.Result.Match([]byte(result))
}

// MatchHandler match handler
func MatchHandler(param, result string) bool {
	for _, reg := range handlerSigReg {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
}

// MatchSource match source
func MatchSource(param, result string) bool {
	for _, reg := range sourceSigReg {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
}

// MatchSink match sink
func MatchSink(param, result string) bool {
	for _, reg := range sinkSigReg {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
}
