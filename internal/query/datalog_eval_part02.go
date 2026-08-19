package query

import (
	"fmt"
	"strings"
)

func applyDatalogRule(
	rule DatalogRule, byPredicate map[string][]DatalogFact,
	depths map[DatalogFactKey]int, budget *datalogWorkBudget,
) ([]DatalogFact, error) {
	results := make([]DatalogFact, 0)
	var join func(int, datalogBinding, []DatalogFact)
	join = func(index int, binding datalogBinding, support []DatalogFact) {
		if index == len(rule.Body) {
			subject := datalogTermValue(rule.Head.Subject, binding)
			object := datalogTermValue(rule.Head.Object, binding)
			depth := 1
			namespace := ""
			for _, fact := range support {
				if fact.Namespace != "" {
					namespace = fact.Namespace
				}
				if candidateDepth := depths[fact.Key()] + 1; candidateDepth > depth {
					depth = candidateDepth
				}
			}
			results = append(results, DatalogFact{
				Namespace: namespace, Subject: subject,
				Predicate: rule.Head.Predicate, Object: object,
				Origin: DatalogDerived, RuleID: rule.ID, Depth: depth,
				Support: datalogSupport(support),
			})
			return
		}
		atom := rule.Body[index]
		for _, fact := range byPredicate[atom.Predicate] {
			if !budget.take() {
				return
			}
			next, ok := bindDatalogAtom(atom, fact, binding)
			if !ok {
				continue
			}
			join(index+1, next, append(support, fact))
		}
	}
	join(0, make(datalogBinding), nil)
	if budget.exhausted {
		return nil, fmt.Errorf("%w: maximum work %d", ErrDatalogBudget, budget.limit)
	}
	return results, nil
}
func bindDatalogAtom(atom DatalogAtom, fact DatalogFact, binding datalogBinding) (datalogBinding, bool) {
	next := make(datalogBinding, len(binding)+2)
	for name, value := range binding {
		next[name] = value
	}
	for _, pair := range [][2]DatalogTerm{{atom.Subject, {Constant: fact.Subject}}, {atom.Object, {Constant: fact.Object}}} {
		if pair[0].Variable != "" {
			name := strings.TrimPrefix(pair[0].Variable, "?")
			if existing, exists := next[name]; exists && existing != pair[1].Constant {
				return nil, false
			}
			next[name] = pair[1].Constant
		} else if pair[0].Constant != pair[1].Constant {
			return nil, false
		}
	}
	return next, true
}
