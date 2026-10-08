package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func validateJointFillSmoke(raw, source, feedback, evaluation []byte, budget int, replay bool, selected string) (string, error) {
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
	index, decision, stop := 2, "COMPLETE_FINITE", "LOCAL_AND_CALLER_CASES_MATCHED"
	if budget == 2 {
		index, decision, stop = 0, "PARTIAL_FINITE", "PROGRAM_BUDGET_EXHAUSTED"
	}
	if budget != 2 && budget != 3 || replay && budget != 3 || !jointSmokeBool(r.Generated, !replay) ||
		c.Schema != "gooo/joint-construction/v4" || c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != decision || c.StopReason != stop ||
		!jointSmokeInt(c.Budget, budget) || c.Space != "3" || !jointSmokeInt(c.SelectedAttempt, index) || len(c.Attempts) != budget ||
		!slices.Equal(c.Kinds, []string{"source_fill_index"}) || c.Source == "" || c.Selected.SHA == "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) ||
		strings.Contains(c.Source, "__GOOO_BODY_HOLE_") || strings.Contains(c.Source, "assembling") ||
		c.Initial.Model.Loaded != nil && *c.Initial.Model.Loaded || c.Initial.FillModel.Loaded != nil && *c.Initial.FillModel.Loaded ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) ||
		!jointSmokeInt(e.Separation.Consumed, 0) || !jointSmokeInt(e.Separation.Other, 4) {
		return "", fmt.Errorf("caller source-fill identity, selection or input separation differs")
	}
	if len(c.Initial.Preparations) != 1 {
		return "", fmt.Errorf("caller source-fill initial preparation differs")
	}
	initial := c.Initial.Preparations[0].Generation.Report.Fill
	if initial == nil || initial.Selected != "late_unbounded" || !jointSmokeInt(initial.Calls, 0) {
		return "", fmt.Errorf("initial local selection was changed")
	}
	if replay && (selected == "" || selected != c.Selected.SHA) {
		return "", fmt.Errorf("source-fill replay changed the selected program")
	}
	for i, a := range c.Attempts {
		if err := validateJointFillLocal(a, i, c.OriginalSHA); err != nil {
			return "", err
		}
		if a.FillCandidates[0].PlanSHA != c.Attempts[0].FillCandidates[0].PlanSHA {
			return "", fmt.Errorf("source-fill plan differs between attempts")
		}
		actual := []json.RawMessage{json.RawMessage([]string{"9", "-7", "-8"}[i])}
		if err := validateJointFillRuntime(a.Runtime, construction, actual); err != nil {
			return "", err
		}
	}
	if c.Attempts[index].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA ||
		c.Attempts[index].FillCandidates[0].SelectedSHA != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(c.Source))) {
		return "", fmt.Errorf("caller source-fill selected program differs from execution")
	}
	var actual []json.RawMessage
	if budget == 2 {
		actual = []json.RawMessage{[]byte("2"), []byte("-12"), []byte("5"), []byte("9007199254740994")}
	}
	if err := validateJointFillRuntime(e.Runtime, cases, actual); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}

func validateJointFillLocal(a jointSmokeAttempt, index int, sourceSHA string) error {
	return validateJointFillLocalAt(a, index, index, 3, sourceSHA)
}

func validateJointFillLocalAt(a jointSmokeAttempt, index, selector, count int, sourceSHA string) error {
	if a.Rejection != nil || len(a.Candidates) != 0 || len(a.SearchCandidates) != 0 || len(a.FillCandidates) != 1 ||
		len(a.Masks) != 1 || a.Masks[0] != selector || !jointSmokeInt(a.Passed, 1) || !jointSmokeInt(a.Total, 1) {
		return fmt.Errorf("caller source-fill candidate order or training denominator differs")
	}
	c := a.FillCandidates[0]
	condition, cap := "input.used > input.limit", "input.used"
	if index > 0 {
		condition, cap = "input.used >= input.limit", "input.limit - 1"
	}
	if index == 2 {
		cap = "input.limit"
	}
	if c.Schema != "gooo/fill-candidate/v1" || c.Activity != "PlanBudget" || c.ActivityID != "budgetplan://activity/plan-budget" ||
		c.Rejection != nil || c.InputSHA != sourceSHA || c.SelectedSHA == "" || c.PlanSHA == "" || !jointSmokeInt(c.Count, count) ||
		c.ID != []string{"late_unbounded", "early_wrong_cap", "bounded"}[index] || c.Method != "caller_selected_assignment" ||
		len(c.Holes) != 2 || c.Holes[0].ID != "boundary" || c.Holes[0].Expression != condition || c.Holes[1].ID != "cap" || c.Holes[1].Expression != cap ||
		!jointSmokeInt(c.Passed, 1) || !jointSmokeInt(c.Total, 1) || !jointSmokeInt(c.HoldoutPassed, index/2) || !jointSmokeInt(c.HoldoutTotal, 1) ||
		len(c.Cases) != 0 || len(c.Holdout) != 0 || len(c.Values) != 1 || len(c.ValueHoldout) != 1 {
		return fmt.Errorf("caller source-fill assignment, source or separate holdout counts differ")
	}
	if !jointFillCaseMatches(c.Values[0], `{"used":0,"limit":8}`, `{"next":1,"exhausted":false}`, `{"next":1,"exhausted":false}`, true) ||
		!jointFillCaseMatches(c.ValueHoldout[0], `{"used":9,"limit":8}`, `{"next":8,"exhausted":true}`,
			[]string{`{"next":9,"exhausted":true}`, `{"next":7,"exhausted":true}`, `{"next":8,"exhausted":true}`}[index], index == 2) {
		return fmt.Errorf("caller source-fill original local values differ")
	}
	return nil
}

func jointFillCaseMatches(row jointSmokeFillCase, input, expected, actual string, passed bool) bool {
	return len(row.Inputs) == 1 && samePackageSourceValue(row.Inputs[0], []byte(input)) &&
		samePackageSourceValue(row.Expected, []byte(expected)) && samePackageSourceValue(row.Actual, []byte(actual)) && jointSmokeBool(row.Passed, passed)
}
