package analyzer

func defaultConfigure() *Configure {
	return &Configure{
		logger: defaultLogger(),
	}
}

type Configure struct {
	logger Logger

	Mode Mode
}

// WithLogger set analyzer logger
func (c *Configure) WithLogger(logger Logger) *Configure {
	c.logger = logger
	return c
}

func (c *Configure) CheckMode(mode Mode) bool {
	return c.Mode&mode != 0
}

// Mode analyzer work mode
type Mode uint

const (
	// ModeDebug debug mode
	ModeDebug Mode = 1 << iota

	// ModeUltimate ultimate mode
	ModeUltimate
)
