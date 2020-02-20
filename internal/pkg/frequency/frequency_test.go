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
	input        string
	expectedFreq *Frequency
	expectedErr  error
}

func (tc testCase) test(id int) {
	freq, err := ParseFrequency(tc.input)
	if err != tc.expectedErr {
		if tc.expectedErr != nil {
			tc.t.Errorf("Test case %d failed, expected error %s\n", id, tc.expectedErr.Error())
		} else {
			tc.t.Errorf("Test case %d failed, unexpected error %v\n", id, err)
		}
		return
	}
	if err != nil {
		return
	}
	for i, comp := range freq.components {
		if len(tc.expectedFreq.components) < i+1 {
			tc.t.Errorf("Test case %d failed, %d unexpected components: %+v", id, len(freq.components)-len(tc.expectedFreq.components), freq.components[i:])
			return
		}
		expectedComp := tc.expectedFreq.components[i]
		if err := checkComponent(tc.t, expectedComp, comp); err != nil {
			tc.t.Errorf("Test case %d failed: %v\n", id, err)
			return
		}
	}
	diff := len(tc.expectedFreq.components) - len(freq.components)
	if diff > 0 {
		tc.t.Errorf("Test case %d failed, %d expected additional components: %+v", id, diff, tc.expectedFreq.components[len(freq.components):])
	}
}

var tt = []testCase{
	testCase{
		input:        "2m",
		expectedErr:  nil,
		expectedFreq: &Frequency{components: []frequencyComponent{twoMins}},
	},
	testCase{
		input:        "6h",
		expectedErr:  nil,
		expectedFreq: &Frequency{components: []frequencyComponent{sixHours}},
	},
	testCase{
		input:        "6hours",
		expectedErr:  errInvalidExpr("6hours"),
		expectedFreq: nil,
	},
}

func TestFrequency(t *testing.T) {
	for i, tc := range tt {
		tc.t = t
		tc.test(i)
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
