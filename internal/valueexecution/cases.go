package valueexecution

import "math"

func executeCases(program Program) []CaseResult {
	fixtures := []struct {
		id       string
		input    int64
		expected int64
	}{
		{"negative", -2, -1}, {"negative-to-zero", -1, 0}, {"zero", 0, 1},
		{"positive", 41, 42}, {"maximum-boundary", math.MaxInt64 - 1, math.MaxInt64},
	}
	results := make([]CaseResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		actual, firstErr := program.Execute([]int64{fixture.input})
		replay, secondErr := program.Execute([]int64{fixture.input})
		results = append(results, CaseResult{
			ID: fixture.id, Input: fixture.input, Expected: fixture.expected, Actual: actual, Replay: replay,
			Passed:        firstErr == nil && actual == fixture.expected,
			ReplayMatched: firstErr == nil && secondErr == nil && actual == replay,
		})
	}
	return results
}

func caseCounts(cases []CaseResult) (int, int) {
	passed, replayed := 0, 0
	for _, result := range cases {
		if result.Passed {
			passed++
		}
		if result.ReplayMatched {
			replayed++
		}
	}
	return passed, replayed
}
