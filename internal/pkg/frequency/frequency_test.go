package frequency

import (
	"fmt"
	"testing"
	"time"
)

var (
	twoMins      = frequencyComponent{Amount: 2, Unit: time.Minute}
	sixHours     = frequencyComponent{Amount: 6, Unit: time.Hour}
	fivehalfDays = frequencyComponent{Amount: 5.5, Unit: time.Hour * 24}
)

type testCase struct {
	t                *testing.T
	name             string
	input            string
	expectedFreq     *Frequency
	expectedErr      error
	expectedDuration time.Duration
}

func (tc testCase) test() {
	freq, err := ParseFrequency(tc.input)
	if err != nil {
		if tc.expectedErr != nil {
			if tc.expectedErr == err {
				return
			}
			tc.t.Errorf("expected error %s, actual error %s\n", tc.expectedErr.Error(), err.Error())
			return
		}
		tc.t.Errorf("unexpected error %v\n", err)
		return
	}
	for i, comp := range freq.components {
		if len(tc.expectedFreq.components) < i+1 {
			tc.t.Errorf("%d unexpected components: %+v", len(freq.components)-len(tc.expectedFreq.components), freq.components[i:])
			return
		}
		expectedComp := tc.expectedFreq.components[i]
		if err := checkComponent(tc.t, expectedComp, comp); err != nil {
			tc.t.Errorf("%v\n", err)
			return
		}
	}
	diff := len(tc.expectedFreq.components) - len(freq.components)
	if diff > 0 {
		tc.t.Errorf("%d expected additional components: %+v", diff, tc.expectedFreq.components[len(freq.components):])
	}
}

var tt = []testCase{
	{
		name:             "twominutes_correct_format",
		input:            "2m",
		expectedErr:      nil,
		expectedFreq:     &Frequency{components: []frequencyComponent{twoMins}},
		expectedDuration: 2 * time.Minute,
	},
	{
		name:             "sixhours_correct_format",
		input:            "6h",
		expectedErr:      nil,
		expectedFreq:     &Frequency{components: []frequencyComponent{sixHours}},
		expectedDuration: 6 * time.Hour,
	},
	{
		name:         "sixhours_incorrect_format",
		input:        "6hours",
		expectedErr:  errInvalidExpr("6hours"),
		expectedFreq: nil,
	},
	{
		name:             "sixhourstwominutes_correct_format",
		input:            "6h2m",
		expectedErr:      nil,
		expectedFreq:     &Frequency{components: []frequencyComponent{sixHours, twoMins}},
		expectedDuration: (6 * time.Hour) + (2 * time.Minute),
	},
	{
		name:         "sixhourstwominutes_incorrect_format",
		input:        "6h2minutes",
		expectedErr:  errInvalidExpr("6h2minutes"),
		expectedFreq: nil,
	},
	{
		name:         "sixhourstwominutes_bad_order",
		input:        "2m6h",
		expectedErr:  errWrongOrder("2m6h"),
		expectedFreq: &Frequency{components: []frequencyComponent{sixHours, twoMins}},
	},
	{
		name:             "fivehalfdays_correct_format",
		input:            "5.5d",
		expectedFreq:     &Frequency{components: []frequencyComponent{fivehalfDays}},
		expectedErr:      nil,
		expectedDuration: 5.5 * 24 * time.Hour,
	},
}

func TestFrequency(t *testing.T) {
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.t = t
			tc.test()
		})
	}
}

func checkComponent(t *testing.T, expected, actual frequencyComponent) error {
	if expected.Unit != actual.Unit {
		return fmt.Errorf("wrong unit, expected %s but got %s", expected.Unit, actual.Unit)
	}
	if expected.Amount != actual.Amount {
		return fmt.Errorf("wrong amount, expected %f but got %f", expected.Amount, actual.Amount)
	}
	return nil
}
