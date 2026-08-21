package query

import (
	"fmt"
)

func (graph Graph) CandidateFacts() []Fact {
	facts := make([]Fact, 0, len(graph.candidates))
	for _, fact := range graph.candidates {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// Candidates is the conventional spelling for the candidate layer.
func (graph Graph) Candidates() []Fact { return graph.CandidateFacts() }

// AllFacts returns a detached deterministic ordering of both fact layers.
func (graph Graph) AllFacts() []Fact {
	facts := append(graph.DeterministicFacts(), graph.CandidateFacts()...)
	sortFacts(facts)
	return facts
}

// Relations returns all canonical relation rows in deterministic order.
func (graph Graph) Relations() []Fact { return graph.AllFacts() }
func (graph Graph) HasFact(key FactKey) bool {
	key, err := normalizeKey(key)
	if err != nil {
		return false
	}
	_, exists := graph.deterministic[key]
	return exists
}
func (graph Graph) HasCandidate(key FactKey) bool {
	key, err := normalizeKey(key)
	if err != nil {
		return false
	}
	_, exists := graph.candidates[key]
	return exists
}
func (graph Graph) requireEndpoint(id ID) error {
	if _, ok := graph.nodes[id]; !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEndpoint, id)
	}
	return nil
}
func normalizeKey(key FactKey) (FactKey, error) {
	fact, err := NewFact(key.Subject, key.Predicate, key.Object).Normalized()
	if err != nil {
		return FactKey{}, err
	}
	return fact.Key(), nil
}
func mergeFacts(existing, incoming Fact) Fact {
	if existing.Subject == "" {
		return incoming
	}
	if existing.Reason == "" && incoming.Reason != "" {
		existing.Reason = incoming.Reason
	}
	return existing
}
