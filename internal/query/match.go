package query

// ExactMatch returns the deterministic and candidate facts for one complete
// triple. It does not treat names, prefixes, or relation direction as fuzzy
// matches; IDs and predicates are canonicalized before lookup.
func (graph Graph) ExactMatch(query ExactQuery) (MatchResult, error) {
	normalized, err := query.normalized()
	if err != nil {
		return MatchResult{}, err
	}
	key := FactKey{Subject: normalized.Subject, Predicate: normalized.Predicate, Object: normalized.Object}
	result := MatchResult{}
	if fact, exists := graph.deterministic[key]; exists {
		result.Deterministic = []Fact{fact}
	}
	if fact, exists := graph.candidates[key]; exists {
		result.Candidates = []Fact{fact}
	}
	return result, nil
}

// Match is the query-oriented spelling of ExactMatch.
func (graph Graph) Match(query ExactQuery) (MatchResult, error) {
	return graph.ExactMatch(query)
}
