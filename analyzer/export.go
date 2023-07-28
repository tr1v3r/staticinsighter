package analyzer

import (
	"context"

	"github.com/riverchu/pkg/log"
	"golang.org/x/tools/go/ssa"
)

// NewAnalyzer build a new analyzer
func NewAnalyzer(ctx context.Context) *Analyzer {
	return &Analyzer{
		ctx:       ctx,
		Configure: defaultConfigure(),
		funcs:     make(map[*ssa.Function]bool),
	}
}

// SetLogLevel set log level
func SetLogLevel(level log.Level) {
	defaultAnalyzer.Configure.logger.SetLevel(level)
}

// Analyze ...
func Analyze(paths ...string) error {
	return defaultAnalyzer.Analyze(paths...)
}
