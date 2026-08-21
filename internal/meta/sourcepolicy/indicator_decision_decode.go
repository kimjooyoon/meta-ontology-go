package sourcepolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (indicator *Indicator) UnmarshalJSON(payload []byte) error {
	type plainIndicator Indicator
	var wire struct {
		plainIndicator
		Decision          IndicatorDecision `json:"decision"`
		EvaluationState   EvaluationState   `json:"evaluation_state"`
		FailureReason     FailureReason     `json:"failure_reason"`
		FailureCode       string            `json:"failure_code,omitempty"`
		EnforcementEffect EnforcementEffect `json:"enforcement_effect"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	candidate := Indicator(wire.plainIndicator)
	actual := IndicatorOutcome{
		Decision: wire.Decision, EvaluationState: wire.EvaluationState,
		FailureReason: wire.FailureReason, FailureCode: wire.FailureCode,
		EnforcementEffect: wire.EnforcementEffect,
	}
	if actual != candidate.Outcome() {
		return fmt.Errorf("indicator outcome does not match metric fact")
	}
	*indicator = candidate
	return nil
}
