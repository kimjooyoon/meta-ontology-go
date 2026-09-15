package valueexecution

func counterexampleCount(results []CounterexampleResult) int {
	passed := 0
	for _, result := range results {
		if result.Passed && result.ReplayMatched {
			passed++
		}
	}
	return passed
}

func counterexamplePassed(results []CounterexampleResult, id string) bool {
	for _, result := range results {
		if result.ID == id {
			return result.Passed && result.ReplayMatched
		}
	}
	return false
}
