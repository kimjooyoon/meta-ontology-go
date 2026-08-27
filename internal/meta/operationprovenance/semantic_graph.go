package operationprovenance

import "strings"

func fixtureFromMetrics(metrics []metricSpec, scenario scenarioSpec) fixture {
	working := fixture{ID: scenario.ID, Mutation: mutationDescription(scenario), Nodes: make(map[string]string), Metrics: append([]metricSpec(nil), metrics...)}
	for _, metric := range metrics {
		working.Nodes["metric:"+metric.ID] = "metric"
		for _, value := range []struct{ id, kind string }{{metric.Producer, "producer"}, {metric.Consumer, "consumer"}, {metric.MetaOperation, "meta-operation"}, {metric.EvidencePath, "evidence-path"}} {
			if value.id != "" {
				working.Nodes[value.id] = value.kind
			}
		}
		for _, link := range relations(metric) {
			if link.From != "" && link.To != "" {
				working.Edges = append(working.Edges, edge{From: link.From, To: link.To, Kind: link.Kind})
			}
		}
	}
	return working
}

func mutationDescription(scenario scenarioSpec) string {
	return "remove_relation=" + scenario.RemoveRelation + ";dependency=" + scenario.Dependency + ";reason=" + scenario.Reason
}

func relations(metric metricSpec) []relation {
	return []relation{{"PRODUCES", metric.Producer, "metric:" + metric.ID}, {"CONSUMES", "metric:" + metric.ID, metric.Consumer}, {"OPERATES", metric.MetaOperation, "metric:" + metric.ID}, {"EVIDENCED_BY", "metric:" + metric.ID, metric.EvidencePath}}
}

func removeRelation(edges []edge, metrics []metricSpec, kind, metricID string) []edge {
	var wanted relation
	for _, metric := range metrics {
		if metric.ID != metricID {
			continue
		}
		for _, link := range relations(metric) {
			if link.Kind == kind {
				wanted = link
			}
		}
	}
	filtered := make([]edge, 0, len(edges))
	for _, current := range edges {
		if current.Kind == wanted.Kind && current.From == wanted.From && current.To == wanted.To {
			continue
		}
		filtered = append(filtered, current)
	}
	return filtered
}

func dependencyMutation(metrics []metricSpec, dependency string) {
	parts := strings.SplitN(dependency, ">", 2)
	if len(parts) != 2 {
		return
	}
	for index := range metrics {
		if metrics[index].ID == parts[1] {
			metrics[index].DependsOn = append(metrics[index].DependsOn, parts[0])
		}
	}
}
