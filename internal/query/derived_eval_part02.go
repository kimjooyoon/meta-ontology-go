package query

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
