package analyzer

import (
	"testing"
)

// MatchHandler match handler
func (m *Matcher) UnderlyingMatch(recv, name string, params, results []string) bool {
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
		Params  []string
		Results []string

		expected bool
	}{
		// handler
		{"", "", []string{"context.Context", "*app.RequestContext"}, nil, true},
		{"", "", []string{"context.Context", "*github.com/cloudwego/hertz/pkg/app.RequestContext"}, nil, true},
		{"", "", []string{"*hertz.RequestContext"}, nil, true},

		// source
		{"*github.com/cloudwego/hertz/pkg/app.RequestContext", "GetString", []string{"string"}, []string{"string"}, true},
		{"*github.com/cloudwego/hertz/pkg/app.RequestContext", "BindAndValidate", []string{"interface{}"}, []string{"error"}, true},

		// sink
		{"", `SQLInject`, []string{"string"}, []string{"string"}, true},
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
