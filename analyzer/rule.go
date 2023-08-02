package analyzer

var (
	handlerSigRules = []SigRule{
		{"hertz", "", "", []string{`^\*[\w./]*hertz\.RequestContext$`}, nil},
		{"hertz", "", "", []string{`^context\.Context$`, `^\*[\w./]*app\.RequestContext$`}, nil},
	}

	sourceSigRules = []SigRule{
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^BindAndValidate$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetString$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetHeader$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetRawData$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^ClientIP$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^UserAgent$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^ContentType$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetString$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^Get$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^Param$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^QueryArgs$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^PostArgs$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^Query$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^DefaultQuery$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetQuery$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^PostForm$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^DefaultPostForm$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^GetPostForm$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^FormFile$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^MultipartForm$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^Path$`, nil, nil},
		{"hertz", `^\*[\w./]*app\.RequestContext$`, `^FullPath$`, nil, nil},
	}
	sinkSigRules = []SigRule{
		// {"diy", ``, `^SQLInject$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Raw$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Exec$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Group$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Order$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Joins$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Where$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Select$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Or$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Not$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Having$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Find$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Delete$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Take$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^First$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Last$`, nil, nil},
		{"gorm", `gorm\.DB$`, `^Preload$`, nil, nil},
	}
)

type SigRule struct {
	Frame  string   `json:"frame"`
	Recv   string   `json:"recv"`
	Name   string   `json:"name"`
	Param  []string `json:"param"`
	Result []string `json:"result"`
}
