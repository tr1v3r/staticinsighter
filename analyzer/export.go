package analyzer

import (
	"context"
)

// NewAnalyzer build a new analyzer
func NewAnalyzer(ctx context.Context) *Analyzer {
	return &Analyzer{
		ctx:       ctx,
		Configure: defaultConfigure(),
	}
}

// SetMode set analyze mode
func SetMode(mode Mode) {
	defaultAnalyzer.SetMode(mode)
}

// Analyze ...
func Analyze(paths ...string) error {
	return defaultAnalyzer.Analyze(paths...)
}
