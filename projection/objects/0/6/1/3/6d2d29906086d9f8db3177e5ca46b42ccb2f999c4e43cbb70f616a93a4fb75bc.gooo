package semantic

import (
	"fmt"
)

func (g Graph) addContractFacts(activity ID, span Span, ids []ID, predicate Relation) ([]Fact, error) {
	facts := make([]Fact, 0, len(ids))
	for _, object := range ids {
		fact, err := g.prepareFact(NewFact(activity, predicate, object).WithSpan(span))
		if err != nil {
			return nil, err
		}
		facts = append(facts, fact)
	}
	return facts, nil
}
func (g Graph) addContractOutputs(activity ID, span Span, entities []ID) ([]Fact, error) {
	facts := make([]Fact, 0, len(entities))
	for _, entity := range entities {
		fact := NewWasGeneratedByFact(entity, activity).WithSpan(span)
		normalized, err := g.prepareFact(fact)
		if err != nil {
			return nil, err
		}
		facts = append(facts, normalized)
	}
	return facts, nil
}
func normalizeFactKey(key FactKey) (FactKey, error) {
	subject, err := ParseIdentity(key.Subject.String())
	if err != nil {
		return FactKey{}, err
	}
	object, err := ParseIdentity(key.Object.String())
	if err != nil {
		return FactKey{}, err
	}
	if !key.Predicate.Valid() {
		return FactKey{}, fmt.Errorf("%w: %s", ErrUnknownRelation, key.Predicate)
	}
	return FactKey{Subject: subject, Predicate: key.Predicate, Object: object}, nil
}
