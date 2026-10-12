package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
)

// EvaluateSourceRecordHoldout evaluates a final callable source after a caller's
// choice. The original Gooo declaration still owns the withheld expectations.
func EvaluateSourceRecordHoldout(ctx context.Context, filename string, original, selected []byte,
	activity string) (*RecordHoldoutReceipt, error) {
	spec, err := SourceAssembly(ctx, filename, original, activity)
	if err != nil || spec == nil || len(spec.ValueHoldoutCases) == 0 || len(spec.Choices) == 0 || !IsRecordAssembly(spec) {
		return nil, err
	}
	if len(selected) == 0 || len(selected) > 128<<10 {
		return nil, fmt.Errorf("record holdout requires bounded selected source")
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
		return nil, fmt.Errorf("record holdout selected identity or typed declarations differ")
	}
	projection, err := GenerateWithPlanner(ctx, filename, selected, activity, "", "")
	if err != nil {
		return nil, err
	}
	return evaluateProjectedRecordHoldout(ctx, original, selected, spec.ValueHoldoutCases, projection)
}

// EqualRecordHoldoutObservations normalizes only JSON value formatting. Every
// source binding, result, error, count and declared expected value must match.
func EqualRecordHoldoutObservations(a, b *RecordHoldoutReceipt) bool {
	left, err := canonicalRecordHoldout(a)
	if err != nil {
		return false
	}
	right, err := canonicalRecordHoldout(b)
	return err == nil && reflect.DeepEqual(left, right)
}
