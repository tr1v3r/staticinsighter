package analyzer

import (
	"fmt"
	"regexp"
)

var (
	// (c context.Context, ctx *app.RequestContext)
	handlerSigRules = []SigRule{
		{"hertz", "", `^\([\w\s]*context\.Context,[\w\s]*\*[\w./]*app.RequestContext\)$`, `^\(\)$`},
	}
	sourceSigRules = []SigRule{}
	sinkSigRules   = []SigRule{}
)

type SigRule struct {
	Frame  string `json:"frame"`
	Recv   string `json:"recv"`
	Param  string `json:"param"`
	Result string `json:"result"`
}

var matcher = new(Matcher)

func init() {
	if err := matcher.LoadRules(handlerSigRules, sourceSigRules, sinkSigRules); err != nil {
		panic(fmt.Errorf("matcher load rules fail: %w", err))
	}
}

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
		if err := sig.Load(rule.Frame, rule.Recv, rule.Param, rule.Result); err != nil {
			return nil, err
		}
		sigs = append(sigs, sig)
	}
	return
}

func (m *Matcher) MatchMain(fName string) bool {
	return fName == "main"
}

func (m *Matcher) MatchInit(fName string) bool {
	return fName == "init"
}

// MatchHandler match handler
func (m *Matcher) MatchHandler(param, result string) bool {
	for _, reg := range m.handlerSigs {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
}

// MatchSource match source
func (m *Matcher) MatchSource(param, result string) bool {
	for _, reg := range m.sourceSigs {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
}

// MatchSink match sink
func (m *Matcher) MatchSink(param, result string) bool {
	for _, reg := range m.sinkSigs {
		if reg.Match(param, result) {
			return true
		}
	}
	return false
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
