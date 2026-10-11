package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type helperFillDetails struct {
	Initial struct {
		Preparations []struct {
			InputSHA   string `json:"input_source_sha256"`
			Generation struct {
				Report struct {
					ActivityID  string                       `json:"activity_id"`
					CallClosure *bodycodegen.PureCallClosure `json:"call_closure"`
					BodyFill    *helperFillReceipt           `json:"body_fill"`
				}
			}
		}
		Validations []struct {
			Stage       string
			Observation bodycodegen.SourceFillValidation
		} `json:"fill_validations"`
	}
	Attempts []struct {
		Fill []bodycodegen.FillCandidate `json:"fill_candidates"`
	}
}

type helperFillReceipt struct {
	bodycodegen.IRBodyFillReceipt
	Predictions *int  `json:"local_model_predictions"`
	Calls       *int  `json:"external_provider_calls"`
	Known       *bool `json:"external_provider_calls_known"`
}

func validateHelperFillHistory(raw, source, feedback, evaluation []byte, replay bool, retained string) (string, error) {
	var r jointSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	var envelope jointConstructionEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", err
	}
	historySHA, err := checkpointConstructionDigest(envelope.Construction)
	if err != nil || replay && (retained == "" || retained != historySHA) {
		return "", fmt.Errorf("helper fill replay changed its original construction history")
	}
	c, e := r.Construction, r.Evaluation
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v5" ||
		c.Stage != "COMPLETE" || c.Decision != "COMPLETE_FINITE" || c.Failure != "" ||
		c.StopReason != "LOCAL_AND_CALLER_CASES_MATCHED" || !jointSmokeInt(c.Budget, 4) || c.Space != "4" ||
		!jointSmokeInt(c.SelectedAttempt, 0) || len(c.Attempts) != 1 || c.Selected.SHA == "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) || c.Source == "" ||
		strings.Contains(c.Source, "__GOOO_BODY_HOLE_") || strings.Contains(c.Source, "assembling") ||
		!slices.Equal(c.Kinds, []string{"record_mask", "source_fill_index"}) ||
		!jointSmokeBool(c.Initial.Model.Loaded, false) || c.Initial.FillModel.Loaded != nil && *c.Initial.FillModel.Loaded ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) {
		return "", fmt.Errorf("helper fill source, bounded selection or inference differs")
	}
	a := c.Attempts[0]
	if a.Rejection != nil || !slices.Equal(a.Masks, []int{0, 1}) || !jointSmokeInt(a.Passed, 3) ||
		!jointSmokeInt(a.Total, 3) || a.Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA {
		return "", fmt.Errorf("helper fill selected a different source combination")
	}
	for i, suite := range [][]byte{feedback, evaluation} {
		count, runtime := 1, a.Runtime
		if i == 1 {
			count, runtime = 3, e.Runtime
		}
		rows, err := jointSmokeCases(suite, count)
		if err != nil {
			return "", err
		}
		if err := validateJointFillRuntimeFor(runtime, rows, nil, "Main", "fillhistory://activity/main"); err != nil {
			return "", err
		}
	}
	separation := envelope.Evaluation.Runtime.Separation
	if separation.Status != "UNKNOWN" || !jointSmokeInt(separation.Unknown, 1) ||
		!jointSmokeInt(separation.Overlap, 2) || !jointSmokeInt(separation.Disjoint, 0) {
		return "", fmt.Errorf("helper fill lost overlap or incomplete input coverage")
	}
	var detail helperFillDetails
	if err := json.Unmarshal(envelope.Construction, &detail); err != nil {
		return "", err
	}
	if err := validateHelperFillDetails(detail, c.OriginalSHA, fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(c.Source)))); err != nil {
		return "", err
	}
	return historySHA, nil
}

func validateHelperFillDetails(d helperFillDetails, original, selected string) error {
	if len(d.Initial.Preparations) != 2 || len(d.Initial.Validations) != 1 ||
		len(d.Attempts) != 1 || len(d.Attempts[0].Fill) != 1 {
		return fmt.Errorf("helper fill preparation, preflight or assignment is missing")
	}
	step := d.Initial.Preparations[1]
	f, closure := step.Generation.Report.BodyFill, step.Generation.Report.CallClosure
	if step.InputSHA == "" || step.InputSHA == original || f == nil || closure == nil || len(closure.Activities) != 1 ||
		step.Generation.Report.ActivityID != "fillhistory://activity/build" ||
		f.OriginalSourceDigest != step.InputSHA || f.IRPlanSHA256 == "" || f.SelectedCandidateID != "works" ||
		!jointSmokeInt(f.Predictions, 0) || !jointSmokeBool(f.Known, true) || !jointSmokeInt(f.Calls, 0) ||
		f.TestCasesPassed != 2 || f.TestCasesTotal != 2 || f.HoldoutCasesPassed != 1 || f.HoldoutCasesTotal != 1 ||
		len(f.RejectedCandidates) != 1 || f.RejectedCandidates[0].CandidateID != "fails" ||
		f.RejectedCandidates[0].Stage != "TRAINING_EVALUATION" || !strings.Contains(f.RejectedCandidates[0].Reason, "zero divisor") ||
		len(f.CandidateScores) != 1 || f.CandidateScores[0].ID != "works" ||
		!f.CandidateScores[0].TypecheckPassed || f.CandidateScores[0].TestCasesPassed != 2 || f.CandidateScores[0].TestCasesTotal != 2 {
		return fmt.Errorf("helper fill erased a rejection or changed finite selection")
	}
	helper := closure.Activities[0]
	if helper.Name != "Seed" || helper.ActivityID != "fillhistory://activity/seed" || helper.ProgramSHA256 == "" {
		return fmt.Errorf("helper fill lost its called activity identity")
	}
	if err := validateHelperFillPasses(f.LocalCalls, helper.ProgramSHA256, false); err != nil {
		return err
	}
	v := d.Initial.Validations[0]
	if v.Stage != "ACTIVITY_GENERATION" || v.Observation.Schema != "gooo/source-fill-validation/v1" ||
		v.Observation.SourceSHA256 != step.InputSHA || v.Observation.Activity != "Build" ||
		v.Observation.ActivityID != "fillhistory://activity/build" || !v.Observation.Completed || v.Observation.Error != "" {
		return fmt.Errorf("helper fill preflight lost its prepared source checkpoint")
	}
	if err := validateHelperFillPasses(v.Observation.LocalCalls, helper.ProgramSHA256, false); err != nil {
		return err
	}
	x := d.Attempts[0].Fill[0]
	want := []bodycodegen.IRBodyFillHoleFill{{HoleID: "seed", Expression: "input - 6"}, {HoleID: "divisor", Expression: "input"}}
	if x.Rejection != nil || x.Schema != "gooo/fill-candidate/v1" || x.ActivityID != "fillhistory://activity/build" ||
		x.InputSourceSHA256 != step.InputSHA || x.PlanSHA256 != f.IRPlanSHA256 || x.SelectedSourceSHA256 != selected ||
		x.CandidateID != "works" || x.CandidateCount != 2 || x.SelectionMethod != "caller_selected_assignment" ||
		!slices.Equal(x.HoleFills, want) || x.TestCasesPassed != 2 || x.TestCasesTotal != 2 ||
		x.HoldoutCasesPassed != 1 || x.HoldoutCasesTotal != 1 {
		return fmt.Errorf("helper fill caller assignment differs")
	}
	if err := validateHelperFillPasses(x.LocalCalls, helper.ProgramSHA256, true); err != nil {
		return err
	}
	for _, rows := range [][]bodycodegen.IRBodyFillCaseResult{f.SelectedCaseResults, x.CaseResults} {
		if !helperFillCasesMatch(rows, [][2]int64{{4, 1}, {2, 2}}) {
			return fmt.Errorf("helper fill training values differ")
		}
	}
	for _, rows := range [][]bodycodegen.IRBodyFillCaseResult{f.HoldoutCaseResults, x.HoldoutCaseResults} {
		if !helperFillCasesMatch(rows, [][2]int64{{3, 1}}) {
			return fmt.Errorf("helper fill separate holdout values differ")
		}
	}
	return nil
}

func helperFillCasesMatch(rows []bodycodegen.IRBodyFillCaseResult, want [][2]int64) bool {
	if len(rows) != len(want) {
		return false
	}
	for i, row := range rows {
		if row.Input != want[i][0] || len(row.Inputs) != 0 || row.Expected != want[i][1] || row.Actual != want[i][1] || !row.Passed {
			return false
		}
	}
	return true
}
