package proofchoicejudge

type routeResult struct {
	Route            string
	Resolution       string
	ObservationState string
	Reason           string
	Observations     []string
	EvidenceDigest   string
	Provenance       []string
	FailClosed       bool
}

func selectRoute(subject value, bundle evidenceBundle) routeResult {
	candidates := []evidence{}
	unknown := false
	for _, current := range evidenceFor(subject, bundle) {
		if current.State == "OBSERVED" {
			candidates = append(candidates, current)
		}
		if current.State == "UNKNOWN" {
			unknown = true
		}
	}
	if len(candidates) > 1 {
		return routeResult{Resolution: "EXACT", Reason: "PROOF_ROUTE_CONTRADICTION", FailClosed: true}
	}
	if len(candidates) == 0 {
		state, reason := "INSUFFICIENT", "EVIDENCE_INSUFFICIENT"
		if unknown || len(subject.EvidenceCapabilities) == 0 {
			state, reason = "UNKNOWN", "EVIDENCE_UNKNOWN"
		}
		return routeResult{Resolution: "LOWER_RESOLUTION", ObservationState: state, Reason: reason}
	}
	current := candidates[0]
	return routeResult{Route: current.Route, Resolution: "EXACT", ObservationState: "OBSERVED", Reason: "EVIDENCE_DERIVED_ROUTE", Observations: current.ObservationIDs, EvidenceDigest: current.EvidenceDigest, Provenance: current.Provenance}
}

func validRoute(route string) bool {
	return route == "FOUNDATION" || route == "COHERENCE" || route == "REGRESSION"
}
