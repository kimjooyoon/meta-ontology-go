package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (plan *Plan) UnmarshalJSON(payload []byte) error {
	type plainPlan Plan
	var wire struct {
		plainPlan
		Summary IndicatorDecisionSummary `json:"indicator_decision_summary"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	candidate := Plan(wire.plainPlan)
	if wire.Summary != candidate.indicatorDecisionSummary() {
		return fmt.Errorf("indicator decision summary does not match plan")
	}
	*plan = candidate
	return nil
}

func (manifest *ExecutionManifest) UnmarshalJSON(payload []byte) error {
	type plainManifest ExecutionManifest
	var wire struct {
		plainManifest
		Summary ExecutionDecisionSummary `json:"indicator_decision_summary"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	candidate := ExecutionManifest(wire.plainManifest)
	if wire.Summary != candidate.indicatorDecisionSummary() {
		return fmt.Errorf("indicator decision summary does not match execution manifest")
	}
	*manifest = candidate
	return nil
}
