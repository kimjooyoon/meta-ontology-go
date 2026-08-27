package audienceresolution

func claimTransitions(indicators []Indicator) []ClaimTransition {
	result := make([]ClaimTransition, 0, len(indicators))
	for _, indicator := range indicators {
		result = append(result, ClaimTransition{IndicatorID: indicator.ID, Before: indicator.ClaimBefore,
			After: indicator.ClaimAfter, Reason: indicator.Reason})
	}
	return result
}

func blockedCounterexamples(values []Counterexample) int {
	count := 0
	for _, value := range values {
		if value.Blocked {
			count++
		}
	}
	return count
}

func counterexamplesValid(values []Counterexample) bool {
	return len(values) == 2 && counterexampleValid(values, "counterexample.missing-information") &&
		counterexampleValid(values, "counterexample.decision-contradiction")
}
