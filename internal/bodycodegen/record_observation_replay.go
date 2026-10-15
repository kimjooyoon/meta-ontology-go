package bodycodegen

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func validateRecordObservationProbes(ctx context.Context, p recordAssemblyPlan) error {
	if isCriteriaQuestionPlan(p) {
		return validateCriteriaQuestionInputs(ctx, p)
	}
	if p.spec.Observation == nil {
		return nil
	}
	probes := make([]assemblyspec.ValueCase, len(p.spec.Observation.ValueInputs))
	for i, input := range p.spec.Observation.ValueInputs {
		probes[i] = assemblyspec.ValueCase{Inputs: input, Expected: p.spec.ValueCases[0].Expected}
	}
	return validateRecordAssemblyCases(ctx, p.body.base.source, p.body.activity.Name, p.body.records, probes)
}

func canonicalRecordObservation(r *RecordObservationReceipt) (*RecordObservationReceipt, error) {
	if r == nil {
		return nil, nil
	}
	result := *r
	var err error
	result.OracleInitialCases, err = canonicalRecordCaseValues(r.OracleInitialCases)
	return &result, err
}

// Reproduce questions only at their original source checkpoint. Subsequent
// checkpoints inherit those answers and check their current oracle against them.
func replayRecordObservation(ctx context.Context, filename string, source []byte, plans []recordAssemblyPlan,
	history []RecordAssemblyControlStage, prior, expected *RecordAssemblyReceipt) error {
	if isCriteriaQuestionPlan(plans[0]) {
		return replayCriteriaQuestions(ctx, source, plans, history, prior, expected)
	}
	if prior.CriteriaQuestions != nil {
		return fmt.Errorf("criteria questions are absent from source")
	}
	original := source
	if len(history) > 0 && history[0].Source != "" {
		original = []byte(history[0].Source)
	}
	r := newRecordAssemblyReceipt(original, plans[0])
	if err := observeRecordAssembly(ctx, filename, original, &plans[0], r); err != nil {
		return err
	}
	observed, err := canonicalRecordObservation(prior.Observation)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(r.Observation, observed) {
		return fmt.Errorf("record source observation does not replay")
	}
	expected.Observation = r.Observation
	if r.Observation == nil {
		return nil
	}
	current := original
	for i := 1; i < len(plans); i++ {
		if i == len(history) {
			current = source
		} else if history[i].Source != "" {
			current = []byte(history[i].Source)
		}
		plans[i].effectiveCases = append([]assemblyspec.ValueCase(nil), plans[0].effectiveCases...)
		check := &RecordObservationReceipt{Options: r.Observation.Options}
		if _, err := prepareRecordOracle(ctx, filename, current, plans[i], plans[i].effectiveCases, check); err != nil {
			return fmt.Errorf("record checkpoint %d changed retained answers: %w", i, err)
		}
		if check.OracleActivityID != r.Observation.OracleActivityID {
			return fmt.Errorf("record checkpoint oracle identity differs")
		}
	}
	return nil
}
