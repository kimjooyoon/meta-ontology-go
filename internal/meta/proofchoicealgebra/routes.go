package proofchoicealgebra

func selectRoute(value Value, bundle evidenceBundle) routeResult {
	candidates := []Evidence{}
	unknown := false
	for _, evidence := range evidenceFor(value, bundle) {
		if evidence.State == "OBSERVED" {
			candidates = append(candidates, evidence)
		}
		if evidence.State == UnknownState {
			unknown = true
		}
	}
	if len(candidates) > 1 {
		return routeResult{Resolution: Exact, Reason: "PROOF_ROUTE_CONTRADICTION", FailClosed: true}
	}
	if len(candidates) == 0 {
		state := InsufficientState
		reason := "EVIDENCE_INSUFFICIENT"
		if unknown || len(value.EvidenceCapabilities) == 0 {
			state, reason = UnknownState, "EVIDENCE_UNKNOWN"
		}
		return routeResult{Resolution: Lower, ObservationState: state, Reason: reason}
	}
	evidence := candidates[0]
	return routeResult{
		Route: evidence.Route, Resolution: Exact, ObservationState: "OBSERVED",
		Reason: "EVIDENCE_DERIVED_ROUTE", Observations: evidence.ObservationIDs,
		EvidenceDigest: evidence.EvidenceDigest, Provenance: evidence.Provenance,
	}
}
