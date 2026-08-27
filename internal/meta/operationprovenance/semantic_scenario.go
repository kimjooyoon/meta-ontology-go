package operationprovenance

import (
	"fmt"
	"strings"
)

func evaluateScenario(metrics []metricSpec, scenario scenarioSpec, sourceDigest, semanticDigest string) (ScenarioResult, error) {
	working := fixtureFromMetrics(metrics, scenario)
	if scenario.RemoveRelation != "" {
		parts := strings.SplitN(scenario.RemoveRelation, ":", 2)
		if len(parts) != 2 {
			return ScenarioResult{}, fmt.Errorf("scenario %s has malformed relation mutation", scenario.ID)
		}
		working.Edges = removeRelation(working.Edges, metrics, parts[0], parts[1])
	}
	if scenario.Dependency != "" {
		if !strings.Contains(scenario.Dependency, ">") {
			return ScenarioResult{}, fmt.Errorf("scenario %s has malformed dependency mutation", scenario.ID)
		}
		dependencyMutation(working.Metrics, scenario.Dependency)
	}
	return evaluateFixture(working, sourceDigest, semanticDigest), nil
}
