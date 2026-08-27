package operationprovenance

func graphSummary(f fixture) GraphSummary {
	kinds := map[string]int{}
	for _, edge := range f.Edges {
		kinds[edge.Kind]++
	}
	return GraphSummary{Nodes: len(f.Nodes), Edges: len(f.Edges), EdgeKinds: kinds}
}

func resolution(exact bool) string {
	if exact {
		return "EXACT"
	}
	return "LOWER_RESOLUTION"
}

func sourceResolution(issues []Issue) string {
	if len(issues) == 0 {
		return "EXACT"
	}
	return "LOWER_RESOLUTION"
}

func scenarioDecision(decisions map[string]int) string {
	if decisions["FAIL_CLOSED"] > 0 {
		return "FAIL_CLOSED"
	}
	if decisions["UNKNOWN"] > 0 {
		return "UNKNOWN"
	}
	return "PASS"
}

func dependencyIssue(reason, blocked string) *Issue {
	return &Issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: reason, Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{blocked}}
}
