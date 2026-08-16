package pressureshadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

const R4SafeSchemaVersion = "gooo/workfrontier-pressure-safe-r4/v1"

type R4SafeInput struct {
	Schema       string               `json:"schema"`
	R4Input      workfrontier.R4Input `json:"r4_input"`
	PathCoverage []PathCoverage       `json:"path_coverage"`
}

func (input *R4SafeInput) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoded, err := decodeR4SafeWire(data)
	if err != nil {
		return err
	}
	if decoded.Schema != R4SafeSchemaVersion {
		return fmt.Errorf("invalid schema")
	}
	if err := validateR4SafeSyntax(decoded); err != nil {
		return err
	}
	*input = decoded
	return nil
}

func decodeR4SafeWire(data []byte) (R4SafeInput, error) {
	type wire R4SafeInput
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wire
	if err := decoder.Decode(&decoded); err != nil {
		return R4SafeInput{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return R4SafeInput{}, fmt.Errorf("trailing JSON value")
	}
	return R4SafeInput(decoded), nil
}

type R4SafeResult struct {
	Schema               string                `json:"schema"`
	InputDigest          string                `json:"input_digest"`
	R4Result             workfrontier.R4Result `json:"r4_result"`
	R4ResultDigest       string                `json:"r4_result_digest"`
	PressureResult       S1B1Result            `json:"pressure_result"`
	PressureResultDigest string                `json:"pressure_result_digest"`
	Decision             Decision              `json:"decision"`
	Reason               Reason                `json:"reason"`
	FullSuiteRequired    bool                  `json:"full_suite_required"`
	SafeSelectedIDs      []string              `json:"safe_selected_ids"`
	SafeWorkIDs          []string              `json:"safe_work_ids"`
	ExecutionAuthorized  bool                  `json:"execution_authorized"`
	EnforcementEffect    EnforcementEffect     `json:"enforcement_effect"`
	ResultDigest         string                `json:"result_digest"`
	ReplayDigest         string                `json:"replay_digest"`
}

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
	if missing {
		return finishR4Safe(result, DecisionUnknown, ReasonRequiredInputMissing, true)
	}
	if pressure.Decision == DecisionFailClosed {
		return finishR4Safe(result, DecisionFailClosed, ReasonPressureCoverageFailClosed, true)
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
	err := json.Unmarshal(data, &input)
	if err != nil {
		result := newR4SafeResult(R4SafeInput{}, workfrontier.R4Result{}, S1B1Result{})
		result.InputDigest = digestBytes(append([]byte("r4-safe-invalid-input\x00"), data...))
		return finishR4Safe(result, DecisionFailClosed, ReasonInvalidInput, true)
	}
	return ValidateR4Safe(input)
}

func projectR4SafeInput(input R4SafeInput) Input {
	r4 := input.R4Input
	selector := workfrontier.Input{
		SchemaVersion: r4.SchemaVersion, SnapshotDigest: r4.SnapshotDigest,
		PolicyDigest: r4.PolicyDigest, RegistryDigest: r4.RegistryDigest,
		MinimumSelectedPressures: r4.MinimumSelectedPressures, Capacity: r4.Capacity,
		Pressures: append([]workfrontier.Pressure{}, r4.Pressures...),
		States:    append([]workfrontier.ObligationState{}, r4.States...),
		Paths:     append([]workfrontier.RepairPath{}, r4.Paths...),
	}
	selector.SchemaVersion = workfrontier.SchemaVersion
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
		SafeSelectedIDs: nil, SafeWorkIDs: nil, EnforcementEffect: EnforcementNoEffect,
	}
}

func finishR4Safe(result R4SafeResult, decision Decision, reason Reason, fullSuite bool) R4SafeResult {
	result.Decision, result.Reason, result.FullSuiteRequired = decision, reason, fullSuite
	result.ExecutionAuthorized = false
	result.EnforcementEffect = EnforcementNoEffect
	if decision != DecisionPass {
		result.SafeSelectedIDs, result.SafeWorkIDs = nil, nil
	}
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	result.ResultDigest = digestBytes(data)
	result.ReplayDigest = digestBytes([]byte("r4-safe-replay\x00" + result.InputDigest + "\x00" +
		result.R4ResultDigest + "\x00" + result.PressureResultDigest + "\x00" + result.ResultDigest))
	return result
}

func CanonicalR4SafeInputBytes(input R4SafeInput) ([]byte, error) {
	if err := validateR4SafeSyntax(input); err != nil {
		return nil, err
	}
	r4Bytes, err := workfrontier.EncodeR4JSON(input.R4Input)
	if err != nil {
		return nil, err
	}
	shadow, err := CanonicalInputBytes(projectR4SafeInput(input))
	if err != nil {
		return nil, err
	}
	var rows struct {
		PathCoverage json.RawMessage `json:"path_coverage"`
	}
	if err := json.Unmarshal(shadow, &rows); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Schema       string          `json:"schema"`
		R4Input      json.RawMessage `json:"r4_input"`
		PathCoverage json.RawMessage `json:"path_coverage"`
	}{input.Schema, bytes.TrimSpace(r4Bytes), rows.PathCoverage})
}

func validateR4SafeSyntax(input R4SafeInput) error {
	if input.Schema != R4SafeSchemaVersion || input.R4Input.SchemaVersion != workfrontier.R4SchemaVersion {
		return fmt.Errorf("invalid schema")
	}
	projected := projectR4SafeInput(input)
	return validateSyntax(projected)
}

func inspectR4Safe(input R4SafeInput) (invalid, missing bool) {
	projected := projectR4SafeInput(input)
	_, extra, _, _ := b1Issues(projected)
	if len(orphanPathIDs(selectorPathIDs(projected), coverageRows(projected)))+len(extra) > 0 {
		return true, false
	}
	registered, recorded := []string{}, []string{}
	for _, pressure := range input.R4Input.Pressures {
		registered = append(registered, pressureID(pressure))
	}
	for _, row := range input.PathCoverage {
		for _, record := range row.Coverage.PressureRecords {
			recorded = append(recorded, record.PressureID)
		}
	}
	if len(pressureDifference(recorded, registered)) > 0 {
		return true, false
	}
	return false, len(pressureDifference(registered, recorded)) > 0
}
