package analyzer

import (
	"testing"
)

// MatchHandler match handler
func (m *Matcher) UnderlyingMatch(recv, name, params, results string) bool {
	for _, reg := range append(m.handlerSigs, append(m.sourceSigs, m.sinkSigs...)...) {
		if reg.Match(recv, name, params, results) {
			return true
		}
	}
	return false
}

func TestMatchHandler(t *testing.T) {
	var testdatas = []struct {
		Recv    string
		Name    string
		Params  string
		Results string

		expected bool
	}{
		// handler
		{"", "", "(c context.Context, ctx *app.RequestContext)", "()", true},
		{"", "", "(c context.Context, ctx *github.com/cloudwego/hertz/pkg/app.RequestContext)", "()", true},
		{"", "", "(c *hertz.RequestContext)", "()", true},

		// source
		{"*github.com/cloudwego/hertz/pkg/app.RequestContext", "GetString", "(key string)", "(s string)", true},
		{"*github.com/cloudwego/hertz/pkg/app.RequestContext", "BindAndValidate", "(obj interface{})", "(error)", true},

		// sink
		{"", `SQLInject`, "(s string)", "(error)", true},
	}

	matcher := NewMatcher()
	_ = matcher.LoadRules(handlerSigRules, sourceSigRules, sinkSigRules)
	for _, item := range testdatas {
		if matcher.UnderlyingMatch(item.Recv, item.Name, item.Params, item.Results) != item.expected {
			t.Errorf("match handler (%s)%s%s %s fail: expect: %t, got: %t",
				item.Recv, item.Name, item.Params, item.Results, item.expected, !item.expected)
		}
	}
}
