package semantic

import (
	"sort"
)

func (g Graph) AllFacts() []Fact {
	facts := make([]Fact, 0, len(g.facts)+len(g.candidates))
	facts = append(facts, g.DeterministicFacts()...)
	facts = append(facts, g.Candidates()...)
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i].Key(), facts[j].Key()
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		return facts[i].Status < facts[j].Status
	})
	return facts
}
func sortFacts(facts []Fact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i].Key(), facts[j].Key()
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		return left.Object < right.Object
	})
}
func (g Graph) HasFact(key FactKey) bool {
	key, err := normalizeFactKey(key)
	if err != nil {
		return false
	}
	_, ok := g.facts[key]
	return ok
}
func (g Graph) HasCandidate(key FactKey) bool {
	key, err := normalizeFactKey(key)
	if err != nil {
		return false
	}
	_, ok := g.candidates[key]
	return ok
}
