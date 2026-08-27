package proofchoicealgebra

import "strings"

func selectRoute(value Value, observations map[string]Value) routeResult {
	if len(value.Observations) == 0 || len(value.AdmissibleRoutes) == 0 {
		return routeResult{Resolution: Unknown, Reason: "EVIDENCE_UNKNOWN"}
	}
	for _, route := range value.AdmissibleRoutes {
		if !route.Valid() {
			return routeResult{Resolution: Unknown, Reason: "EVIDENCE_UNKNOWN"}
		}
	}
	for _, id := range value.Observations {
		observation, exists := observations[id]
		if !exists || observation.Kind != ObservationKind || !observation.Observed {
			return routeResult{Resolution: Unknown, Reason: "EVIDENCE_UNKNOWN"}
		}
	}
	matches := make([]Route, 0, len(value.AdmissibleRoutes))
	for _, route := range value.AdmissibleRoutes {
		if routeMatches(route, value.Observations, observations) {
			matches = append(matches, route)
		}
	}
	if len(matches) > 1 {
		return routeResult{Resolution: FailClosed, Reason: "PROOF_ROUTE_CONTRADICTION", Observations: value.Observations}
	}
	if len(matches) == 0 {
		return routeResult{Resolution: Lower, Reason: "EVIDENCE_INSUFFICIENT", Observations: value.Observations}
	}
	return routeResult{Route: matches[0], Resolution: Exact, Reason: "EVIDENCE_DERIVED_ROUTE", Observations: value.Observations}
}

func routeMatches(route Route, ids []string, observations map[string]Value) bool {
	for _, id := range ids {
		observation := observations[id]
		if !strings.EqualFold(observation.EvidenceKind, string(route)) {
			return false
		}
		predicate, expected := routeSignal(route)
		if !strings.EqualFold(observation.Predicate, predicate) || !strings.EqualFold(observation.Value, expected) {
			return false
		}
	}
	return true
}

func routeSignal(route Route) (string, string) {
	switch route {
	case Foundation:
		return "identity", "stable"
	case Coherence:
		return "relations", "agree"
	case Regression:
		return "replay", "equal"
	default:
		return "", ""
	}
}
