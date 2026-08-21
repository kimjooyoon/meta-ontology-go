package semantic

import (
	"fmt"
)

func validateStoredFact(g Graph, key FactKey, fact Fact, expected FactStatus, issues *ValidationErrors) {
	normalized, err := fact.Normalized()
	if err != nil {
		issues.add("fact", err.Error(), fact.Subject, fact.Object)
		return
	}
	if normalized.Status != expected {
		issues.add("fact-status", fmt.Sprintf("fact is stored as %s but marked %s", expected, normalized.Status), fact.Subject, fact.Object)
	}
	if normalized.Key() != key {
		issues.add("fact-key", "fact key is not normalized", fact.Subject, fact.Object)
	}
	subject, subjectOK := g.nodes[normalized.Subject]
	object, objectOK := g.nodes[normalized.Object]
	if !subjectOK {
		issues.add("missing-subject", "fact subject is not declared", normalized.Subject, normalized.Object)
	}
	if !objectOK {
		issues.add("missing-object", "fact object is not declared", normalized.Subject, normalized.Object)
	}
	if subjectOK && objectOK {
		if err := normalized.Predicate.ValidateKinds(subject.Kind, object.Kind); err != nil {
			issues.add("relation-kind", err.Error(), normalized.Subject, normalized.Object)
		}
	}
}
func factKeyLess(left, right FactKey) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}
func (g Graph) Normalized() (Graph, error) {
	out := NewGraph()
	for _, node := range g.Nodes() {
		if err := out.AddNode(node); err != nil {
			return Graph{}, err
		}
	}
	for _, fact := range g.AllFacts() {
		if err := out.AddFact(fact); err != nil {
			return Graph{}, err
		}
	}
	if err := out.Validate(); err != nil {
		return Graph{}, err
	}
	return out, nil
}
func (g *Graph) Normalize() error {
	normalized, err := g.Normalized()
	if err != nil {
		return err
	}
	*g = normalized
	return nil
}
