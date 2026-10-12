package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
)

// The original source owns the fixed predicates. Only the chosen candidate is
// projected from the final caller source; no held outcome participates in search.
func EvaluateSourceRecordCriteriaHoldout(ctx context.Context, filename string, original, selected []byte,
	activity string) (*RecordCriteriaEvidence, error) {
	spec, err := SourceAssembly(ctx, filename, original, activity)
	if err != nil || spec == nil || spec.Criteria == nil || len(spec.Criteria.HoldoutInputs) == 0 {
		return nil, err
	}
	if len(selected) == 0 || len(selected) > 128<<10 {
		return nil, fmt.Errorf("criteria holdout requires bounded selected source")
	}
	base, err := prepareActivityBodySource(filename, original, activity)
	if err != nil {
		return nil, err
	}
	current, err := prepareActivityBodySource(filename, selected, activity)
	if err != nil {
		return nil, err
	}
	if base.activityID != current.activityID || base.outputType != current.outputType ||
		!reflect.DeepEqual(base.parameters, current.parameters) || !reflect.DeepEqual(base.allRecords, current.allRecords) {
		return nil, fmt.Errorf("criteria holdout selected identity or typed declarations differ")
	}
	criteria, err := prepareRecordCriteria(ctx, filename, original, base, spec.Criteria)
	if err != nil {
		return nil, err
	}
	projection, err := GenerateWithPlanner(ctx, filename, selected, activity, "", "")
	if err != nil {
		return nil, err
	}
	p := recordAssemblyPlan{spec: spec, body: base, criteria: criteria, sourceSHA256: digest(original)}
	return evaluateRecordCriteria(ctx, p, generatedRoute{source: []byte(projection.Source), report: projection.Report}, spec.Criteria.HoldoutInputs, "holdout")
}

func EqualRecordCriteriaEvidence(a, b *RecordCriteriaEvidence) bool {
	return equalLocalCallEvidence(a, b)
}
