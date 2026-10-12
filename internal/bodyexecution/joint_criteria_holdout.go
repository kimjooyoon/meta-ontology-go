package bodyexecution

import (
	"context"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type JointCriteriaHoldout struct {
	Evaluation     *bodycodegen.RecordCriteriaEvidence `json:"evaluation"`
	CallerExposure RecordHoldoutCallerExposure         `json:"caller_exposure"`
}

func jointHasCriteria(r JointConstruction) bool {
	for _, step := range r.Initial.ConstructionSteps() {
		if a := step.Generation.Report.RecordAssembly; a != nil && a.Criteria != nil {
			return true
		}
	}
	return false
}

func evaluateJointCriteriaHoldouts(ctx context.Context, filename string, source []byte,
	prior JointConstruction) ([]JointCriteriaHoldout, error) {
	if prior.SelectedAttempt < 0 {
		return nil, nil
	}
	var results []JointCriteriaHoldout
	for _, step := range prior.Initial.ConstructionSteps() {
		if r := step.Generation.Report.RecordAssembly; r == nil || r.Criteria == nil {
			continue
		}
		r, err := bodycodegen.EvaluateSourceRecordCriteriaHoldout(ctx, filename, source, []byte(prior.SelectedSource), step.Generation.Report.Activity)
		if err != nil {
			return results, err
		}
		if r != nil {
			results = append(results, JointCriteriaHoldout{Evaluation: r})
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	calls, err := prepareJointCallContext(ctx, filename, source, prior)
	if err != nil {
		return results, err
	}
	sets, known := jointHoldoutCallerInputs(calls, prior)
	for i := range results {
		// Reuse input-exposure measurement only; this adapter has no expected
		// values and is never a scoring suite or a published record receipt.
		e := results[i].Evaluation
		inputs := &bodycodegen.RecordHoldoutReceipt{ActivityID: e.ActivityID}
		for _, row := range e.Inputs {
			inputs.Cases = append(inputs.Cases, assemblyspec.ValueCase{Inputs: row.InputJSON})
		}
		results[i].CallerExposure = measureRecordHoldoutCallerExposure(calls, &sets, known, inputs)
	}
	return results, nil
}

func equalJointCriteriaHoldouts(a, b []JointCriteriaHoldout) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bodycodegen.EqualRecordCriteriaEvidence(a[i].Evaluation, b[i].Evaluation) || !reflect.DeepEqual(a[i].CallerExposure, b[i].CallerExposure) {
			return false
		}
	}
	return true
}
