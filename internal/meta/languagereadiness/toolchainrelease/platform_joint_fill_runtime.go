package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateJointFillRuntime(r jointSmokeRuntime, cases []jointSmokeCase, actual []json.RawMessage) error {
	return validateJointFillRuntimeFor(r, cases, actual, "Main", "budgetplan://activity/main")
}

func validateJointFillRuntimeFor(r jointSmokeRuntime, cases []jointSmokeCase, actual []json.RawMessage, key, id string) error {
	if r.Stage != "COMPLETE" || r.Failure != "" || r.SHA == "" || !jointSmokeInt(r.Total, len(cases)) ||
		!jointSmokeInt(r.Calls, 0) || !jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) ||
		len(r.Traces) != len(cases) || actual != nil && len(actual) != len(cases) {
		return fmt.Errorf("caller source-fill native identity or denominator differs")
	}
	passed := 0
	seen := make([]bool, len(cases))
	for _, trace := range r.Traces {
		i := trace.Index
		if i < 0 || i >= len(cases) || seen[i] || len(trace.Deliveries) != 1 {
			return fmt.Errorf("caller source-fill row identity differs")
		}
		seen[i] = true
		d, expected := trace.Deliveries[0], cases[i].Expected[key]
		want := expected
		if actual != nil {
			want = actual[i]
		}
		match := samePackageSourceValue(want, expected)
		if d.ID != id || !samePackageSourceValue(d.Input, cases[i].Inputs[key]) ||
			!samePackageSourceValue(d.Expected, expected) || !samePackageSourceValue(d.Actual, want) || !jointSmokeBool(d.Passed, match) {
			return fmt.Errorf("caller source-fill exact actual/expected value differs at row %d", i)
		}
		if match {
			passed++
		}
	}
	if !jointSmokeInt(r.Passed, passed) {
		return fmt.Errorf("caller source-fill native numerator differs")
	}
	return nil
}
