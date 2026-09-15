package query

func normalizeDatalogBounds(request DatalogQuery) (DatalogQuery, error) {
	if len(request.Patterns) == 0 || len(request.Patterns) > MaxDatalogBodyAtoms {
		return DatalogQuery{}, datalogError("query requires 1..%d patterns", MaxDatalogBodyAtoms)
	}
	if len(request.Rules) > MaxDatalogRules {
		return DatalogQuery{}, datalogError("rule count exceeds %d", MaxDatalogRules)
	}
	if request.Limit < 0 || request.Limit > MaxDatalogLimit {
		return DatalogQuery{}, datalogError("limit must be 0..%d", MaxDatalogLimit)
	}
	if request.Limit == 0 {
		request.Limit = DefaultDatalogLimit
	}
	if request.MaxDerivedFacts < 0 || request.MaxDerivedFacts > DefaultDatalogDerivedLimit {
		return DatalogQuery{}, datalogError("max derived facts must be 0..%d", DefaultDatalogDerivedLimit)
	}
	if request.MaxDerivedFacts == 0 {
		request.MaxDerivedFacts = DefaultDatalogDerivedLimit
	}
	if request.MaxDepth < 0 || request.MaxDepth > MaxDatalogDepth {
		return DatalogQuery{}, datalogError("max depth must be 0..%d", MaxDatalogDepth)
	}
	if request.MaxDepth == 0 {
		request.MaxDepth = DefaultDatalogDepth
	}
	if request.MaxWork < 0 || request.MaxWork > MaxDatalogWork {
		return DatalogQuery{}, datalogError("max work must be 0..%d", MaxDatalogWork)
	}
	if request.MaxWork == 0 {
		request.MaxWork = DefaultDatalogWork
	}
	return request, nil
}
func validateDatalogPredicates(patterns []DatalogAtom, rules []DatalogRule, heads map[string]struct{}) error {
	known := map[string]struct{}{
		string(Used): {}, string(WasGeneratedBy): {}, string(WasDerivedFrom): {}, string(WasAssociatedWith): {},
	}
	for predicate := range heads {
		known[predicate] = struct{}{}
	}
	for _, rule := range rules {
		for _, atom := range rule.Body {
			if _, exists := known[atom.Predicate]; !exists {
				return datalogError("rule %q references unknown predicate %q", rule.ID, atom.Predicate)
			}
		}
	}
	for _, pattern := range patterns {
		if _, exists := known[pattern.Predicate]; !exists {
			return datalogError("query references unknown predicate %q", pattern.Predicate)
		}
	}
	return nil
}
