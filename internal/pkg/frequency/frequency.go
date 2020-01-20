package frequency

import (
	"strings"
	"time"
)

const (
	second = "s"
	minute = "m"
	hour   = "h"
	day    = "d"
	week   = "w"
)

// Frequency represents a frequency to reapeat things.
type Frequency struct {
	components []frequencyComponent
	cronExpr   string
}

// ToCronExpr returns a cron-formatted string that represents the given
// frequency.
func (f Frequency) ToCronExpr() string {
	return ""
}

// ToDuration returns the length of time the frequency represents.
func (f Frequency) ToDuration() time.Duration {
	return 0
}

// ParseFrequency will take a frequency expression string and parse it to a
// frequency object.
func ParseFrequency(expr string) (Frequency, error) {
	lowered := strings.ToLower(expr)
	return Frequency{
		components: []frequencyComponent{},
	}, nil
}

// frequencyComponent is a part of a frequency object, comprised of a unit and
// an amount.
type frequencyComponent struct {
	Unit   time.Duration
	Amount float32
}
