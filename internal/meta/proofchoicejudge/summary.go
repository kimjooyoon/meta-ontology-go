package proofchoicejudge

func baseSummary() summary {
	return summary{FixedDenominator: 3, RouteDenominator: 3, CaseDenominator: 1, ChoiceCounts: routeCounts()}
}

func routeCounts() map[string]int {
	return map[string]int{"FOUNDATION": 0, "COHERENCE": 0, "REGRESSION": 0}
}

func summarize(items []item, transitions []transition, bundle evidenceBundle) summary {
	result := baseSummary()
	result.Items, result.Transitions = len(items), len(transitions)
	for _, current := range bundle.All {
		result.Observations += len(current.ObservationSlots)
	}
	for _, current := range items {
		switch current.Kind {
		case "claim":
			result.Claims++
		case "metric":
			result.Metrics++
			result.MetricSlotNumerator += current.Numerator
			result.MetricSlotDenominator += current.Denominator
		}
		if current.Resolution == "EXACT" {
			result.ExactChoices++
			if validRoute(current.Choice) {
				result.ChoiceCounts[current.Choice]++
			}
		}
	}
	if result.Items > 0 {
		result.ChoiceCoverageBPS = result.ExactChoices * 10000 / result.Items
	}
	result.ClaimDenominator = result.Claims
	return result
}
