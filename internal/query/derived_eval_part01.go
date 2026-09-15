package query

type derivedState struct {
	node      ID
	depth     int
	candidate bool
}
type derivedVisitKey struct {
	node      ID
	candidate bool
}

func (graph Graph) deriveInverse(root ID, rule derivedRule, options DerivedOptions) ([]DerivedFact, []DerivedFact) {
	var deterministic []DerivedFact
	if options.Selection != SelectCandidate {
		deterministic = graph.deriveInverseRows(
			root, rule, FactDeterministic, options.Limit, nil, newQueryWorkQuota(options.Limit),
		)
	}
	if options.Selection == SelectDeterministic || len(deterministic) >= options.Limit {
		return deterministic, nil
	}
	blocked := derivedKeys(deterministic)
	candidates := graph.deriveInverseRows(
		root, rule, FactCandidate, options.Limit-len(deterministic), blocked,
		newQueryWorkQuota(options.Limit),
	)
	if options.Selection == SelectCandidate {
		return nil, candidates
	}
	return deterministic, candidates
}
func (graph Graph) deriveInverseRows(
	root ID, rule derivedRule, status FactStatus, limit int, blocked map[derivedKey]struct{},
	quota *queryWorkQuota,
) []DerivedFact {
	rows := make([]DerivedFact, 0, limit)
	for _, fact := range graph.AllFacts() {
		if fact.Status != status || fact.Predicate != rule.base || fact.Object != root {
			continue
		}
		if !quota.take() {
			return rows
		}
		derived := newDerivedFact(rule, fact.Object, fact.Subject, 1, fact.Status)
		key := derivedKey{derived.Subject, derived.Predicate, derived.Object}
		if _, exists := blocked[key]; exists {
			continue
		}
		rows = append(rows, derived)
		if len(rows) == limit {
			break
		}
	}
	return rows
}
