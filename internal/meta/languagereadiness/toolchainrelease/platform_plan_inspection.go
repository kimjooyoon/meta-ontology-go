package toolchainrelease

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func validatePlanInputInspection(raw, source []byte, example planInputSmokeCase) (bodyexecution.CompositionInspection, error) {
	var observed struct {
		bodyexecution.CompositionInspection
		Calls *int `json:"model_calls"`
		Tests *int `json:"candidate_tests"`
		Runs  *int `json:"native_executions"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		return observed.CompositionInspection, err
	}
	expected, err := bodyexecution.InspectComposition(context.Background(), example.source, source, example.entry)
	if err != nil {
		return expected, err
	}
	if observed.Calls == nil || *observed.Calls != 0 || observed.Tests == nil || *observed.Tests != 0 ||
		observed.Runs == nil || *observed.Runs != 0 || !reflect.DeepEqual(observed.CompositionInspection, expected) {
		return expected, fmt.Errorf("native inspection differs from the exact structural source plan or zero counters")
	}
	return expected, nil
}

func validatePlanInputTemplate(raw []byte, plan bodyexecution.CompositionInspection) error {
	expected, err := bodyexecution.CompositionInputTemplate(plan)
	if err != nil {
		return err
	}
	var actualValue, expectedValue any
	if err := json.Unmarshal(raw, &actualValue); err != nil {
		return err
	}
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		return err
	}
	if !reflect.DeepEqual(actualValue, expectedValue) {
		return fmt.Errorf("input template differs from source-owned typed placeholders")
	}
	return nil
}
