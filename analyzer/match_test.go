package analyzer

import "testing"

func TestMatchHandler(t *testing.T) {
	var testdatas = []struct {
		Param  string
		Result string

		expected bool
	}{
		{"(c context.Context, ctx *app.RequestContext)", "()", true},
		{"(c context.Context, ctx *github.com/cloudwego/hertz/pkg/app.RequestContext)", "()", true},
	}

	for _, item := range testdatas {
		if MatchHandler(item.Param, item.Result) != item.expected {
			t.Errorf("match handler %s %s fail: expect: %t, got: %t", item.Param, item.Result, item.expected, !item.expected)
		}
	}
}
