package operationprovenance

func cloneObservations(input map[string][]RelationObservation) map[string][]RelationObservation {
	output := make(map[string][]RelationObservation, len(input))
	for id, observations := range input {
		output[id] = append([]RelationObservation(nil), observations...)
	}
	return output
}

func mutationDescription(scenario scenarioSpec) string {
	return "remove_relation=" + scenario.RemoveRelation + ";dependency=" + scenario.Dependency + ";reason=" + scenario.Reason
}

func hasEdge(edges []edge, wanted relation) bool {
	for _, edge := range edges {
		if edge.Kind == wanted.Kind && edge.From == wanted.From && edge.To == wanted.To {
			return true
		}
	}
	return false
}

func removeEdge(edges []edge, wanted relation) []edge {
	filtered := make([]edge, 0, len(edges)-1)
	for _, edge := range edges {
		if edge.Kind == wanted.Kind && edge.From == wanted.From && edge.To == wanted.To {
			continue
		}
		filtered = append(filtered, edge)
	}
	return filtered
}
