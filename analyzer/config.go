package analyzer

import "github.com/riverchu/pkg/log"

// Mode analyzer work mode
type Mode uint

const (
	// ModeDebug debug mode
	ModeDebug Mode = 1 << iota

	// ModeTraceLog print trace log
	ModeTraceLog

	// ModeUltimate ultimate mode
	ModeUltimate
)

func defaultConfigure() *Configure {
	return &Configure{
		logger: log.NewLogger(),
	}
}

type Configure struct {
	logger log.Logger

	Mode Mode
}

// WithLogger set analyzer logger
func (c *Configure) WithLogger(logger log.Logger) *Configure {
	c.logger = logger
	return c
}

// SetMode set analyzer run mode
func (c *Configure) SetMode(mode Mode) {
	c.Mode |= mode

	if c.CheckMode(ModeDebug) {
		c.logger.SetLevel(log.DebugLevel)
	}
	if c.CheckMode(ModeTraceLog) {
		c.logger.SetLevel(log.TraceLevel)
	}
}

// CheckMode check analyzer run mode
func (c *Configure) CheckMode(mode Mode) bool {
	return c.Mode&mode != 0
}
