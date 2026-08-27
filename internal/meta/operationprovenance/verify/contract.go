package verify

import (
	"fmt"
	"strings"
)

func validateContract(metrics []cMetric, scenarios []cScenario) (map[string]int, error) {
	counts, ids := map[string]int{}, map[string]bool{}
	for _, metric := range metrics {
		if ids[metric.id] {
			return nil, fmt.Errorf("consumer duplicate metric id %s", metric.id)
		}
		ids[metric.id] = true
		counts[metric.family]++
	}
	for _, family := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		if counts[family] != 1 {
			return nil, fmt.Errorf("consumer family %s cardinality is %d", family, counts[family])
		}
	}
	if len(metrics) != 3 || len(scenarios) != 4 {
		return nil, fmt.Errorf("consumer metric/scenario cardinality is not 3/4")
	}
	scenarioIDs := map[string]bool{}
	base, removals, dependencies := 0, 0, 0
	for _, scenario := range scenarios {
		if scenario.id == "" {
			return nil, fmt.Errorf("consumer scenario id is empty")
		}
		if scenarioIDs[scenario.id] {
			return nil, fmt.Errorf("consumer duplicate scenario id %s", scenario.id)
		}
		scenarioIDs[scenario.id] = true
		if scenario.removeRelation == "" && scenario.dependency == "" {
			base++
		}
		if scenario.removeRelation != "" {
			removals++
		}
		if scenario.dependency != "" {
			dependencies++
		}
	}
	if base != 1 || removals != 3 || dependencies != 1 {
		return nil, fmt.Errorf("consumer scenario mutation cardinality is not 1/3/1")
	}
	return counts, validateTargets(metrics, scenarios)
}

func validateTargets(metrics []cMetric, scenarios []cScenario) error {
	ids := map[string]bool{}
	for _, metric := range metrics {
		ids[metric.id] = true
	}
	for _, scenario := range scenarios {
		if scenario.removeRelation != "" {
			parts := splitMutation(scenario.removeRelation)
			if len(parts) != 2 || !ids[parts[1]] || !validRelationKind(parts[0]) {
				return fmt.Errorf("consumer mutation targets unknown metric")
			}
		}
		if scenario.dependency != "" {
			parts := strings.SplitN(scenario.dependency, ">", 2)
			if len(parts) != 2 || parts[0] == parts[1] || !ids[parts[0]] || !ids[parts[1]] {
				return fmt.Errorf("consumer dependency targets unknown metric")
			}
		}
	}
	return nil
}
