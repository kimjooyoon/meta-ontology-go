package pressureshadow

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func ValidateR4Safe(input R4SafeInput) R4SafeResult {
	r4 := workfrontier.EvaluateR4(input.R4Input)
	pressure := ValidateS1B1(projectR4SafeInput(input))
	result := newR4SafeResult(input, r4, pressure)
	if err := validateR4SafeSyntax(input); err != nil {
		return finishR4Safe(result, DecisionFailClosed, ReasonInvalidInput, true)
	}
	invalid, missing := inspectR4Safe(input)
	if invalid {
		return finishR4Safe(result, DecisionFailClosed, ReasonInvalidInput, true)
	}
	switch r4.Status {
	case workfrontier.R4StatusFailClosed:
		return finishR4Safe(result, DecisionFailClosed, ReasonUpstreamFailClosed, true)
	case workfrontier.R4StatusUnknown:
		return finishR4Safe(result, DecisionUnknown, ReasonUpstreamUnknown, true)
	}
	if pressure.Decision == DecisionFailClosed {
		return finishR4Safe(result, DecisionFailClosed, ReasonPressureCoverageFailClosed, true)
	}
	if missing {
		return finishR4Safe(result, DecisionUnknown, ReasonRequiredInputMissing, true)
	}
	if pressure.Decision == DecisionUnknown {
		return finishR4Safe(result, DecisionUnknown, ReasonPressureCoverageUnknown, true)
	}
	if r4.Status == workfrontier.R4StatusPass {
		result.SafeSelectedIDs = append([]string{}, r4.SelectedIDs...)
		result.SafeWorkIDs = append([]string{}, r4.WorkIDs...)
	}
	return finishR4Safe(result, DecisionPass, ReasonNone, false)
}
func ValidateR4SafeBytes(data []byte) R4SafeResult {
	var input R4SafeInput
	if err := json.Unmarshal(data, &input); err != nil {
		result := newR4SafeResult(R4SafeInput{}, workfrontier.R4Result{}, S1B1Result{})
		result.InputDigest = digestBytes(append([]byte("r4-safe-invalid-input\x00"), data...))
		return finishR4Safe(result, DecisionFailClosed, ReasonInvalidInput, true)
	}
	return ValidateR4Safe(input)
}
func projectR4SafeInput(input R4SafeInput) Input {
	r4 := input.R4Input
	selector := workfrontier.Input{
		SchemaVersion: workfrontier.SchemaVersion, SnapshotDigest: r4.SnapshotDigest,
		PolicyDigest: r4.PolicyDigest, RegistryDigest: r4.RegistryDigest,
		MinimumSelectedPressures: r4.MinimumSelectedPressures, Capacity: r4.Capacity,
		Pressures: append([]workfrontier.Pressure{}, r4.Pressures...),
		States:    append([]workfrontier.ObligationState{}, r4.States...),
		Paths:     append([]workfrontier.RepairPath{}, r4.Paths...),
	}
	return Input{Schema: SchemaVersion, Selector: selector, PathCoverage: input.PathCoverage}
}
func newR4SafeResult(input R4SafeInput, r4 workfrontier.R4Result, pressure S1B1Result) R4SafeResult {
	raw, _ := workfrontier.EncodeR4ResultJSON(r4)
	inputBytes, err := CanonicalR4SafeInputBytes(input)
	if err != nil {
		inputBytes = []byte("r4-safe-input-error:" + err.Error())
	}
	return R4SafeResult{
		Schema: R4SafeSchemaVersion, InputDigest: digestBytes(inputBytes),
		R4Result: r4, R4ResultDigest: digestBytes(append([]byte("r4-safe-r4-result\x00"), raw...)),
		PressureResult: pressure, PressureResultDigest: CanonicalS1B1ResultDigest(pressure),
	}
}
