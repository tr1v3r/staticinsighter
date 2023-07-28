package analyzer

import "github.com/riverchu/pkg/log"

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
