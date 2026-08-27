package operationprovenance

import (
	"fmt"
	"strings"
)

func fixtureFromArtifacts(metrics []metricSpec, scenario scenarioSpec, observations map[string][]RelationObservation) (fixture, error) {
	working := fixture{ID: scenario.ID, Mutation: mutationDescription(scenario), Nodes: map[string]string{}, Metrics: append([]metricSpec(nil), metrics...), Artifacts: cloneObservations(observations)}
	for _, metric := range metrics {
		working.Nodes["metric:"+metric.ID] = "metric"
		for _, observation := range observations[metric.ID] {
			if observation.RelationStatus != "PASS" {
				continue
			}
			link := observedRelation(metric, observation)
			working.Nodes[link.From] = observation.Relation
			working.Nodes[link.To] = "metric"
			working.Edges = append(working.Edges, edge{From: link.From, To: link.To, Kind: link.Kind})
		}
	}
	if err := applyRelationMutation(&working, scenario); err != nil {
		return fixture{}, err
	}
	if err := applyDependencyMutation(&working, scenario); err != nil {
		return fixture{}, err
	}
	return working, nil
}

func applyRelationMutation(working *fixture, scenario scenarioSpec) error {
	if scenario.RemoveRelation == "" {
		return nil
	}
	parts := splitMutation(scenario.RemoveRelation)
	if len(parts) != 2 {
		return fmt.Errorf("scenario %s has malformed relation mutation", scenario.ID)
	}
	wanted := relationForID(working.Metrics, parts[0], parts[1])
	if wanted.Kind == "" || !hasEdge(working.Edges, wanted) {
		return fmt.Errorf("scenario %s relation mutation is a no-op", scenario.ID)
	}
	working.Edges = removeEdge(working.Edges, wanted)
	markRemoved(working.Artifacts[parts[1]], parts[0], scenario)
	return nil
}

func applyDependencyMutation(working *fixture, scenario scenarioSpec) error {
	if scenario.Dependency == "" {
		return nil
	}
	parts := strings.SplitN(scenario.Dependency, ">", 2)
	if len(parts) != 2 || parts[0] == parts[1] || !hasMetric(working.Metrics, parts[0]) || !hasMetric(working.Metrics, parts[1]) {
		return fmt.Errorf("scenario %s dependency mutation is invalid", scenario.ID)
	}
	if hasDependency(working.Metrics, parts[1], parts[0]) {
		return fmt.Errorf("scenario %s dependency mutation is a no-op", scenario.ID)
	}
	for index := range working.Metrics {
		if working.Metrics[index].ID == parts[1] {
			working.Metrics[index].DependsOn = append(working.Metrics[index].DependsOn, parts[0])
		}
	}
	return nil
}
