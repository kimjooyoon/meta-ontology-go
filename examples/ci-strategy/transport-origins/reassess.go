package transportorigins

import (
	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
	resolution "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-resolution"
)

// ReassessSavedReports retains the prior assessment and compares a read-only
// report when supplied. A nil current report asks what observation is needed.
func ReassessSavedReports(priorRaw, currentRaw []byte) (resolution.R_Resolution, error) {
	prior, priorFrame, err := decodeSavedReport(priorRaw)
	if err != nil {
		return resolution.R_Resolution{}, err
	}
	priorView := projectSavedReport(prior, priorFrame)
	var current transport.Report
	var currentView R_OriginView
	if currentRaw != nil {
		var frame R_SavedFrame
		current, frame, err = decodeSavedReport(currentRaw)
		if err != nil {
			return resolution.R_Resolution{}, err
		}
		currentView = projectSavedReport(current, frame)
	}
	frame := resolution.R_PairFrame{
		F_generation_sha256: contextDigest(resolution.GoooGenerationOrigin),
		F_prior_operation:   prior.Scope.F_operation, F_current_operation: current.Scope.F_operation,
		F_current_present: currentRaw != nil, F_current_apply_requested: current.ApplyRequested,
		F_current_write_attempted: current.Write.Attempted, F_current_readback_present: current.After != nil,
	}
	return resolution.Evaluate(resolution.R_Outcome(prior.Outcome), resolution.R_TransportObservation(current.Before),
		resolution.R_OriginView(priorView), resolution.R_OriginView(currentView), frame), nil
}
