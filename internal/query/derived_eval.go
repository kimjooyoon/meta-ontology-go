package query

import "sort"

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
		deterministic = graph.deriveInverseRows(root, rule, FactDeterministic, options.Limit, nil)
	}
	if options.Selection == SelectDeterministic || len(deterministic) >= options.Limit {
		return deterministic, nil
	}
	blocked := derivedKeys(deterministic)
	candidates := graph.deriveInverseRows(
		root, rule, FactCandidate, options.Limit-len(deterministic), blocked,
	)
	if options.Selection == SelectCandidate {
		return nil, candidates
	}
	return deterministic, candidates
}

func (graph Graph) deriveInverseRows(
	root ID, rule derivedRule, status FactStatus, limit int, blocked map[derivedKey]struct{},
) []DerivedFact {
	rows := make(map[derivedKey]DerivedFact)
	for _, fact := range graph.AllFacts() {
		if fact.Status != status || fact.Predicate != rule.base || fact.Object != root {
			continue
		}
		derived := newDerivedFact(rule, fact.Object, fact.Subject, 1, fact.Status)
		key := derivedKey{derived.Subject, derived.Predicate, derived.Object}
		if _, exists := blocked[key]; exists {
			continue
		}
		rows[key] = derived
	}
	return limitSortedDerived(rows, limit)
}

func (graph Graph) deriveDependsOn(root ID, options DerivedOptions) ([]DerivedFact, []DerivedFact) {
	var deterministic []DerivedFact
	if options.Selection != SelectCandidate {
		deterministic = graph.deriveDependsOnRows(
			root, options.MaxDepth, SelectDeterministic, false, options.Limit, nil,
		)
	}
	if options.Selection == SelectDeterministic || len(deterministic) >= options.Limit {
		return deterministic, nil
	}
	candidates := graph.deriveDependsOnRows(
		root, options.MaxDepth, options.Selection, true,
		options.Limit-len(deterministic), derivedKeys(deterministic),
	)
	if options.Selection == SelectCandidate {
		return nil, candidates
	}
	return deterministic, candidates
}

func (graph Graph) deriveDependsOnRows(
	root ID, maxDepth int, selection FactSelection, wantCandidate bool,
	limit int, blocked map[derivedKey]struct{},
) []DerivedFact {
	rows := make(map[derivedKey]DerivedFact)
	frontier := []derivedState{{node: root}}
	visited := make(map[derivedVisitKey]struct{})
	visited[derivedVisitKey{node: root}] = struct{}{}
	for depth := 1; depth <= maxDepth && len(frontier) > 0; depth++ {
		next := make([]derivedState, 0)
		for _, state := range frontier {
			for _, fact := range graph.edges(state.node, TraversalOptions{
				Predicate: WasDerivedFrom, Direction: Outgoing, Selection: selection,
			}, selection) {
				if fact.Object == root {
					continue
				}
				candidate := state.candidate || fact.Status == FactCandidate
				nextState := derivedState{node: fact.Object, depth: depth, candidate: candidate}
				visitKey := derivedVisitKey{node: nextState.node, candidate: candidate}
				if _, exists := visited[visitKey]; exists {
					continue
				}
				visited[visitKey] = struct{}{}
				next = append(next, nextState)
				if candidate != wantCandidate {
					continue
				}
				derived := newDerivedFact(
					derivedRule{id: RuleDependsOn, predicate: DerivedDependsOn},
					root, fact.Object, depth, statusForCandidate(candidate),
				)
				key := derivedKey{derived.Subject, derived.Predicate, derived.Object}
				if _, exists := blocked[key]; exists {
					continue
				}
				recordDerivedRow(rows, key, derived)
			}
		}
		sortDerivedStates(next)
		frontier = next
	}
	return limitSortedDerived(rows, limit)
}

func statusForCandidate(candidate bool) FactStatus {
	if candidate {
		return FactCandidate
	}
	return FactDeterministic
}

func newDerivedFact(rule derivedRule, subject, object ID, depth int, status FactStatus) DerivedFact {
	return DerivedFact{
		Subject: subject, Predicate: rule.predicate, Object: object,
		RuleID: rule.id, Depth: depth, Status: DerivedFactStatus,
		SourceLayer: status.String(),
	}
}

func derivedKeys(rows []DerivedFact) map[derivedKey]struct{} {
	keys := make(map[derivedKey]struct{}, len(rows))
	for _, row := range rows {
		keys[derivedKey{row.Subject, row.Predicate, row.Object}] = struct{}{}
	}
	return keys
}

func recordDerivedRow(rows map[derivedKey]DerivedFact, key derivedKey, row DerivedFact) {
	if existing, ok := rows[key]; !ok || row.Depth < existing.Depth {
		rows[key] = row
	}
}

func limitSortedDerived(rows map[derivedKey]DerivedFact, limit int) []DerivedFact {
	result := sortedDerived(rows)
	if len(result) > limit {
		result = result[:limit]
	}
	return append([]DerivedFact(nil), result...)
}

func sortedDerived(rows map[derivedKey]DerivedFact) []DerivedFact {
	result := make([]DerivedFact, 0, len(rows))
	for _, row := range rows {
		result = append(result, row)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Depth != right.Depth {
			return left.Depth < right.Depth
		}
		return left.SourceLayer < right.SourceLayer
	})
	return result
}

func sortDerivedStates(states []derivedState) {
	sort.Slice(states, func(i, j int) bool {
		if states[i].node != states[j].node {
			return states[i].node < states[j].node
		}
		if states[i].candidate != states[j].candidate {
			return !states[i].candidate
		}
		return states[i].depth < states[j].depth
	})
}
