package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func validateJointFillRejectionSmoke(raw, source, feedback, evaluation []byte, budget int, replay bool, selected string) (string, error) {
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
	index, decision, stop := 4, "COMPLETE_FINITE", "LOCAL_AND_CALLER_CASES_MATCHED"
	if budget == 3 {
		index, decision, stop = 0, "PARTIAL_FINITE", "PROGRAM_BUDGET_EXHAUSTED"
	}
	if budget != 3 && budget != 5 || replay && budget != 5 || !jointSmokeBool(r.Generated, !replay) ||
		c.Schema != "gooo/joint-construction/v5" || c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != decision || c.StopReason != stop ||
		!jointSmokeInt(c.Budget, budget) || c.Space != "5" || !jointSmokeInt(c.SelectedAttempt, index) || len(c.Attempts) != budget ||
		!slices.Equal(c.Kinds, []string{"source_fill_index"}) || c.Source == "" || c.Selected.SHA == "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) ||
		strings.Contains(c.Source, "__GOOO_BODY_HOLE_") || strings.Contains(c.Source, "assembling") ||
		c.Initial.Model.Loaded != nil && *c.Initial.Model.Loaded || c.Initial.FillModel.Loaded != nil && *c.Initial.FillModel.Loaded ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) ||
		!jointSmokeInt(e.Separation.Consumed, 0) || !jointSmokeInt(e.Separation.Other, 4) {
		return "", fmt.Errorf("rejected-fill identity, budget or input separation differs")
	}
	if replay && (selected == "" || selected != c.Selected.SHA) {
		return "", fmt.Errorf("rejected-fill replay changed the selected program")
	}
	if err := validateJointFillRejectionInitial(r); err != nil {
		return "", err
	}
	plan := c.Initial.Preparations[0].Generation.Report.Fill.PlanSHA
	for i, a := range c.Attempts {
		if len(a.FillCandidates) != 1 || a.FillCandidates[0].PlanSHA != plan {
			return "", fmt.Errorf("rejected-fill plan or candidate count differs")
		}
		if i == 1 || i == 2 {
			if err := validateJointRejectedFill(a, i, c.OriginalSHA); err != nil {
				return "", err
			}
			continue
		}
		local := i
		if i > 0 {
			local -= 2
		}
		if err := validateJointFillLocalAt(a, local, i, 5, c.OriginalSHA); err != nil {
			return "", err
		}
		actual := []json.RawMessage{json.RawMessage([]string{"9", "-7", "-8"}[local])}
		if err := validateJointFillRuntime(a.Runtime, construction, actual); err != nil {
			return "", err
		}
	}
	if c.Attempts[index].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA ||
		c.Attempts[index].FillCandidates[0].SelectedSHA != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(c.Source))) {
		return "", fmt.Errorf("rejected-fill selected source differs from execution")
	}
	var actual []json.RawMessage
	if budget == 3 {
		actual = []json.RawMessage{[]byte("2"), []byte("-12"), []byte("5"), []byte("9007199254740994")}
	}
	if err := validateJointFillRuntime(e.Runtime, cases, actual); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}

func validateJointFillRejectionInitial(r jointSmokeOutput) error {
	if len(r.Construction.Initial.Preparations) != 1 {
		return fmt.Errorf("rejected-fill initial count differs")
	}
	i := r.Construction.Initial.Preparations[0].Generation.Report.Fill
	if i == nil || i.Selected != "late_unbounded" || i.PlanSHA == "" || !jointSmokeInt(i.Calls, 0) ||
		len(i.Scores) != 3 || len(i.Rejected) != 2 {
		return fmt.Errorf("rejected-fill initial score/rejection partition differs")
	}
	for index, score := range i.Scores {
		if score.ID != []string{"late_unbounded", "early_wrong_cap", "bounded"}[index] ||
			!jointSmokeBool(score.Typed, true) || !jointSmokeInt(score.Passed, 1) || !jointSmokeInt(score.Total, 1) {
			return fmt.Errorf("rejected-fill initial score differs")
		}
	}
	for index, rejection := range i.Rejected {
		if err := validateJointFillRejection(rejection, index+1); err != nil {
			return err
		}
		attempt := r.Construction.Attempts[index+1]
		if attempt.Rejection == nil || attempt.Rejection.Reason != rejection.Reason {
			return fmt.Errorf("initial and whole-program rejection reasons differ")
		}
	}
	return nil
}
