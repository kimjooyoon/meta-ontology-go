package sourcepolicy

import "encoding/json"

func (indicator Indicator) MarshalJSON() ([]byte, error) {
	type plainIndicator Indicator
	outcome := indicator.Outcome()
	return json.Marshal(struct {
		plainIndicator
		Decision          IndicatorDecision `json:"decision"`
		EvaluationState   EvaluationState   `json:"evaluation_state"`
		FailureReason     FailureReason     `json:"failure_reason"`
		FailureCode       string            `json:"failure_code,omitempty"`
		EnforcementEffect EnforcementEffect `json:"enforcement_effect"`
	}{
		plainIndicator:    plainIndicator(indicator),
		Decision:          outcome.Decision,
		EvaluationState:   outcome.EvaluationState,
		FailureReason:     outcome.FailureReason,
		FailureCode:       outcome.FailureCode,
		EnforcementEffect: outcome.EnforcementEffect,
	})
}
