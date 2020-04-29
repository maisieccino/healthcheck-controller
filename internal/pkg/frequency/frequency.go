package frequency

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	second   = "s"
	minute   = "m"
	hour     = "h"
	day      = "d"
	week     = "w"
	dayUnit  = time.Hour * 24
	weekUnit = dayUnit * 7
)

var (
	frequencyToken         = `(\d+(\.\d+)?[smhdw])`
	frequencyTokenPattern  = regexp.MustCompile(frequencyToken)
	frequencyStringPattern = regexp.MustCompile(frequencyToken + "+$")
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
	frequencyTokens := frequencyTokenPattern.FindAllString(lowered, -1)

	components := make([]frequencyComponent, 0)
	highestUnit := time.Hour * 6000
	for _, token := range frequencyTokens {
		length := len(token)
		unitStr := string(token[length-1])
		amount, err := strconv.ParseFloat(token[0:length-1], 32)
		if err != nil {
			return EmptyFrequency(), errParsingToken(token)
		}
		var unit time.Duration
		switch unitStr {
		case week:
			unit = weekUnit
			break
		case day:
			unit = dayUnit
			break
		case hour:
			unit = time.Hour
			break
		case minute:
			unit = time.Minute
			break
		case second:
			unit = time.Second
			break
		}
		if unit > highestUnit {
			return EmptyFrequency(), errWrongOrder(expr)
		}
		highestUnit = unit
		newComponent := frequencyComponent{
			Unit:   unit,
			Amount: float32(amount),
		}
		components = append(components, newComponent)
	}
	return Frequency{
		components: components,
	}, nil
}

// frequencyComponent is a part of a frequency object, comprised of a unit and
// an amount.
type frequencyComponent struct {
	Unit   time.Duration
	Amount float32
}
