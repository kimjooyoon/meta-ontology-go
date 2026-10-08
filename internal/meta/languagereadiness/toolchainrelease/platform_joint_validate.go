package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

type jointSmokeCase struct {
	Inputs, Expected map[string]json.RawMessage
}

func validateJointSmoke(raw, feedback, evaluation []byte, replay bool, selected string) (string, error) {
	var r jointSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	construction, err := jointSmokeCases(feedback, 1)
	if err != nil {
		return "", err
	}
	cases, err := jointSmokeCases(evaluation, 4)
	if err != nil {
		return "", err
	}
	c, e := r.Construction, r.Evaluation
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v1" ||
		c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != "COMPLETE_FINITE" ||
		c.StopReason != "LOCAL_AND_CALLER_CASES_MATCHED" || !jointSmokeInt(c.Budget, 2) ||
		c.Space != "2" || !jointSmokeInt(c.SelectedAttempt, 1) || len(c.Attempts) != 2 || c.Source == "" ||
		c.Selected.SHA == "" || !jointSmokeBool(c.Initial.Model.Loaded, false) ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) ||
		!jointSmokeInt(e.Separation.Consumed, 1) || !jointSmokeInt(e.Separation.Other, 3) {
		return "", fmt.Errorf("caller construction, replay or input separation differs")
	}
	if replay && (selected == "" || c.Selected.SHA != selected) {
		return "", fmt.Errorf("caller replay changed the selected program")
	}
	for i, attempt := range c.Attempts {
		if err := validateJointSmokeLocal(attempt, i); err != nil {
			return "", err
		}
		if err := validateJointSmokeRuntime(attempt.Runtime, construction, i == 0); err != nil {
			return "", err
		}
	}
	if c.Attempts[1].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA {
		return "", fmt.Errorf("caller evaluation does not use the selected program")
	}
	if err := validateJointSmokeRuntime(e.Runtime, cases, false); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}

func jointSmokeCases(raw []byte, count int) ([]jointSmokeCase, error) {
	var r struct {
		Schema string
		Cases  []jointSmokeCase
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	if r.Schema != "gooo/body-composition-cases/v1" || len(r.Cases) != count {
		return nil, fmt.Errorf("caller smoke reference cases differ")
	}
	for _, row := range r.Cases {
		if len(row.Inputs) != 1 || len(row.Expected) != 1 || len(row.Inputs["Main"]) == 0 || len(row.Expected["Main"]) == 0 {
			return nil, fmt.Errorf("caller smoke reference must name Main")
		}
	}
	return r.Cases, nil
}

func validateJointSmokeLocal(a jointSmokeAttempt, mask int) error {
	if len(a.Masks) != 1 || a.Masks[0] != mask || !jointSmokeInt(a.Passed, 1) || !jointSmokeInt(a.Total, 1) || len(a.Candidates) != 1 {
		return fmt.Errorf("caller construction lost its original local obligation")
	}
	c := a.Candidates[0]
	if c.Activity != "Choose" || !jointSmokeInt(c.Attempt.Mask, mask) ||
		!jointSmokeInt(c.Attempt.Passed, 1) || !jointSmokeInt(c.Attempt.Total, 1) || len(c.Cases) != 1 {
		return fmt.Errorf("caller construction local candidate differs")
	}
	value := c.Cases[0]
	if len(value.Inputs) != 1 || !samePackageSourceValue(value.Inputs[0], []byte("0")) ||
		!samePackageSourceValue(value.Actual, []byte(`{"value":0}`)) ||
		!samePackageSourceValue(value.Expected, []byte(`{"value":0}`)) || !jointSmokeBool(value.Passed, true) {
		return fmt.Errorf("caller construction changed the local actual value")
	}
	return nil
}

func validateJointSmokeRuntime(r jointSmokeRuntime, cases []jointSmokeCase, baseline bool) error {
	wantPassed := len(cases)
	if baseline {
		wantPassed = 0
	}
	if r.Stage != "COMPLETE" || r.Failure != "" || r.SHA == "" ||
		!jointSmokeInt(r.Passed, wantPassed) || !jointSmokeInt(r.Total, len(cases)) ||
		!jointSmokeInt(r.Calls, 0) || !jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || len(r.Traces) != len(cases) {
		return fmt.Errorf("caller native result counts or replay differ")
	}
	seen := make([]bool, len(cases))
	for _, trace := range r.Traces {
		i := trace.Index
		if i < 0 || i >= len(cases) || seen[i] || len(trace.Deliveries) != 1 {
			return fmt.Errorf("caller native input row differs")
		}
		seen[i] = true
		d, expected := trace.Deliveries[0], cases[i].Expected["Main"]
		actual := expected
		if baseline {
			actual = cases[i].Inputs["Main"]
		}
		if d.ID != "caller://activity/main" || !samePackageSourceValue(d.Input, cases[i].Inputs["Main"]) ||
			!samePackageSourceValue(d.Expected, expected) || !samePackageSourceValue(d.Actual, actual) ||
			!jointSmokeBool(d.Passed, !baseline) {
			return fmt.Errorf("caller native actual value differs at row %d", i)
		}
	}
	return nil
}

func jointSmokeInt(v *int, want int) bool    { return v != nil && *v == want }
func jointSmokeBool(v *bool, want bool) bool { return v != nil && *v == want }
