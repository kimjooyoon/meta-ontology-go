package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func validateNativeArithmeticSmoke(raw, source, feedback, evaluation []byte, budget int, replay bool, selected string) (string, error) {
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
	if err = validateNativeArithmeticIdentity(r, source, budget, replay, selected); err != nil {
		return "", err
	}
	if err = validateJointFillInitialIDs(r, []string{"late_unbounded", "early_wrong_cap", "caller_zero_divisor", "bounded"}); err != nil {
		return "", err
	}
	var native struct {
		Construction struct {
			Cases    json.RawMessage `json:"construction_cases"`
			Attempts []struct{ Runtime nativeSmokeRuntime }
		}
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		return "", err
	}
	if !samePackageSourceValue(native.Construction.Cases, feedback) {
		return "", fmt.Errorf("native arithmetic original cases differ")
	}
	if err = validateNativeArithmeticAttempts(r, native.Construction.Attempts[4].Runtime, construction); err != nil {
		return "", err
	}
	return validateNativeArithmeticFinal(r, cases, budget)
}

func validateNativeArithmeticIdentity(r jointSmokeOutput, source []byte, budget int, replay bool, selected string) error {
	c, e := r.Construction, r.Evaluation
	index, decision, stop := 5, "COMPLETE_FINITE", "LOCAL_AND_CALLER_CASES_MATCHED"
	if budget == 5 {
		index, decision, stop = 0, "PARTIAL_FINITE", "PROGRAM_BUDGET_EXHAUSTED"
	}
	if budget != 5 && budget != 6 || replay && budget != 6 || !jointSmokeBool(r.Generated, !replay) ||
		c.Schema != "gooo/joint-construction/v6" || c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != decision || c.StopReason != stop ||
		!jointSmokeInt(c.Budget, budget) || c.Space != "6" || !jointSmokeInt(c.SelectedAttempt, index) || len(c.Attempts) != budget ||
		!slices.Equal(c.Kinds, []string{"source_fill_index"}) || c.Source == "" || c.Selected.SHA == "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) || strings.Contains(c.Source, "__GOOO_BODY_HOLE_") || strings.Contains(c.Source, "assembling") ||
		c.Initial.Model.Loaded != nil && *c.Initial.Model.Loaded || c.Initial.FillModel.Loaded != nil && *c.Initial.FillModel.Loaded ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) || !jointSmokeInt(e.Separation.Consumed, 0) || !jointSmokeInt(e.Separation.Other, 4) {
		return fmt.Errorf("native arithmetic identity, selection or input separation differs")
	}
	if replay && (selected == "" || selected != c.Selected.SHA) {
		return fmt.Errorf("native arithmetic replay changed selected program")
	}
	return nil
}

func validateNativeArithmeticFinal(r jointSmokeOutput, cases []jointSmokeCase, budget int) (string, error) {
	c, e := r.Construction, r.Evaluation
	index := 5
	var actual []json.RawMessage
	if budget == 5 {
		index = 0
		actual = []json.RawMessage{[]byte("2"), []byte("-12"), []byte("5"), []byte("9007199254740994")}
	}
	if c.Attempts[index].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA ||
		c.Attempts[index].FillCandidates[0].SelectedSHA != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(c.Source))) {
		return "", fmt.Errorf("native arithmetic selected source differs from execution")
	}
	if err := validateJointFillRuntime(e.Runtime, cases, actual); err != nil {
		return "", err
	}
	if err := validateNativeSmokeRuns(e.Runtime.Runs); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}
