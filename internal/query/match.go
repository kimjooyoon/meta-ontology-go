package query

// ExactMatch returns the deterministic and candidate facts for one complete
// triple. It does not treat names, prefixes, or relation direction as fuzzy
// matches; IDs and predicates are canonicalized before lookup.
func (graph Graph) ExactMatch(query ExactQuery) (MatchResult, error) {
	return graph.ExactMatchWithOptions(query, MatchOptions{})
}

// ExactMatchWithOptions performs a read-only exact lookup over the selected
// fact layer and fails closed when either endpoint is not in the graph.
func (graph Graph) ExactMatchWithOptions(query ExactQuery, options MatchOptions) (MatchResult, error) {
	normalized, err := query.normalized()
	if err != nil {
		return MatchResult{}, err
	}
	options, err = options.normalized()
	if err != nil {
		return MatchResult{}, err
	}
	if err := graph.requireEndpoint(normalized.Subject); err != nil {
		return MatchResult{}, err
	}
	if err := graph.requireEndpoint(normalized.Object); err != nil {
		return MatchResult{}, err
	}
	key := FactKey{Subject: normalized.Subject, Predicate: normalized.Predicate, Object: normalized.Object}
	result := MatchResult{Metadata: graph.Metadata()}
	if fact, exists := graph.deterministic[key]; exists && options.Selection.includes(fact.Status) {
		result.Deterministic = []Fact{fact}
	}
	if fact, exists := graph.candidates[key]; exists && options.Selection.includes(fact.Status) {
		result.Candidates = []Fact{fact}
	}
	return result, nil
}

// ExactMatchFiltered is a concise layer-filtered exact lookup.
func (graph Graph) ExactMatchFiltered(query ExactQuery, selection FactSelection) (MatchResult, error) {
	return graph.ExactMatchWithOptions(query, MatchOptions{Selection: selection})
}

// Match is the query-oriented spelling of ExactMatch.
func (graph Graph) Match(query ExactQuery) (MatchResult, error) {
	return graph.ExactMatch(query)
}
