package analyzer

import (
	"fmt"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/ssa"
)

func NewMatcher() *Matcher { return new(Matcher) }

// Matcher ...
type Matcher struct {
	handlerSigs []*FuncSigInfo
	sourceSigs  []*FuncSigInfo
	sinkSigs    []*FuncSigInfo
}

func (m *Matcher) LoadRules(handlers, sources, sinks []SigRule) error {
	sigs, err := m.loadRules(handlers)
	if err != nil {
		return fmt.Errorf("load handler rules fail: %w", err)
	}
	m.handlerSigs = append(m.handlerSigs, sigs...)

	sigs, err = m.loadRules(sources)
	if err != nil {
		return fmt.Errorf("load source rules fail: %w", err)
	}
	m.sourceSigs = append(m.handlerSigs, sigs...)

	sigs, err = m.loadRules(sinks)
	if err != nil {
		return fmt.Errorf("load sink rules fail: %w", err)
	}
	m.sinkSigs = append(m.handlerSigs, sigs...)
	return nil
}

func (m *Matcher) loadRules(rules []SigRule) (sigs []*FuncSigInfo, err error) {
	for _, rule := range rules {
		sig := new(FuncSigInfo)
		if err := sig.Load(rule.Frame, rule.Name, rule.Recv, rule.Param, rule.Result); err != nil {
			return nil, err
		}
		sigs = append(sigs, sig)
	}
	return
}

func (m *Matcher) MatchMain(fn *ssa.Function) bool { return fn.Name() == "main" }
func (m *Matcher) MatchInit(fn *ssa.Function) bool { return fn.Name() == "init" }

// MatchHandler match handler
func (m *Matcher) MatchHandler(fn *ssa.Function) bool {
	sig := fn.Signature
	recv, name, params, results := m.recv(sig), fn.Name(), m.getTypes(sig.Params()), m.getTypes(sig.Results())
	for _, reg := range m.handlerSigs {
		if reg.Match(recv, name, params, results) {
			return true
		}
	}
	return false
}

// MatchSource match source
func (m *Matcher) MatchSource(fn *ssa.Function) bool {
	sig := fn.Signature
	recv, name, params, results := m.recv(sig), fn.Name(), m.getTypes(sig.Params()), m.getTypes(sig.Results())
	for _, reg := range m.sourceSigs {
		if reg.Match(recv, name, params, results) {
			return true
		}
	}
	return false
}

// MatchSink match sink
func (m *Matcher) MatchSink(fn *ssa.Function) bool {
	sig := fn.Signature
	recv, name, params, results := m.recv(sig), fn.Name(), m.getTypes(sig.Params()), m.getTypes(sig.Results())
	for _, reg := range m.sinkSigs {
		if reg.Match(recv, name, params, results) {
			return true
		}
	}
	return false
}

func (m *Matcher) getTypes(params *types.Tuple) (types []string) {
	for i := 0; i < params.Len(); i++ {
		types = append(types, params.At(i).Type().String())
	}
	return
}

func (m *Matcher) recv(sig *types.Signature) string {
	if recv := sig.Recv(); recv != nil {
		return recv.Type().String()
	}
	return ""
}

// FuncSigInfo ...
type FuncSigInfo struct {
	// Frame framework name
	Frame string

	// reg rules
	Recv    *regexp.Regexp
	Name    *regexp.Regexp
	Params  []*regexp.Regexp
	Results []*regexp.Regexp
}

func (i *FuncSigInfo) Load(frame, nameReg, recvReg string, paramReg, resultReg []string) (err error) {
	i.Frame = frame
	if recvReg != "" {
		i.Recv, err = regexp.Compile(recvReg)
		if err != nil {
			return fmt.Errorf("compile recv reg fail: %w", err)
		}
	}
	if nameReg != "" {
		i.Name, err = regexp.Compile(nameReg)
		if err != nil {
			return fmt.Errorf("compile name reg fail: %w", err)
		}
	}
	if len(paramReg) > 0 {
		for _, rule := range paramReg {
			reg, err := regexp.Compile(rule)
			if err != nil {
				return fmt.Errorf("compile param reg fail: %w", err)
			}
			i.Params = append(i.Params, reg)
		}
	}
	if len(resultReg) > 0 {
		for _, rule := range resultReg {
			reg, err := regexp.Compile(rule)
			if err != nil {
				return fmt.Errorf("compile result reg fail: %w", err)
			}
			i.Results = append(i.Results, reg)
		}
	}
	return nil
}

func (i *FuncSigInfo) Match(recv, name string, params, results []string) bool {
	return (i.Recv == nil || i.Recv.Match([]byte(recv))) &&
		(i.Name == nil || i.Name.Match([]byte(name))) &&
		(i.Params == nil || i.iterateMatch(i.Params, params)) &&
		(i.Results == nil || i.iterateMatch(i.Results, results))
}

func (i *FuncSigInfo) iterateMatch(regs []*regexp.Regexp, datas []string) bool {
	if len(regs) != len(datas) {
		return false
	}
	for i, reg := range regs {
		if !reg.Match([]byte(datas[i])) {
			return false
		}
	}
	return true
}
