package proofchoicealgebra

func summarize(items []Item, transitions []Transition, observations int) Summary {
	result := Summary{Items: len(items), Transitions: len(transitions), FixedDenominator: FixedDenom, ChoiceCounts: map[Route]int{Foundation: 0, Coherence: 0, Regression: 0}}
	result.Observations = observations
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
			result.ChoiceCounts[item.Choice]++
		}
	}
	if result.Items > 0 {
		result.ChoiceCoverageBPS = result.ExactChoices * 10000 / result.Items
	}
	for _, transition := range transitions {
		switch transition.To {
		case "DISCHARGED":
			result.Discharged++
		case "OPEN":
			result.OpenPreserved++
		case "REFUTED":
			result.Refuted++
		}
	}
	return result
}
