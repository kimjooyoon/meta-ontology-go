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

func (graph Graph) deriveDependsOn(root ID, options DerivedOptions) ([]DerivedFact, []DerivedFact) {
	nodes := graph.Nodes()
	var deterministic []DerivedFact
	if options.Selection != SelectCandidate {
		deterministic = graph.deriveTargets(
			root, nodes, options.MaxDepth, SelectDeterministic, false, options.Limit, nil,
			newQueryWorkQuota(options.Limit),
		)
	}
	if options.Selection == SelectDeterministic || len(deterministic) >= options.Limit {
		return deterministic, nil
	}
	candidates := graph.deriveTargets(
		root, nodes, options.MaxDepth, options.Selection, true,
		options.Limit-len(deterministic), derivedKeys(deterministic),
		newQueryWorkQuota(options.Limit),
	)
	if options.Selection == SelectCandidate {
		return nil, candidates
	}
	return deterministic, candidates
}

func (graph Graph) deriveTargets(
	root ID, nodes []Node, maxDepth int, selection FactSelection,
	wantCandidate bool, limit int, blocked map[derivedKey]struct{}, quota *queryWorkQuota,
) []DerivedFact {
	if limit <= 0 {
		return nil
	}
	index := graph.reverseDerivedEdges(selection)
	rows := make([]DerivedFact, 0, limit)
	for _, node := range nodes {
		if node.ID == root {
			continue
		}
		if len(index[node.ID]) == 0 {
			continue
		}
		derived, ok, complete := findDerivedTarget(
			root, node.ID, maxDepth, selection, wantCandidate, index, quota,
		)
		if !complete {
			return rows
		}
		if !ok {
			continue
		}
		key := derivedKey{derived.Subject, derived.Predicate, derived.Object}
		if _, exists := blocked[key]; exists {
			continue
		}
		rows = append(rows, derived)
		if len(rows) == limit {
			return rows
		}
	}
	return rows
}

func (graph Graph) reverseDerivedEdges(selection FactSelection) map[ID][]Fact {
	index := make(map[ID][]Fact)
	for _, fact := range graph.AllFacts() {
		if selection.includes(fact.Status) && fact.Predicate == WasDerivedFrom {
			index[fact.Object] = append(index[fact.Object], fact)
		}
	}
	for _, facts := range index {
		sortFacts(facts)
	}
	return index
}

func findDerivedTarget(
	root, target ID, maxDepth int, selection FactSelection, wantCandidate bool,
	index map[ID][]Fact, quota *queryWorkQuota,
) (DerivedFact, bool, bool) {
	frontier := []derivedState{{node: target}}
	visited := make(map[derivedVisitKey]struct{})
	visited[derivedVisitKey{node: target}] = struct{}{}
	for head := 0; head < len(frontier); head++ {
		state := frontier[head]
		if state.depth == maxDepth {
			continue
		}
		for _, fact := range index[state.node] {
			if !quota.take() {
				return DerivedFact{}, false, false
			}
			candidate := state.candidate || fact.Status == FactCandidate
			depth := state.depth + 1
			if fact.Subject == root {
				if candidate != wantCandidate {
					continue
				}
				return newDerivedFact(
					derivedRule{id: RuleDependsOn, predicate: DerivedDependsOn},
					root, target, depth, statusForCandidate(candidate),
				), true, true
			}
			nextState := derivedState{node: fact.Subject, depth: depth, candidate: candidate}
			visitKey := derivedVisitKey{node: nextState.node, candidate: candidate}
			if _, exists := visited[visitKey]; exists {
				continue
			}
			visited[visitKey] = struct{}{}
			frontier = append(frontier, nextState)
		}
	}
	return DerivedFact{}, false, true
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
