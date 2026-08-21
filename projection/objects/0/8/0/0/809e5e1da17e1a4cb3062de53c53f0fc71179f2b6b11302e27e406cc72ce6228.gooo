package query

import (
	"fmt"
)

type datalogBinding map[string]ID
type datalogWorkBudget struct {
	remaining int
	limit     int
	exhausted bool
}

func newDatalogWorkBudget(limit int) *datalogWorkBudget {
	return &datalogWorkBudget{remaining: limit, limit: limit}
}
func (budget *datalogWorkBudget) take() bool {
	if budget.remaining == 0 {
		budget.exhausted = true
		return false
	}
	budget.remaining--
	return true
}
func deriveDatalog(
	base []DatalogFact, rules []DatalogRule, limit, maxDepth int, budget *datalogWorkBudget,
) ([]DatalogFact, error) {
	byPredicate := make(map[string][]DatalogFact)
	known := make(map[DatalogFactKey]struct{})
	depths := make(map[DatalogFactKey]int)
	for _, fact := range base {
		byPredicate[fact.Predicate] = append(byPredicate[fact.Predicate], fact)
		known[fact.Key()] = struct{}{}
		depths[fact.Key()] = fact.Depth
	}
	for predicate := range byPredicate {
		sortDatalogFacts(byPredicate[predicate])
	}
	derived := make([]DatalogFact, 0)
	for {
		changed := false
		for _, rule := range rules {
			conclusions, err := applyDatalogRule(rule, byPredicate, depths, budget)
			if err != nil {
				return nil, err
			}
			for _, conclusion := range conclusions {
				if _, exists := known[conclusion.Key()]; exists {
					continue
				}
				if conclusion.Depth > maxDepth {
					return nil, fmt.Errorf("%w: maximum derivation depth %d", ErrDatalogBudget, maxDepth)
				}
				if len(derived) >= limit {
					return nil, fmt.Errorf("%w: maximum derived facts %d", ErrDatalogBudget, limit)
				}
				known[conclusion.Key()] = struct{}{}
				depths[conclusion.Key()] = conclusion.Depth
				derived = append(derived, conclusion)
				byPredicate[conclusion.Predicate] = append(byPredicate[conclusion.Predicate], conclusion)
				sortDatalogFacts(byPredicate[conclusion.Predicate])
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	sortDatalogFacts(derived)
	return derived, nil
}
