package verify

func mutationDescription(scenario cScenario) string {
	return "remove_relation=" + scenario.removeRelation + ";dependency=" + scenario.dependency + ";reason=" + scenario.reason
}

func hasMetric(metrics []cMetric, id string) bool {
	for _, metric := range metrics {
		if metric.id == id {
			return true
		}
	}
	return false
}

func hasDependency(metrics []cMetric, target, upstream string) bool {
	for _, metric := range metrics {
		if metric.id == target {
			for _, dependency := range metric.dependsOn {
				if dependency == upstream {
					return true
				}
			}
		}
	}
	return false
}

func hasEdge(edges []cEdge, wanted cRelation) bool {
	for _, edge := range edges {
		if edge.kind == wanted.kind && edge.from == wanted.from && edge.to == wanted.to {
			return true
		}
	}
	return false
}

func removeEdge(edges []cEdge, wanted cRelation) []cEdge {
	result := make([]cEdge, 0, len(edges)-1)
	for _, edge := range edges {
		if edge.kind == wanted.kind && edge.from == wanted.from && edge.to == wanted.to {
			continue
		}
		result = append(result, edge)
	}
	return result
}

func markRemoved(observations []relationObservation, kind string, scenario cScenario) {
	for index := range observations {
		if observations[index].Relation == kind && observations[index].RelationStatus == "PASS" {
			observations[index].RelationStatus, observations[index].Stage, observations[index].Step = "REMOVED", "SCENARIO", "remove-relation"
			observations[index].Reason, observations[index].Cause = scenario.reason, "SCENARIO_MUTATION"
		}
	}
}
