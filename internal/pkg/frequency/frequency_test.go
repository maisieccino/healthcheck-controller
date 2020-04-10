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
	t            *testing.T
	name         string
	input        string
	expectedFreq *Frequency
	expectedErr  error
}

func (tc testCase) test() {
	freq, err := ParseFrequency(tc.input)
	if err != tc.expectedErr {
		if tc.expectedErr != nil {
			tc.t.Errorf("Test case %s failed, expected error %s\n", tc.name, tc.expectedErr.Error())
		} else {
			tc.t.Errorf("Test case %s failed, unexpected error %v\n", tc.name, err)
		}
		return
	}
	if err != nil {
		return
	}
	for i, comp := range freq.components {
		if len(tc.expectedFreq.components) < i+1 {
			tc.t.Errorf("Test case %s failed, %d unexpected components: %+v", tc.name, len(freq.components)-len(tc.expectedFreq.components), freq.components[i:])
			return
		}
		expectedComp := tc.expectedFreq.components[i]
		if err := checkComponent(tc.t, expectedComp, comp); err != nil {
			tc.t.Errorf("Test case %s failed: %v\n", tc.name, err)
			return
		}
	}
	diff := len(tc.expectedFreq.components) - len(freq.components)
	if diff > 0 {
		tc.t.Errorf("Test case %s failed, %d expected additional components: %+v", tc.name, diff, tc.expectedFreq.components[len(freq.components):])
	}
}

var tt = []testCase{
	{
		name:         "twominutes_correct_format",
		input:        "2m",
		expectedErr:  nil,
		expectedFreq: &Frequency{components: []frequencyComponent{twoMins}},
	},
	{
		name:         "sixhours_correct_format",
		input:        "6h",
		expectedErr:  nil,
		expectedFreq: &Frequency{components: []frequencyComponent{sixHours}},
	},
	{
		name:         "sixhours_incorrect_format",
		input:        "6hours",
		expectedErr:  errInvalidExpr("6hours"),
		expectedFreq: nil,
	},
	{
		name:         "sixhourstwominutes_correct_format",
		input:        "6h2m",
		expectedErr:  nil,
		expectedFreq: &Frequency{components: []frequencyComponent{sixHours, twoMins}},
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
		name:         "fivehalfdays_correct_format",
		input:        "5.5d",
		expectedFreq: &Frequency{components: []frequencyComponent{fivehalfDays}},
		expectedErr:  nil,
	},
}

func TestFrequency(t *testing.T) {
	for _, tc := range tt {
		tc.t = t
		tc.test()
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
