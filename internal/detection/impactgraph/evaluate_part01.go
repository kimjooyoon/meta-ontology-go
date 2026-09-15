package impactgraph

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
