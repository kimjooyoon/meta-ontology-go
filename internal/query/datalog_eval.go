package query

import (
	"fmt"
	"strings"
)

type datalogBinding map[string]ID

func deriveDatalog(base []DatalogFact, rules []DatalogRule, limit int) ([]DatalogFact, error) {
	byPredicate := make(map[string][]DatalogFact)
	known := make(map[DatalogFactKey]struct{})
	for _, fact := range base {
		byPredicate[fact.Predicate] = append(byPredicate[fact.Predicate], fact)
		known[fact.Key()] = struct{}{}
	}
	for predicate := range byPredicate {
		sortDatalogFacts(byPredicate[predicate])
	}
	derived := make([]DatalogFact, 0)
	for {
		changed := false
		for _, rule := range rules {
			for _, conclusion := range applyDatalogRule(rule, byPredicate) {
				if _, exists := known[conclusion.Key()]; exists {
					continue
				}
				if len(derived) >= limit {
					return nil, fmt.Errorf("%w: maximum derived facts %d", ErrDatalogBudget, limit)
				}
				known[conclusion.Key()] = struct{}{}
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

func applyDatalogRule(rule DatalogRule, byPredicate map[string][]DatalogFact) []DatalogFact {
	results := make([]DatalogFact, 0)
	var join func(int, datalogBinding, []DatalogFact)
	join = func(index int, binding datalogBinding, support []DatalogFact) {
		if index == len(rule.Body) {
			subject := datalogTermValue(rule.Head.Subject, binding)
			object := datalogTermValue(rule.Head.Object, binding)
			results = append(results, DatalogFact{
				Subject: subject, Predicate: rule.Head.Predicate, Object: object,
				Origin: DatalogDerived, RuleID: rule.ID, Support: datalogSupport(support),
			})
			return
		}
		atom := rule.Body[index]
		for _, fact := range byPredicate[atom.Predicate] {
			next, ok := bindDatalogAtom(atom, fact, binding)
			if !ok {
				continue
			}
			join(index+1, next, append(support, fact))
		}
	}
	join(0, make(datalogBinding), nil)
	return results
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

func datalogTermValue(term DatalogTerm, binding datalogBinding) ID {
	if term.Variable == "" {
		return term.Constant
	}
	return binding[strings.TrimPrefix(term.Variable, "?")]
}

func datalogSupport(support []DatalogFact) []DatalogFactKey {
	keys := make([]DatalogFactKey, len(support))
	for index, fact := range support {
		keys[index] = fact.Key()
	}
	return keys
}
