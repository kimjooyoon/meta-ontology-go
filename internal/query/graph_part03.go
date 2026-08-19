package query

import (
	"fmt"
)

// Add inserts a fact, canonicalizing its IDs and relation. A deterministic
// fact removes a candidate with the same triple; a candidate cannot shadow a
// deterministic fact.
func (graph *Graph) Add(fact Fact) error {
	normalized, err := fact.Normalized()
	if err != nil {
		return err
	}
	graph.ensure()
	if err := graph.validateFactEndpoints(normalized); err != nil {
		return err
	}
	graph.ensureImplicitEndpoint(normalized.Subject)
	graph.ensureImplicitEndpoint(normalized.Object)
	key := normalized.Key()
	if normalized.Status == FactCandidate {
		if _, exists := graph.deterministic[key]; exists {
			return nil
		}
		graph.candidates[key] = mergeFacts(graph.candidates[key], normalized)
		return nil
	}
	graph.deterministic[key] = mergeFacts(graph.deterministic[key], normalized)
	delete(graph.candidates, key)
	return nil
}
func (graph *Graph) ensureImplicitEndpoint(id ID) {
	if _, exists := graph.nodes[id]; !exists {
		graph.nodes[id] = Node{ID: id, Kind: UnknownNodeKind}
	}
}
func (graph Graph) validateFactEndpoints(fact Fact) error {
	subject, subjectOK := graph.nodes[fact.Subject]
	object, objectOK := graph.nodes[fact.Object]
	if !subjectOK || !objectOK {
		return nil
	}
	requiredSubject, requiredObject, known := relationNodeKinds(fact.Predicate)
	if !known {
		return fmt.Errorf("%w: %q", ErrInvalidRelation, fact.Predicate)
	}
	if subject.Kind != UnknownNodeKind && subject.Kind != requiredSubject {
		return fmt.Errorf("%w: %s requires subject %s, got %s", ErrInvalidFact, fact.Predicate, requiredSubject, subject.Kind)
	}
	if object.Kind != UnknownNodeKind && object.Kind != requiredObject {
		return fmt.Errorf("%w: %s requires object %s, got %s", ErrInvalidFact, fact.Predicate, requiredObject, object.Kind)
	}
	return nil
}
func (graph *Graph) AddDeterministic(fact Fact) error {
	fact.Status = FactDeterministic
	return graph.Add(fact)
}
func (graph *Graph) AddCandidate(fact Fact) error {
	fact.Status = FactCandidate
	return graph.Add(fact)
}
func (graph Graph) DeterministicFacts() []Fact {
	facts := make([]Fact, 0, len(graph.deterministic))
	for _, fact := range graph.deterministic {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// Facts is the conventional spelling for the deterministic layer.
func (graph Graph) Facts() []Fact { return graph.DeterministicFacts() }
