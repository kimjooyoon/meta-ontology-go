package proofchoicejudge

import "strings"

type routeResult struct {
	Route        string
	Resolution   string
	Reason       string
	Observations []string
}

func selectRoute(subject value, observations map[string]value) routeResult {
	if len(subject.Observations) == 0 || len(subject.AdmissibleRoutes) == 0 {
		return routeResult{Resolution: "UNKNOWN", Reason: "EVIDENCE_UNKNOWN"}
	}
	for _, route := range subject.AdmissibleRoutes {
		if !validRoute(route) {
			return routeResult{Resolution: "UNKNOWN", Reason: "EVIDENCE_UNKNOWN"}
		}
	}
	for _, id := range subject.Observations {
		item, exists := observations[id]
		if !exists || item.Kind != "observation" || !item.Observed {
			return routeResult{Resolution: "UNKNOWN", Reason: "EVIDENCE_UNKNOWN"}
		}
	}
	matches := []string{}
	for _, route := range subject.AdmissibleRoutes {
		if routeMatches(route, subject.Observations, observations) {
			matches = append(matches, route)
		}
	}
	if len(matches) > 1 {
		return routeResult{Resolution: "FAIL_CLOSED", Reason: "PROOF_ROUTE_CONTRADICTION", Observations: subject.Observations}
	}
	if len(matches) == 0 {
		return routeResult{Resolution: "LOWER_RESOLUTION", Reason: "EVIDENCE_INSUFFICIENT", Observations: subject.Observations}
	}
	return routeResult{Route: matches[0], Resolution: "EXACT", Reason: "EVIDENCE_DERIVED_ROUTE", Observations: subject.Observations}
}

func validRoute(route string) bool {
	return route == "FOUNDATION" || route == "COHERENCE" || route == "REGRESSION"
}

func routeMatches(route string, ids []string, observations map[string]value) bool {
	predicate, expected := routeSignal(route)
	for _, id := range ids {
		item := observations[id]
		if !strings.EqualFold(item.EvidenceKind, route) || !strings.EqualFold(item.Predicate, predicate) || !strings.EqualFold(item.Value, expected) {
			return false
		}
	}
	return true
}

func routeSignal(route string) (string, string) {
	switch route {
	case "FOUNDATION":
		return "identity", "stable"
	case "COHERENCE":
		return "relations", "agree"
	case "REGRESSION":
		return "replay", "equal"
	default:
		return "", ""
	}
}
