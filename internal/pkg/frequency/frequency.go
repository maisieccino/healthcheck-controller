package frequency

import (
	"regexp"
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

var (
	frequencyStringPattern = regexp.MustCompile(`(\d+(.\d+)?[smhdw])+`)
)

type Frequency struct {
	components []frequencyComponent
	cronExpr   string
}

func EmptyFrequency() Frequency {
	return Frequency{
		components: []frequencyComponent{},
		cronExpr:   "",
	}
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
	if isValid := frequencyStringPattern.MatchString(lowered); !isValid {
		return EmptyFrequency(), errInvalidExpr(expr)
	}
	frequencyStringPattern.FindAllString(lowered, -1)
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
