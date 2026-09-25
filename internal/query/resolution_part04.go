package query

func (graph Graph) resolutionLayer(
	root ID, facts []Fact, selection FactSelection, limit int, blocked map[resolutionKey]struct{},
) []ResolutionRow {
	used := make([]Fact, 0)
	generated := make(map[ID][]Fact)
	for _, fact := range facts {
		if !selection.includes(fact.Status) {
			continue
		}
		switch fact.Predicate {
		case Used:
			if fact.Object == root {
				used = append(used, fact)
			}
		case WasGeneratedBy:
			generated[fact.Object] = append(generated[fact.Object], fact)
		}
	}
	sortFacts(used)
	for _, facts := range generated {
		sortFacts(facts)
	}
	rows := make([]ResolutionRow, 0, limit)
	for _, usedFact := range used {
		activity, ok := graph.Node(usedFact.Subject)
		if !ok || activity.Kind != ActivityNodeKind {
			continue
		}
		for _, generatedFact := range generated[usedFact.Subject] {
			output, ok := graph.Node(generatedFact.Subject)
			if !ok || output.Kind != EntityNodeKind {
				continue
			}
			candidate := usedFact.Status == FactCandidate || generatedFact.Status == FactCandidate
			if selection == SelectDeterministic && candidate {
				continue
			}
			if selection == SelectCandidate && !candidate {
				continue
			}
			row := newResolutionRow(root, usedFact.Subject, generatedFact.Subject, candidate)
			key := resolutionKey{row.Business, row.Activity, row.GeneratedEntity}
			if _, exists := blocked[key]; exists {
				continue
			}
			rows = append(rows, row)
			if len(rows) == limit {

				sortResolutionRows(rows)
				return append([]ResolutionRow(nil), rows...)
			}
		}
	}
	sortResolutionRows(rows)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return append([]ResolutionRow(nil), rows...)
}

type resolutionKey struct {
	business, activity, generated ID
}
