package proofchoicejudge

func baseSummary() summary {
	return summary{FixedDenominator: 3, ChoiceCounts: map[string]int{"FOUNDATION": 0, "COHERENCE": 0, "REGRESSION": 0}}
}

func summarize(items []item, transitions []transition, observations int) summary {
	result := baseSummary()
	result.Items, result.Transitions, result.Observations = len(items), len(transitions), observations
	for _, item := range items {
		switch item.Kind {
		case "claim":
			result.Claims++
		case "metric":
			result.Metrics++
			result.MetricSlotNumerator += item.Numerator
			result.MetricSlotDenominator += item.Denominator
		}
		if item.Resolution == "EXACT" {
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
