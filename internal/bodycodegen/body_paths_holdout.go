package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type PathHoldoutReceipt struct {
	Schema                  string                 `json:"schema"`
	Activity                string                 `json:"activity"`
	ActivityID              string                 `json:"activity_id"`
	DeclarationSourceSHA256 string                 `json:"declaration_source_sha256"`
	SelectedSourceSHA256    string                 `json:"selected_source_sha256"`
	GeneratedSHA256         string                 `json:"generated_sha256"`
	SuiteSHA256             string                 `json:"suite_sha256"`
	Cases                   []IRBodyFillTestCase   `json:"cases"`
	Results                 []IRBodyFillCaseResult `json:"results,omitempty"`
	Passed                  int                    `json:"passed"`
	Total                   int                    `json:"total"`
	AccuracyPercent         *float64               `json:"accuracy_percent,omitempty"`
	Error                   string                 `json:"error,omitempty"`
	Scope                   string                 `json:"scope"`
}

// EvaluateSourcePathHoldout evaluates only the already selected body. It neither
// chooses a body nor calls a model. original owns the withheld source contract;
// selected may contain later caller-guided choices in other activities.
func EvaluateSourcePathHoldout(ctx context.Context, filename string, original, selected []byte,
	activity string) (*PathHoldoutReceipt, error) {
	spec, err := SourceAssembly(ctx, filename, original, activity)
	if err != nil || spec == nil || len(spec.HoldoutCases) == 0 || !IsSourceTypedPathAssembly(spec) {
		return nil, err
	}
	if len(selected) == 0 || len(selected) > 128<<10 {
		return nil, fmt.Errorf("path holdout requires bounded selected source")
	}
	base, err := prepareActivityBodySource(filename, original, activity)
	if err != nil {
		return nil, err
	}
	projection, err := GenerateWithPlanner(ctx, filename, selected, activity, "", "")
	if err != nil {
		return nil, err
	}
	if projection.Report.ActivityID != base.activityID || projection.Report.InputType != "int64" || projection.Report.OutputType != "int64" {
		return nil, fmt.Errorf("path holdout selected activity identity or Integer signature differs")
	}
	return evaluateSelectedPathHoldout(ctx, original, selected, spec, projection)
}

func evaluateSelectedPathHoldout(ctx context.Context, original, selected []byte, spec *assemblyspec.Spec,
	projection Result) (*PathHoldoutReceipt, error) {
	if spec == nil || len(spec.HoldoutCases) == 0 {
		return nil, nil
	}
	cases := make([]IRBodyFillTestCase, len(spec.HoldoutCases))
	for i, c := range spec.HoldoutCases {
		cases[i] = IRBodyFillTestCase{Input: c.Input, Expected: c.Expected}
	}
	raw, _ := json.Marshal(cases)
	r := &PathHoldoutReceipt{Schema: "gooo/typed-path-holdout/v1", Activity: projection.Report.Activity,
		ActivityID: projection.Report.ActivityID, DeclarationSourceSHA256: digest(original), SelectedSourceSHA256: digest(selected),
		GeneratedSHA256: projection.Report.GeneratedDigest, SuiteSHA256: digest(raw), Cases: cases, Total: len(cases),
		Scope: "source holdout cases after this body's selection; excluded from local path ranking and observation; bounded Go AST evaluation, no candidate reselection or native execution; caller exposure is a separate observation"}
	var err error
	r.Results, r.Passed, err = evaluateIntegerCasesContext(ctx, []byte(projection.Source), r.Activity, cases)
	if err != nil {
		r.Error = err.Error()
	} else {
		accuracy := 100 * float64(r.Passed) / float64(r.Total)
		r.AccuracyPercent = &accuracy
	}
	return r, ctx.Err()
}

func pathHoldoutDimension(r *PathHoldoutReceipt) CompletenessDimension {
	d := completenessDimension("typed_path_holdout_accuracy", r.Passed, r.Total,
		"source holdout cases matched after body selection", "These cases never rank candidates or extend observation questions; caller exposure is measured separately.",
		[]string{"holdout_suite_sha256:" + r.SuiteSHA256, "selected_source_sha256:" + r.SelectedSourceSHA256}, false)
	if r.AccuracyPercent == nil {
		d.Status, d.Reason = "UNKNOWN", "Holdout evaluation unavailable: "+r.Error
	} else if r.Passed == 0 {
		d.Status = "PROGRESS"
	}
	return d
}
