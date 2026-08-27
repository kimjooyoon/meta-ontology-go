package operationprovenance

func evaluateFixture(f fixture, sourceDigest, semanticDigest string) ScenarioResult {
	edgeCounts, edgeKinds := countEdges(f)
	byID := make(map[string]metricSpec, len(f.Metrics))
	for _, metric := range f.Metrics {
		byID[metric.ID] = metric
	}
	memo := make(map[string]MetricResult)
	visiting := make(map[string]bool)
	var evaluate func(string) MetricResult
	evaluate = func(id string) MetricResult {
		if result, ok := memo[id]; ok {
			return result
		}
		metric := byID[id]
		if visiting[id] {
			return metricResult(metric, edgeCounts, "UNKNOWN", dependencyIssue("DEPENDENCY_CYCLE", id), sourceDigest, semanticDigest, f)
		}
		visiting[id] = true
		result := metricResult(metric, edgeCounts, "", nil, sourceDigest, semanticDigest, f)
		for _, dependency := range metric.DependsOn {
			upstream, ok := byID[dependency]
			if !ok {
				result = metricResult(metric, edgeCounts, "UNKNOWN", dependencyIssue("UPSTREAM_METRIC_MISSING", dependency), sourceDigest, semanticDigest, f)
				break
			}
			if upstreamResult := evaluate(upstream.ID); upstreamResult.Decision != "PASS" {
				result = metricResult(metric, edgeCounts, "UNKNOWN", dependencyIssue("UPSTREAM_"+upstreamResult.Decision, dependency), sourceDigest, semanticDigest, f)
				break
			}
		}
		visiting[id] = false
		memo[id] = result
		return result
	}
	results := make([]MetricResult, 0, len(f.Metrics))
	decisions, transitions := map[string]int{}, map[string]int{}
	numerator := 0
	for _, metric := range f.Metrics {
		result := evaluate(metric.ID)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
	}
	decision := scenarioDecision(decisions)
	return ScenarioResult{ID: f.ID, Mutation: f.Mutation, Graph: GraphSummary{Nodes: len(f.Nodes), Edges: len(f.Edges), EdgeKinds: edgeKinds}, Numerator: numerator, Denominator: len(results) * relationDenominator, ConformanceDecision: decision, SubjectResolution: "EXACT", Decisions: decisions, Transitions: transitions, Metrics: results}
}

func countEdges(f fixture) (map[string]int, map[string]int) {
	counts, kinds := map[string]int{}, map[string]int{}
	for _, edge := range f.Edges {
		if f.Nodes[edge.From] == "" || f.Nodes[edge.To] == "" {
			continue
		}
		counts[edge.From+"\x00"+edge.To+"\x00"+edge.Kind]++
		kinds[edge.Kind]++
	}
	return counts, kinds
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
