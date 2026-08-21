package closure_test

func fixtureOperations() []map[string]any {
	activities := []string{"BindExactSourceMetrics", "ExemptProjectRootTopology",
		"InterpretDimensionRegistry", "ProjectAlgebraicRootState",
		"ObserveCounterfactualBoundary", "PreserveRepositoryWorkspace",
		"ReplayCounterfactual", "TerminateAtFixedPoint"}
	values := make([]map[string]any, len(activities))
	for index, activity := range activities {
		values[index] = map[string]any{
			"id": "operation-" + string(rune('a'+index)), "activity": activity,
			"repository_writes": false, "promotion_authorized": false,
		}
	}
	return values
}

func fixtureBindings() []map[string]any {
	values := make([]map[string]any, 15)
	for index := range values {
		values[index] = map[string]any{
			"indicator_id": "indicator-" + string(rune('a'+index)),
			"operation_id": "operation-" + string(rune('a'+index%8)),
		}
	}
	return values
}

func fixtureSteps() []map[string]any {
	activities := []string{"ObserveCounterfactualBoundary", "PreserveRepositoryWorkspace",
		"ReplayCounterfactual", "TerminateAtFixedPoint"}
	values := make([]map[string]any, len(activities))
	for index, activity := range activities {
		values[index] = map[string]any{"activity": activity}
	}
	return values
}
