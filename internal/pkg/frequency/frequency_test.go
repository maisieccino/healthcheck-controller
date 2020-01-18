package frequency

import (
	"fmt"
	"testing"
	"time"
)

type testCase struct {
	t            *testing.T
	input        string
	expectedFreq Frequency
	expectedErr  error
}

func (tc testCase) test(id int) {
	freq, err := ParseFrequency(tc.input)
	if err != nil {
		tc.t.Errorf("Test case %d failed, error: %v\n", id, err)
	}
	for i, comp := range freq.components {
		if len(tc.expectedFreq.components) < i+1 {
			tc.t.Errorf("Test case %d failed, %d unexpected components: %+v", id, len(freq.components)-len(tc.expectedFreq.components), freq.components[i:])
			break
		}
		expectedComp := tc.expectedFreq.components[i]
		if err := checkComponent(tc.t, expectedComp, comp); err != nil {
			tc.t.Errorf("Test case %d failed: %v\n", id, err)
		}
	}
	if len(tc.expectedFreq.components) > len(freq.components) {
		tc.t.Errorf("Test case %d failed, %d expected additional components: %+v", id, len(tc.expectedFreq.components)-len(freq.components), tc.expectedFreq.components[len(freq.components):])
	}
}

var tt = []testCase{
	testCase{input: "2m", expectedErr: nil, expectedFreq: Frequency{components: []frequencyComponent{{Unit: time.Minute, Amount: 2}}}},
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
