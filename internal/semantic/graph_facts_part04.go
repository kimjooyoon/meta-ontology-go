package semantic

import (
	"fmt"
)

func (g *Graph) PromoteCandidate(key FactKey) (Fact, error) {
	key, err := normalizeFactKey(key)
	if err != nil {
		return Fact{}, err
	}
	candidate, ok := g.candidates[key]
	if !ok {
		return Fact{}, fmt.Errorf("%w: %s %s %s", ErrCandidateNotFound, key.Subject, key.Predicate, key.Object)
	}
	candidate.Status = FactDeterministic
	normalized, err := g.prepareFact(candidate)
	if err != nil {
		return Fact{}, err
	}

	g.ensure()
	if err := g.storeFact(normalized); err != nil {
		return Fact{}, err
	}
	return g.facts[key], nil
}

// AddActivityContract derives only the deterministic relation patterns
// represented by the compact activity signature. It never guesses
// domain-specific relations such as delegatesTo or validatesThrough.
func (g *Graph) AddActivityContract(contract ActivityContract) error {
	activity, err := ParseIdentity(contract.Activity.String())
	if err != nil {
		return fmt.Errorf("%w: activity: %v", ErrInvalidFact, err)
	}
	span := contract.Span.Normalized()
	if err := span.Validate(); err != nil {
		return err
	}
	facts, err := g.addContractFacts(activity, span, contract.Inputs, Used)
	if err != nil {
		return err
	}
	outputs, err := g.addContractOutputs(activity, span, contract.Outputs)
	if err != nil {
		return err
	}
	facts = append(facts, outputs...)
	for _, agent := range contract.Agents {
		fact := NewWasAssociatedWithFact(activity, agent).WithSpan(span)
		normalized, err := g.prepareFact(fact)
		if err != nil {
			return err
		}
		facts = append(facts, normalized)
	}
	for _, fact := range facts {
		if err := g.storeFact(fact); err != nil {
			return err
		}
	}
	return nil
}
