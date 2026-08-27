package operationprovenance

import "fmt"

var requiredFamilies = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}

func validateContract(metrics []metricSpec, scenarios []scenarioSpec) (map[string]int, error) {
	counts := map[string]int{}
	ids := map[string]bool{}
	for _, metric := range metrics {
		if ids[metric.ID] {
			return nil, fmt.Errorf("duplicate metric id %s", metric.ID)
		}
		ids[metric.ID] = true
		if metric.PriorClaim != "OPEN" {
			return nil, fmt.Errorf("metric %s prior claim is not OPEN", metric.ID)
		}
		counts[metric.Family]++
	}
	for _, family := range requiredFamilies {
		if counts[family] != 1 {
			return nil, fmt.Errorf("family %s cardinality is %d, want 1", family, counts[family])
		}
	}
	if len(metrics) != len(requiredFamilies) {
		return nil, fmt.Errorf("metric cardinality is %d, want %d", len(metrics), len(requiredFamilies))
	}
	scenarioIDs := map[string]bool{}
	for _, scenario := range scenarios {
		if scenarioIDs[scenario.ID] {
			return nil, fmt.Errorf("duplicate scenario id %s", scenario.ID)
		}
		scenarioIDs[scenario.ID] = true
	}
	if len(scenarios) != 4 || !scenarioIDs["complete"] || !scenarioIDs["disconnected"] || !scenarioIDs["direct-unknown"] || !scenarioIDs["dependency-blocked"] {
		return nil, fmt.Errorf("scenario cardinality or identity is incomplete")
	}
	return counts, validateScenarioTargets(metrics, scenarios)
}

func validateScenarioTargets(metrics []metricSpec, scenarios []scenarioSpec) error {
	ids := map[string]bool{}
	for _, metric := range metrics {
		ids[metric.ID] = true
	}
	for _, scenario := range scenarios {
		if scenario.ID == "complete" && (scenario.RemoveRelation != "" || scenario.Dependency != "") {
			return fmt.Errorf("complete scenario unexpectedly mutates graph")
		}
		if scenario.ID == "disconnected" && scenario.RemoveRelation != "CONSUMES:MOP-COHERENCE-001" {
			return fmt.Errorf("disconnected scenario target is not exact")
		}
		if scenario.ID == "direct-unknown" && scenario.RemoveRelation != "EVIDENCED_BY:MOP-FOUNDATION-001" {
			return fmt.Errorf("direct-unknown scenario target is not exact")
		}
		if scenario.ID == "dependency-blocked" && (scenario.RemoveRelation != "EVIDENCED_BY:MOP-FOUNDATION-001" || scenario.Dependency != "MOP-FOUNDATION-001>MOP-REGRESSION-001") {
			return fmt.Errorf("dependency-blocked scenario target is not exact")
		}
	}
	for _, scenario := range scenarios {
		if scenario.RemoveRelation != "" {
			parts := splitMutation(scenario.RemoveRelation)
			if len(parts) != 2 || !ids[parts[1]] {
				return fmt.Errorf("scenario %s removes an unknown metric", scenario.ID)
			}
		}
	}
	return nil
}
