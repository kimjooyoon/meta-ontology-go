package verify

import "strings"

func makeFixture(scenario cScenario, metrics []cMetric) cFixture {
	f := cFixture{id: scenario.id, mutation: mutationDescription(scenario), nodes: map[string]string{}, metrics: append([]cMetric(nil), metrics...)}
	for _, metric := range metrics {
		f.nodes["metric:"+metric.id] = "metric"
		for _, value := range []struct{ id, kind string }{{metric.producer, "producer"}, {metric.consumer, "consumer"}, {metric.operation, "meta-operation"}, {metric.evidence, "evidence-path"}} {
			if value.id != "" {
				f.nodes[value.id] = value.kind
			}
		}
		for _, link := range links(metric) {
			if link.from != "" && link.to != "" {
				f.edges = append(f.edges, cEdge{from: link.from, to: link.to, kind: link.kind})
			}
		}
	}
	return f
}

func mutationDescription(scenario cScenario) string {
	return "remove_relation=" + scenario.removeRelation + ";dependency=" + scenario.dependency + ";reason=" + scenario.reason
}

func links(metric cMetric) []cRelation {
	return []cRelation{{"PRODUCES", metric.producer, "metric:" + metric.id}, {"CONSUMES", "metric:" + metric.id, metric.consumer}, {"OPERATES", metric.operation, "metric:" + metric.id}, {"EVIDENCED_BY", "metric:" + metric.id, metric.evidence}}
}

func removeEdge(edges []cEdge, metrics []cMetric, kind, metricID string) []cEdge {
	var wanted cRelation
	for _, metric := range metrics {
		if metric.id == metricID {
			for _, link := range links(metric) {
				if link.kind == kind {
					wanted = link
				}
			}
		}
	}
	filtered := make([]cEdge, 0, len(edges))
	for _, edge := range edges {
		if edge.kind == wanted.kind && edge.from == wanted.from && edge.to == wanted.to {
			continue
		}
		filtered = append(filtered, edge)
	}
	return filtered
}

func applyDependency(metrics []cMetric, dependency string) {
	parts := strings.SplitN(dependency, ">", 2)
	if len(parts) != 2 {
		return
	}
	for index := range metrics {
		if metrics[index].id == parts[1] {
			metrics[index].dependsOn = append(metrics[index].dependsOn, parts[0])
		}
	}
}
