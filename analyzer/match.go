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
	recv, name, params, results := m.recv(sig), fn.Name(), sig.Params().String(), sig.Results().String()
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
	recv, name, params, results := m.recv(sig), fn.Name(), sig.Params().String(), sig.Results().String()
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
	recv, name, params, results := m.recv(sig), fn.Name(), sig.Params().String(), sig.Results().String()
	for _, reg := range m.sinkSigs {
		if reg.Match(recv, name, params, results) {
			return true
		}
	}
	return false
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
	Recv   *regexp.Regexp
	Name   *regexp.Regexp
	Param  *regexp.Regexp
	Result *regexp.Regexp
}

func (i *FuncSigInfo) Load(frame, nameReg, recvReg, paramReg, resultReg string) (err error) {
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
	if paramReg != "" {
		i.Param, err = regexp.Compile(paramReg)
		if err != nil {
			return fmt.Errorf("compile param reg fail: %w", err)
		}
	}
	if resultReg != "" {
		i.Result, err = regexp.Compile(resultReg)
		if err != nil {
			return fmt.Errorf("compile result reg fail: %w", err)
		}
	}
	return nil
}

func (i *FuncSigInfo) Match(recv, name, param, result string) bool {
	return (i.Recv == nil || i.Recv.Match([]byte(recv))) &&
		(i.Name == nil || i.Name.Match([]byte(name))) &&
		(i.Param == nil || i.Param.Match([]byte(param))) &&
		(i.Result == nil || i.Result.Match([]byte(result)))
}
