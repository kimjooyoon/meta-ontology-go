package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func validateJointSearchSmoke(raw, feedback, evaluation []byte, replay bool, selected string) (string, error) {
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
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v3" ||
		c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != "COMPLETE_FINITE" || c.StopReason != "LOCAL_AND_CALLER_CASES_MATCHED" ||
		!jointSmokeInt(c.Budget, 5) || c.Space != "5" || !jointSmokeInt(c.SelectedAttempt, 2) || len(c.Attempts) != 3 ||
		!slices.Equal(c.Kinds, []string{"source_search_index"}) || c.Source == "" || c.Selected.SHA == "" ||
		c.Initial.Model.Loaded != nil && *c.Initial.Model.Loaded || !jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) ||
		!jointSmokeInt(e.Separation.Consumed, 0) || !jointSmokeInt(e.Separation.Other, 4) {
		return "", fmt.Errorf("caller IR search identity, selection or input separation differs")
	}
	if replay && (selected == "" || selected != c.Selected.SHA) {
		return "", fmt.Errorf("caller IR search replay changed the selected program")
	}
	for i, attempt := range c.Attempts {
		if err := validateJointSearchLocal(attempt, i); err != nil {
			return "", err
		}
		if i != 1 {
			if err := validateJointSearchRuntime(attempt.Runtime, construction, i == 0); err != nil {
				return "", err
			}
		}
	}
	if c.Attempts[2].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA {
		return "", fmt.Errorf("caller IR search evaluation differs from the selected program")
	}
	if err := validateJointSearchRuntime(e.Runtime, cases, false); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}

func validateJointSearchLocal(a jointSmokeAttempt, index int) error {
	if len(a.Masks) != 1 || a.Masks[0] != index || len(a.Candidates) != 0 || len(a.SearchCandidates) != 1 {
		return fmt.Errorf("caller IR search candidate order differs")
	}
	c := a.SearchCandidates[0]
	x := c.Attempt
	if c.Schema != "gooo/search-candidate/v1" || c.Activity != "Choose" || c.InputSHA == "" || c.PlanSHA == "" || x.ID == "" ||
		x.Expression != []string{"input", "0", "-input"}[index] || !jointSmokeInt(x.Total, 1) {
		return fmt.Errorf("caller IR search expression or source identity differs")
	}
	if index == 1 {
		return validateJointSearchRejection(a, c)
	}
	if a.Rejection != nil || c.SelectedSHA == "" || !jointSmokeBool(x.Typed, true) || !jointSmokeBool(x.Scored, true) || x.Error != "" ||
		!jointSmokeInt(x.Passed, 1) || !jointSmokeInt(a.Passed, 1) || !jointSmokeInt(a.Total, 1) || len(x.Cases) != 1 || x.Accuracy == nil || *x.Accuracy != 100 {
		return fmt.Errorf("caller IR search local obligations differ")
	}
	row := x.Cases[0]
	if !samePackageSourceValue(row.Input, []byte("0")) || !samePackageSourceValue(row.Expected, []byte("0")) ||
		!samePackageSourceValue(row.Actual, []byte("0")) || !jointSmokeBool(row.Passed, true) {
		return fmt.Errorf("caller IR search local actual value differs")
	}
	return nil
}

func validateJointSearchRejection(a jointSmokeAttempt, c jointSmokeSearchCandidate) error {
	r, x, native := a.Rejection, c.Attempt, a.Runtime
	if r == nil || r.Stage != "LOCAL_SOURCE_SEARCH" || !jointSmokeInt(r.Slot, 0) || r.Activity != "Choose" || r.CandidateID != x.ID ||
		r.Reason != x.Error || !strings.Contains(x.Error, "division by zero") || c.SelectedSHA != "" ||
		!jointSmokeBool(x.Typed, false) || !jointSmokeBool(x.Scored, false) || !jointSmokeInt(x.Passed, 0) || x.Accuracy != nil || len(x.Cases) != 0 ||
		!jointSmokeInt(a.Passed, 0) || !jointSmokeInt(a.Total, 0) || native.Stage != "" || native.SHA != "" || native.Failure != "" ||
		!jointSmokeInt(native.Passed, 0) || !jointSmokeInt(native.Total, 0) || !jointSmokeInt(native.Calls, 0) ||
		!jointSmokeBool(native.Projection, false) || !jointSmokeBool(native.Replay, false) || len(native.Traces) != 0 {
		return fmt.Errorf("caller IR search rejection claims a score or loses its original error")
	}
	return nil
}

func validateJointSearchRuntime(r jointSmokeRuntime, cases []jointSmokeCase, baseline bool) error {
	passed := len(cases)
	if baseline {
		passed = 0
	}
	if r.Stage != "COMPLETE" || r.Failure != "" || r.SHA == "" || !jointSmokeInt(r.Passed, passed) || !jointSmokeInt(r.Total, len(cases)) ||
		!jointSmokeInt(r.Calls, 0) || !jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || len(r.Traces) != len(cases) {
		return fmt.Errorf("caller IR search native counts differ")
	}
	seen := make([]bool, len(cases))
	for _, trace := range r.Traces {
		i := trace.Index
		if i < 0 || i >= len(cases) || seen[i] || len(trace.Deliveries) != 1 {
			return fmt.Errorf("caller IR search row identity differs")
		}
		seen[i] = true
		d, expected := trace.Deliveries[0], cases[i].Expected["Main"]
		actual := expected
		if baseline {
			actual = []byte("1")
		}
		if d.ID != "candidateerror://activity/main" || !samePackageSourceValue(d.Input, cases[i].Inputs["Main"]) ||
			!samePackageSourceValue(d.Expected, expected) || !samePackageSourceValue(d.Actual, actual) || !jointSmokeBool(d.Passed, !baseline) {
			return fmt.Errorf("caller IR search native actual value differs")
		}
	}
	return nil
}
