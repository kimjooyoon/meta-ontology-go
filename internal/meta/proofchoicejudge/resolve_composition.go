package proofchoicejudge

func composeValue(current value, results map[string]routeResult) (composition, routeResult) {
	seen := map[string]bool{}
	routes := []string{}
	digests, provenance := []string{}, []string{}
	for _, member := range current.Members {
		route, exists := results[member]
		if !exists || route.Resolution != "EXACT" || route.FailClosed || !validRoute(route.Route) || seen[route.Route] {
			return failedComposition(current, routes)
		}
		seen[route.Route] = true
		routes = append(routes, route.Route)
		digests = append(digests, route.EvidenceDigest)
		provenance = append(provenance, route.Provenance...)
	}
	if len(routes) != 3 || !seen["FOUNDATION"] || !seen["COHERENCE"] || !seen["REGRESSION"] {
		return failedComposition(current, routes)
	}
	digest := digestBytes([]byte(join(digests)))
	provenance = uniqueSorted(provenance)
	composition := composition{ID: current.ID, Statement: current.Statement, Members: current.Members, Routes: routes, Operator: "ALL_ROUTES", Result: "DISCHARGED", EvidenceDigest: digest, Provenance: provenance}
	route := routeResult{Resolution: "EXACT", Reason: "PROOF_COMPOSITION_DISCHARGED", EvidenceDigest: digest, Provenance: provenance}
	return composition, route
}

func failedComposition(current value, routes []string) (composition, routeResult) {
	composition := composition{ID: current.ID, Statement: current.Statement, Members: current.Members, Routes: routes, Operator: "ALL_ROUTES", Result: "REFUTED"}
	return composition, routeResult{Resolution: "EXACT", Reason: "PROOF_COMPOSITION_FAILED", FailClosed: true}
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
