package prooftransport

import "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-outcome"

func (report *Report) assessOutcome() {
	observation := transportoutcome.R_Observation{
		F_effect_id: report.Scope.F_effect_id, F_decision_effect_id: report.Decision.F_effect_id,
		F_decision_state: report.Decision.F_state, F_decision_next: report.Decision.F_next_operation,
		F_write_attempted: report.Write.Attempted, F_response_status: report.Write.Status, F_response_error: report.Write.Error,
		F_readback_recorded: report.After != nil && report.Next != nil,
	}
	if report.Next != nil {
		observation.F_readback_effect_id, observation.F_readback_state = report.Next.F_effect_id, report.Next.F_state
	}
	report.Outcome = transportoutcome.Evaluate(observation)
	report.OutcomeGenerationOrigin = transportoutcome.GoooGenerationOrigin
}
