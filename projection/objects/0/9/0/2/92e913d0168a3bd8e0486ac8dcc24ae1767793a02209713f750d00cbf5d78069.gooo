package semantic

import (
	"fmt"
)

func (g *Graph) AddFact(fact Fact) error {
	normalized, err := g.prepareFact(fact)
	if err != nil {
		return err
	}
	return g.storeFact(normalized)
}
func (g Graph) prepareFact(fact Fact) (Fact, error) {
	normalized, err := fact.Normalized()
	if err != nil {
		return Fact{}, err
	}
	if err := g.validateDeclaredFactEndpoints(normalized); err != nil {
		return Fact{}, err
	}

	if err := g.validateDeclaredFactKinds(normalized); err != nil {
		return Fact{}, err
	}
	return normalized, nil
}
func (g *Graph) storeFact(normalized Fact) error {
	g.ensure()
	key := normalized.Key()
	if normalized.Status == FactCandidate {
		return g.addCandidate(normalized, key)
	}
	if existing, ok := g.facts[key]; ok {
		normalized = mergeFact(existing, normalized)
	}
	g.facts[key] = normalized
	delete(g.candidates, key)
	return nil
}
func (g Graph) validateDeclaredFactEndpoints(fact Fact) error {
	if _, ok := g.nodes[fact.Subject]; !ok {
		return fmt.Errorf("%w: fact subject %s is not declared", ErrNodeNotFound, fact.Subject)
	}
	if _, ok := g.nodes[fact.Object]; !ok {
		return fmt.Errorf("%w: fact object %s is not declared", ErrNodeNotFound, fact.Object)
	}
	return nil
}
func (g Graph) validateDeclaredFactKinds(fact Fact) error {
	subject, subjectOK := g.nodes[fact.Subject]
	object, objectOK := g.nodes[fact.Object]
	if !subjectOK || !objectOK {
		return fmt.Errorf("%w: fact endpoints are not declared", ErrNodeNotFound)
	}
	return fact.Predicate.ValidateKinds(subject.Kind, object.Kind)
}
func (g *Graph) AddCandidate(fact Fact) error {
	fact.Status = FactCandidate
	return g.AddFact(fact)
}
func (g *Graph) addCandidate(fact Fact, key FactKey) error {
	if _, deterministic := g.facts[key]; deterministic {
		return nil
	}
	if existing, ok := g.candidates[key]; ok {
		fact = mergeFact(existing, fact)
	}
	g.candidates[key] = fact
	return nil
}
