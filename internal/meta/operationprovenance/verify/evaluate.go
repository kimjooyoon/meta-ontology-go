package verify

func evaluate(scenario cScenario, metrics []cMetric, artifacts map[string][]relationObservation, sourceDigest, semanticDigest string) (scenarioResult, error) {
	f, err := makeFixture(scenario, metrics, artifacts)
	if err != nil {
		return scenarioResult{}, err
	}
	evaluator := newEvaluator(f, sourceDigest, semanticDigest)
	results := make([]metricResult, 0, len(f.metrics))
	decisions, transitions := map[string]int{}, map[string]int{}
	numerator := 0
	lineageExact := true
	for _, metric := range f.metrics {
		result := evaluator.run(metric.id)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
		lineageExact = lineageExact && result.LineageResolution == "EXACT"
	}
	return scenarioResult{ID: f.id, Mutation: f.mutation, Graph: buildGraphSummary(f), Numerator: numerator, Denominator: len(results) * 4, Decision: scenarioDecision(decisions), SourceResolution: "EXACT", LineageResolution: resolution(lineageExact), Decisions: decisions, Transitions: transitions, Metrics: results}, nil
}

func buildGraphSummary(f cFixture) graphSummary {
	kinds := map[string]int{}
	for _, edge := range f.edges {
		kinds[edge.kind]++
	}
	return graphSummary{Nodes: len(f.nodes), Edges: len(f.edges), EdgeKinds: kinds}
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

func resolution(exact bool) string {
	if exact {
		return "EXACT"
	}
	return "LOWER_RESOLUTION"
}
