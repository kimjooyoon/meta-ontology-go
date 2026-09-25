package generation

import "encoding/json"

func (plan Plan) MarshalJSON() ([]byte, error) {
	type plainPlan Plan
	return json.Marshal(struct {
		plainPlan
		IndicatorDecisionSummary IndicatorDecisionSummary `json:"indicator_decision_summary"`
	}{
		plainPlan:                plainPlan(plan),
		IndicatorDecisionSummary: plan.indicatorDecisionSummary(),
	})
}

func (manifest ExecutionManifest) MarshalJSON() ([]byte, error) {
	type plainManifest ExecutionManifest
	return json.Marshal(struct {
		plainManifest
		IndicatorDecisionSummary ExecutionDecisionSummary `json:"indicator_decision_summary"`
	}{
		plainManifest:            plainManifest(manifest),
		IndicatorDecisionSummary: manifest.indicatorDecisionSummary(),
	})
}
