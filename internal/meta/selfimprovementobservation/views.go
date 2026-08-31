package selfimprovementobservation

func observationViews(indicators []Indicator) []View {
	userIDs := []string{
		"coherence.minimal-value-state", "coherence.value-witnesses",
		"coherence.resource-witnesses", "regression.counterexample-coverage",
		"regression.no-source-effects",
	}
	toolIDs := make([]string, 0, 12)
	allIDs := make([]string, 0, len(indicators))
	for index, indicator := range indicators {
		allIDs = append(allIDs, indicator.ID)
		if index < 12 {
			toolIDs = append(toolIDs, indicator.ID)
		}
	}
	return []View{
		buildView("USER", "USER_VISIBLE", userIDs, indicators),
		buildView("TOOL_AUTHOR", "TOOL_CONTRACT", toolIDs, indicators),
		buildView("GOVERNOR", "FULL_RECEIPT", allIDs, indicators),
	}
}

func buildView(audience, resolution string, ids []string, indicators []Indicator) View {
	states, satisfied := map[string]bool{}, 0
	for _, indicator := range indicators {
		states[indicator.ID] = indicator.Satisfied
	}
	for _, id := range ids {
		if states[id] {
			satisfied++
		}
	}
	return View{Audience: audience, Resolution: resolution, Satisfied: satisfied, Total: len(ids), BasisPoints: basisPoints(satisfied, len(ids)), IndicatorIDs: ids}
}
