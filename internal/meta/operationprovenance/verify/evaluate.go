package verify

import "strings"

func evaluate(scenario cScenario, metrics []cMetric, sourceDigest, semanticDigest string) scenarioResult {
	f := makeFixture(scenario, metrics)
	if scenario.removeRelation != "" {
		parts := strings.SplitN(scenario.removeRelation, ":", 2)
		if len(parts) != 2 {
			return scenarioResult{ID: scenario.id, Mutation: mutationDescription(scenario), Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION"}
		}
		f.edges = removeEdge(f.edges, metrics, parts[0], parts[1])
	}
	if scenario.dependency != "" {
		if !strings.Contains(scenario.dependency, ">") {
			return scenarioResult{ID: scenario.id, Mutation: mutationDescription(scenario), Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION"}
		}
		applyDependency(f.metrics, scenario.dependency)
	}
	edgeCounts, edgeKinds := countEdges(f)
	evaluator := newEvaluator(f, edgeCounts, sourceDigest, semanticDigest)
	results := make([]metricResult, 0, len(f.metrics))
	decisions, transitions := map[string]int{}, map[string]int{}
	numerator := 0
	for _, metric := range f.metrics {
		result := evaluator.run(metric.id)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
	}
	return scenarioResult{ID: f.id, Mutation: f.mutation, Graph: graphSummary{Nodes: len(f.nodes), Edges: len(f.edges), EdgeKinds: edgeKinds}, Numerator: numerator, Denominator: len(results) * 4, Decision: scenarioDecision(decisions), Resolution: "EXACT", Decisions: decisions, Transitions: transitions, Metrics: results}
}

func countEdges(f cFixture) (map[string]int, map[string]int) {
	counts, kinds := map[string]int{}, map[string]int{}
	for _, edge := range f.edges {
		if f.nodes[edge.from] == "" || f.nodes[edge.to] == "" {
			continue
		}
		counts[edge.from+"\x00"+edge.to+"\x00"+edge.kind]++
		kinds[edge.kind]++
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
