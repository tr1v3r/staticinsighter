package analyzer

var (
	handlerSigRules = []SigRule{
		{"hertz", "", "", `^\([\w\s]*context\.Context,[\w\s]*\*[\w./]*app.RequestContext\)$`, `^\(\)$`},
		{"hertz", "", "", `^\([\w\s]*\*[\w./]*hertz.RequestContext\)$`, `^\(\)$`},
	}
	sourceSigRules = []SigRule{
		{"hertz", `^\*[\w./]*app.RequestContext$`, `^GetString$`, `^\(key string\)$`, `^\(s string\)$`},
		{"hertz", `^\*[\w.\/]*app.RequestContext$`, `^BindAndValidate$`, `^\(obj interface{}\)$`, `^\(error\)$`},
	}
	sinkSigRules = []SigRule{
		{"diy", ``, `^SQLInject$`, ``, ``},
	}
)

type SigRule struct {
	Frame  string `json:"frame"`
	Recv   string `json:"recv"`
	Name   string `json:"name"`
	Param  string `json:"param"`
	Result string `json:"result"`
}
