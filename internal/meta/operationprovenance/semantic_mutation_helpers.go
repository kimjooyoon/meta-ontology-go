package operationprovenance

func hasMetric(metrics []metricSpec, id string) bool {
	for _, metric := range metrics {
		if metric.ID == id {
			return true
		}
	}
	return false
}

func hasDependency(metrics []metricSpec, target, upstream string) bool {
	for _, metric := range metrics {
		if metric.ID == target {
			for _, dependency := range metric.DependsOn {
				if dependency == upstream {
					return true
				}
			}
		}
	}
	return false
}

func markRemoved(observations []RelationObservation, kind string, scenario scenarioSpec) {
	for index := range observations {
		if observations[index].Relation == kind && observations[index].RelationStatus == "PASS" {
			observations[index].RelationStatus = "REMOVED"
			observations[index].Stage = "SCENARIO"
			observations[index].Step = "remove-relation"
			observations[index].Reason = scenario.Reason
			observations[index].Cause = "SCENARIO_MUTATION"
		}
	}
}
