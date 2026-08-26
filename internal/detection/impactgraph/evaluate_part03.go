package impactgraph

import (
	"sort"
)

func reachableObligations(roots []string, adjacency map[string][]string, byID map[string]NodeKind) []string {
	visited := make(map[string]struct{}, len(byID))
	queue := append([]string(nil), roots...)
	required := make(map[string]struct{})
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}
		if byID[current] == NodeKindObligation {
			required[current] = struct{}{}
		}
		queue = append(queue, adjacency[current]...)
	}
	result := make([]string, 0, len(required))
	for id := range required {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
func difference(left, right []string) []string {
	set := make(map[string]struct{}, len(right))
	for _, id := range right {
		set[id] = struct{}{}
	}
	result := make([]string, 0, len(left))
	for _, id := range left {
		if _, exists := set[id]; !exists {
			result = append(result, id)
		}
	}
	return result
}
func unknownEvaluation(code string) Evaluation {
	return Evaluation{
		Required: []string{}, ExecutedRequired: []string{}, Missed: []string{}, Extra: []string{},
		Decision: DecisionUnknown, FailureCode: code, FullSuiteRequired: true,
	}
}
