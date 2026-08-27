package proofchoicealgebra

func composeValue(value Value, results map[string]routeResult) (Composition, routeResult) {
	seen := map[Route]bool{}
	routes, provenance, digests := []Route{}, []string{}, []string{}
	for _, member := range value.Members {
		result, exists := results[member]
		if !exists || result.Resolution != Exact || result.FailClosed || !result.Route.Valid() || seen[result.Route] {
			return failedComposition(value, routes)
		}
		seen[result.Route] = true
		routes = append(routes, result.Route)
		provenance = append(provenance, result.Provenance...)
		digests = append(digests, result.EvidenceDigest)
	}
	if len(routes) != 3 || !seen[Foundation] || !seen[Coherence] || !seen[Regression] {
		return failedComposition(value, routes)
	}
	digest := digestBytes([]byte(join(digests)))
	provenance = uniqueSorted(provenance)
	composition := Composition{ID: value.ID, Statement: value.Statement, Members: value.Members, Routes: routes, Operator: "ALL_ROUTES", Result: "DISCHARGED", EvidenceDigest: digest, Provenance: provenance}
	return composition, routeResult{Resolution: Exact, Reason: "PROOF_COMPOSITION_DISCHARGED", EvidenceDigest: digest, Provenance: provenance}
}

func failedComposition(value Value, routes []Route) (Composition, routeResult) {
	composition := Composition{ID: value.ID, Statement: value.Statement, Members: value.Members, Routes: routes, Operator: "ALL_ROUTES", Result: "REFUTED"}
	return composition, routeResult{Resolution: Exact, Reason: "PROOF_COMPOSITION_FAILED", FailClosed: true}
}

func join(values []string) string {
	result := ""
	for _, value := range values {
		result += value + "\n"
	}
	return result
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
