package query

import (
	"fmt"
)

// EvaluateDatalog evaluates positive rules over deterministic graph facts and
// then matches the requested patterns. It is a read-only projection and does
// not promote candidates or alter the graph hash.
func (graph Graph) EvaluateDatalog(request DatalogQuery) (DatalogResult, error) {
	normalized, rules, err := normalizeDatalogQuery(request)
	if err != nil {
		return DatalogResult{}, err
	}

	declared := make([]DatalogFact, 0, len(graph.DeterministicFacts()))
	for _, fact := range graph.DeterministicFacts() {
		declared = append(declared, DatalogFact{
			Namespace: graphScope(graph),
			Subject:   fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
			Origin: DatalogDeclared, Depth: 0,
		})
	}
	sortDatalogFacts(declared)
	var derived []DatalogFact
	workBudget := newDatalogWorkBudget(normalized.MaxWork)
	if normalized.IncludeDerived {
		derived, err = deriveDatalog(
			declared, rules, normalized.MaxDerivedFacts,
			normalized.MaxDepth, workBudget,
		)
		if err != nil {
			return DatalogResult{Complete: false}, err
		}
	}

	universe := append([]DatalogFact(nil), declared...)
	if normalized.IncludeDerived {
		universe = append(universe, derived...)
	}
	if normalized.IncludeCandidates {
		for _, fact := range graph.CandidateFacts() {
			universe = append(universe, DatalogFact{
				Namespace: graphScope(graph),
				Subject:   fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
				Origin: DatalogCandidate, Depth: 0,
			})
		}
	}
	sortDatalogFacts(universe)
	rows, err := matchDatalogPatterns(normalized.Patterns, universe, workBudget)
	if err != nil {
		return DatalogResult{Complete: false}, err
	}
	complete := true
	if len(rows) > normalized.Limit {
		rows = rows[:normalized.Limit]
		complete = false
		return DatalogResult{Rows: rows, Derived: derived, Complete: false},
			fmt.Errorf("%w: maximum result rows %d", ErrDatalogBudget, normalized.Limit)
	}
	return DatalogResult{Rows: rows, Derived: derived, Complete: complete}, nil
}
func graphScope(graph Graph) string {
	if graph.binding == nil {
		return ""
	}
	return graph.binding.namespace
}
