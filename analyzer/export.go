package analyzer

// SetMode set analyze mode
func SetMode(mode Mode) {
	defaultAnalyzer.SetMode(mode)
}

// Analyze ...
func Analyze(paths ...string) error {
	return defaultAnalyzer.Analyze(paths...)
}
