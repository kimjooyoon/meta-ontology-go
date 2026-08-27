package proofchoicealgebra

func indicators(bundle Bundle) []Indicator {
	result := make([]Indicator, 0, len(bundle.Items)+len(bundle.Transitions))
	for _, item := range bundle.Items {
		decision := Pass
		if !item.Choice.Valid() {
			decision = FailClosed
		}
		result = append(result, Indicator{ID: item.ID, Kind: string(item.Kind), Choice: item.Choice, Decision: decision, Relation: "choice", Value: item.Choice.String(), Limit: "exactly-one"})
	}
	for _, transition := range bundle.Transitions {
		decision := Pass
		if !transition.Choice.Valid() {
			decision = FailClosed
		}
		result = append(result, Indicator{ID: "transition:" + transition.ClaimID, Kind: "PERSISTENT_TRANSITION", Choice: transition.Choice, Decision: decision, Relation: "preserves", Value: transition.To, Limit: transition.From})
	}
	return result
}

func countContradictions(bundle Bundle) int {
	seen := map[string]Choice{}
	count := 0
	for _, item := range bundle.Items {
		if choice, exists := seen[item.ID]; exists && choice != item.Choice {
			count++
		}
		seen[item.ID] = item.Choice
	}
	return count
}
