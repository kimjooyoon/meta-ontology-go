package semantic

import (
	"fmt"
	"sort"
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
	// Candidate observations may be type-incomplete until review. They remain
	// outside authoritative hashes, while Validate and promotion fail closed.
	if normalized.Status == FactDeterministic {
		if err := g.validateDeclaredFactKinds(normalized); err != nil {
			return Fact{}, err
		}
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

func mergeFact(existing, incoming Fact) Fact {
	merged := existing
	if merged.Span.IsZero() && !incoming.Span.IsZero() {
		merged.Span = incoming.Span
	}
	if merged.Reason == "" && incoming.Reason != "" {
		merged.Reason = incoming.Reason
	}
	return merged
}

func (g Graph) Facts() []Fact {
	return g.DeterministicFacts()
}

func (g Graph) DeterministicFacts() []Fact {
	facts := make([]Fact, 0, len(g.facts))
	for _, fact := range g.facts {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

func (g Graph) Candidates() []Fact {
	facts := make([]Fact, 0, len(g.candidates))
	for _, fact := range g.candidates {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// SortedFacts is an explicit alias for adapters that prefer sorted snapshots.
func (g Graph) SortedFacts() []Fact {
	return g.AllFacts()
}

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

func (g *Graph) PromoteCandidate(key FactKey) (Fact, error) {
	key, err := normalizeFactKey(key)
	if err != nil {
		return Fact{}, err
	}
	g.ensure()
	candidate, ok := g.candidates[key]
	if !ok {
		return Fact{}, fmt.Errorf("%w: %s %s %s", ErrCandidateNotFound, key.Subject, key.Predicate, key.Object)
	}
	candidate.Status = FactDeterministic
	if err := g.AddFact(candidate); err != nil {
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
