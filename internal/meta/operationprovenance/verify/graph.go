package verify

import (
	"fmt"
	"strings"
)

func makeFixture(scenario cScenario, metrics []cMetric, artifacts map[string][]relationObservation) (cFixture, error) {
	f := cFixture{id: scenario.id, mutation: mutationDescription(scenario), nodes: map[string]string{}, metrics: append([]cMetric(nil), metrics...), artifacts: cloneArtifacts(artifacts)}
	for _, metric := range metrics {
		f.nodes["metric:"+metric.id] = "metric"
		for _, observation := range f.artifacts[metric.id] {
			if observation.RelationStatus != "PASS" {
				continue
			}
			link := observedLink(metric, observation)
			f.nodes[link.from] = observation.Relation
			f.nodes[link.to] = "metric"
			f.edges = append(f.edges, cEdge{from: link.from, to: link.to, kind: link.kind})
		}
	}
	if scenario.removeRelation != "" {
		parts := splitMutation(scenario.removeRelation)
		if len(parts) != 2 {
			return cFixture{}, fmt.Errorf("scenario %s has malformed relation mutation", scenario.id)
		}
		wanted := relationForID(metrics, parts[0], parts[1])
		if wanted.kind == "" || !hasEdge(f.edges, wanted) {
			return cFixture{}, fmt.Errorf("scenario %s relation mutation is a no-op", scenario.id)
		}
		f.edges = removeEdge(f.edges, wanted)
		markRemoved(f.artifacts[parts[1]], parts[0], scenario)
	}
	if scenario.dependency != "" {
		parts := strings.SplitN(scenario.dependency, ">", 2)
		if len(parts) != 2 || parts[0] == parts[1] || !hasMetric(metrics, parts[0]) || !hasMetric(metrics, parts[1]) || hasDependency(f.metrics, parts[1], parts[0]) {
			return cFixture{}, fmt.Errorf("scenario %s dependency mutation is invalid or a no-op", scenario.id)
		}
		for index := range f.metrics {
			if f.metrics[index].id == parts[1] {
				f.metrics[index].dependsOn = append(f.metrics[index].dependsOn, parts[0])
			}
		}
	}
	return f, nil
}
