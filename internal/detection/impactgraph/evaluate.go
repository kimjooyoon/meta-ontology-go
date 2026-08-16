package impactgraph

import "sort"

// Evaluate computes obligation coverage from directed reachability. The graph
// itself is the only source of semantic meaning; no names or paths are used.
func (graph Graph) Evaluate(changedIDs, executedObligationIDs []string) Evaluation {
	normalized, err := graph.Normalized()
	if err != nil {
		return unknownEvaluation(FailureCodeInvalidRegistry)
	}
	byID := make(map[string]NodeKind, len(normalized.Nodes))
	for _, node := range normalized.Nodes {
		byID[node.ID] = node.Kind
	}
	roots, code := inputIDs(changedIDs, byID, false)
	if code != FailureCodeNone {
		return unknownEvaluation(code)
	}
	executed, code := executedIDs(executedObligationIDs, byID)
	if code != FailureCodeNone {
		return unknownEvaluation(code)
	}

	adjacency := make(map[string][]string, len(normalized.Nodes))
	for _, edge := range normalized.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	required := reachableObligations(roots, adjacency, byID)
	if len(required) == 0 {
		return unknownEvaluation(FailureCodeNoReachableObligations)
	}
	for _, root := range roots {
		if len(reachableObligations([]string{root}, adjacency, byID)) == 0 {
			return unknownEvaluation(FailureCodeNoReachableObligations)
		}
	}

	requiredSet := make(map[string]struct{}, len(required))
	for _, id := range required {
		requiredSet[id] = struct{}{}
	}
	executedRequired := make([]string, 0, len(executed))
	extra := make([]string, 0, len(executed))
	for _, id := range executed {
		if _, required := requiredSet[id]; required {
			executedRequired = append(executedRequired, id)
		} else {
			extra = append(extra, id)
		}
	}
	missed := difference(required, executedRequired)
	decision := DecisionPass
	failureCode := FailureCodeNone
	if len(missed) != 0 {
		decision = DecisionFailClosed
		failureCode = FailureCodeMissedObligations
	}
	return Evaluation{
		Required: required, ExecutedRequired: executedRequired,
		Missed: missed, Extra: extra, Numerator: len(executedRequired),
		Denominator: len(required), Decision: decision, FailureCode: failureCode,
		FullSuiteRequired: false,
	}
}

// Evaluate is the package-level form for callers that keep graphs by value.
func Evaluate(graph Graph, changedIDs, executedObligationIDs []string) Evaluation {
	return graph.Evaluate(changedIDs, executedObligationIDs)
}

func inputIDs(ids []string, byID map[string]NodeKind, obligationsOnly bool) ([]string, string) {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if err := validateID(id); err != nil {
			if obligationsOnly {
				return nil, FailureCodeUnknownExecutedObligation
			}
			return nil, FailureCodeUnknownChangedNode
		}
		kind, registered := byID[id]
		if !registered {
			if obligationsOnly {
				return nil, FailureCodeUnknownExecutedObligation
			}
			return nil, FailureCodeUnknownChangedNode
		}
		if obligationsOnly && kind != NodeKindObligation {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if _, duplicate := seen[id]; duplicate {
			if obligationsOnly {
				return nil, FailureCodeAmbiguousExecutedInput
			}
			return nil, FailureCodeAmbiguousChangedInput
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Strings(result)
	return result, FailureCodeNone
}

func executedIDs(ids []string, byID map[string]NodeKind) ([]string, string) {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if err := validateID(id); err != nil {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if kind, registered := byID[id]; registered && kind != NodeKindObligation {
			return nil, FailureCodeUnknownExecutedObligation
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, FailureCodeAmbiguousExecutedInput
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Strings(result)
	return result, FailureCodeNone
}

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
