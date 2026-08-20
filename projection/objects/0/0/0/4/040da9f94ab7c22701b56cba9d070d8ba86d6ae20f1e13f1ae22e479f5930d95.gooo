package query

// Traverse returns every simple path of depth 1 through MaxDepth from start.
// Deterministic paths use only deterministic facts. Candidate paths may mix
// deterministic and candidate edges, but contain at least one candidate edge.
func (graph Graph) Traverse(start ID, options TraversalOptions) (TraversalResult, error) {
	canonicalStart, err := ParseID(start.String())
	if err != nil {
		return TraversalResult{}, err
	}
	normalized, err := options.normalized()
	if err != nil {
		return TraversalResult{}, err
	}
	if err := graph.requireEndpoint(canonicalStart); err != nil {
		return TraversalResult{}, err
	}
	deterministic, candidates := graph.selectedPaths(canonicalStart, normalized)
	return TraversalResult{Deterministic: deterministic, Candidates: candidates, Metadata: graph.Metadata()}, nil
}
func (graph Graph) selectedPaths(start ID, options TraversalOptions) ([]Path, []Path) {
	deterministic := make([]Path, 0)
	candidates := make([]Path, 0)
	if options.Selection != SelectCandidate {
		deterministicQuota := newQueryWorkQuota(options.Limit)
		deterministic = graph.traversePaths(
			start, options, SelectDeterministic, FactDeterministic, options.Limit, deterministicQuota,
		)
	}
	if options.Selection == SelectDeterministic || (options.Limit > 0 && len(deterministic) >= options.Limit) {
		return deterministic, candidates
	}
	selection := SelectAll
	if options.Selection == SelectCandidate {
		selection = SelectCandidate
	}
	remaining := 0
	if options.Limit > 0 {
		remaining = options.Limit - len(deterministic)
	}
	candidateQuota := newQueryWorkQuota(remaining)
	candidates = graph.traversePaths(start, options, selection, FactCandidate, remaining, candidateQuota)
	return deterministic, candidates
}
