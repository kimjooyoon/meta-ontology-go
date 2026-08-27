package proofchoicealgebra

func baseSummary() Summary {
	return Summary{FixedDenominator: FixedDenom, RouteDenominator: 3, CaseDenominator: 1, ChoiceCounts: routeCounts()}
}

func routeCounts() map[Route]int {
	return map[Route]int{Foundation: 0, Coherence: 0, Regression: 0}
}

func summarize(items []Item, transitions []Transition, bundle evidenceBundle) Summary {
	result := baseSummary()
	result.Items, result.Transitions = len(items), len(transitions)
	for _, evidence := range bundle.All {
		result.Observations += len(evidence.ObservationSlots)
	}
	for _, item := range items {
		switch item.Kind {
		case ClaimKind:
			result.Claims++
		case MetricKind:
			result.Metrics++
			result.MetricSlotNumerator += item.Numerator
			result.MetricSlotDenominator += item.Denominator
		}
		if item.Resolution == Exact {
			result.ExactChoices++
			route := Route(item.Choice)
			if route.Valid() {
				result.ChoiceCounts[route]++
			}
		}
	}
	if result.Items > 0 {
		result.ChoiceCoverageBPS = result.ExactChoices * 10000 / result.Items
	}
	result.ClaimDenominator = result.Claims
	return result
}
