package analyzer

import "context"

// NewAnalyzer build a new analyzer
func NewAnalyzer(ctx context.Context) *Analyzer {
	return &Analyzer{ctx: ctx, Configure: defaultConfigure()}
}

// SetLogLevel set log level
func SetLogLevel(level Level) {
	defaultAnalyzer.Configure.logger.SetLevel(level)
}

// Analyze ...
func Analyze(path string) error {
	return defaultAnalyzer.Analyze(path)
}
