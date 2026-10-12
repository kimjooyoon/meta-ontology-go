package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type RecordHoldoutReceipt struct {
	Activity                string                   `json:"activity"`
	ActivityID              string                   `json:"activity_id"`
	Schema                  string                   `json:"schema"`
	DeclarationSourceSHA256 string                   `json:"declaration_source_sha256"`
	SelectedSourceSHA256    string                   `json:"selected_source_sha256"`
	GeneratedSHA256         string                   `json:"generated_sha256"`
	SuiteSHA256             string                   `json:"suite_sha256"`
	Cases                   []assemblyspec.ValueCase `json:"cases"`
	Results                 []RecordAssemblyCase     `json:"results,omitempty"`
	Passed                  int                      `json:"passed"`
	Total                   int                      `json:"total"`
	FieldsPassed            int                      `json:"fields_passed"`
	FieldsTotal             int                      `json:"fields_total"`
	AccuracyPercent         *float64                 `json:"accuracy_percent,omitempty"`
	Error                   string                   `json:"error,omitempty"`
	Scope                   string                   `json:"scope"`
}

// emitSelectedRecordAssembly belongs to a completed local selection. Explicit
// candidate projection keeps using emitRecordAssembly and never evaluates holds.
func emitSelectedRecordAssembly(ctx context.Context, filename string, source []byte, p recordAssemblyPlan,
	r *RecordAssemblyReceipt) (Result, error) {
	result, err := emitRecordAssembly(ctx, filename, source, p, r)
	if err != nil {
		return Result{}, err
	}
	r.Holdout, err = evaluateProjectedRecordHoldout(ctx, source, []byte(result.GoooSource), p.spec.ValueHoldoutCases, result)
	if err == nil && p.spec.Criteria != nil {
		candidate := generatedRoute{source: []byte(result.Source), report: result.Report}
		r.CriteriaHoldout, err = evaluateRecordCriteria(ctx, p, candidate, p.spec.Criteria.HoldoutInputs, "holdout")
	}
	return result, err
}

// Selection has already ended. Held outcomes cannot change its mask, ranking,
// budget or policy decisions. Replay and continuation evaluate their own body.
func evaluateProjectedRecordHoldout(ctx context.Context, original, selected []byte, held []assemblyspec.ValueCase,
	projection Result) (*RecordHoldoutReceipt, error) {
	if len(held) == 0 {
		return nil, nil
	}
	cases := append([]assemblyspec.ValueCase(nil), held...)
	raw, _ := json.Marshal(cases)
	r := &RecordHoldoutReceipt{Activity: projection.Report.Activity, ActivityID: projection.Report.ActivityID, Schema: "gooo/record-field-holdout/v1", DeclarationSourceSHA256: digest(original),
		SelectedSourceSHA256: digest(selected), GeneratedSHA256: projection.Report.GeneratedDigest,
		SuiteSHA256: digest(raw), Cases: cases, Total: len(cases),
		Scope: "source holdout values after record field selection; excluded from ranking, policy and stopping; bounded AST evaluation; caller and model-training exposure are unmeasured"}
	if record := recordTypeByName(projection.Report.RecordTypes, projection.Report.OutputType); record != nil {
		r.FieldsTotal = r.Total * len(record.Fields)
	}
	var err error
	r.Results, err = evaluateRecordAssemblyRecorded(ctx, []byte(projection.Source), projection.Report.Activity, projection.Report.RecordTypes, cases, nil)
	if err != nil {
		r.Error = err.Error()
	} else {
		score := scoreRecordCases(0, r.Results)
		r.Passed, r.FieldsPassed = score.Passed, score.FieldsPassed
		accuracy := 100 * float64(r.Passed) / float64(r.Total)
		r.AccuracyPercent = &accuracy
	}
	return r, ctx.Err()
}

func canonicalRecordHoldout(r *RecordHoldoutReceipt) (*RecordHoldoutReceipt, error) {
	if r == nil {
		return nil, nil
	}
	result := *r
	var err error
	result.Results, err = canonicalRecordCaseValues(r.Results)
	return &result, err
}

func verifyRecordHoldout(expected, actual *RecordHoldoutReceipt) error {
	canonical, err := canonicalRecordHoldout(actual)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, canonical) {
		return fmt.Errorf("record holdout values do not replay")
	}
	return nil
}

func recordHoldoutDimensions(r *RecordHoldoutReceipt) []CompletenessDimension {
	evidence := []string{"record_holdout.suite_sha256:" + r.SuiteSHA256, "record_holdout.selected_source_sha256:" + r.SelectedSourceSHA256}
	dimensions := []CompletenessDimension{
		completenessDimension("record_holdout_case_accuracy", r.Passed, r.Total, "source record holdout cases after selection",
			"Separate evaluation; caller and model-training exposure are unmeasured.", evidence, false),
		completenessDimension("record_holdout_field_accuracy", r.FieldsPassed, r.FieldsTotal, "source record holdout output fields after selection",
			"Expected fields are counted separately from whole-record matches.", evidence, false),
	}
	for i := range dimensions {
		if r.AccuracyPercent == nil {
			dimensions[i].Status, dimensions[i].Reason = "UNKNOWN", "Record holdout evaluation unavailable: "+r.Error
		} else if dimensions[i].Numerator == 0 {
			dimensions[i].Status = "PROGRESS"
		}
	}
	return dimensions
}
