package bidir

import (
	"fmt"
)

func findRelation(model Model, predicate Predicate, source, target ID) (Relation, bool) {
	for _, relation := range model.Relations {
		if relation.Kind == predicate && relation.Source == source && relation.Target == target {
			return relation, true
		}
	}
	return Relation{}, false
}
func removeRelation(relations []Relation, predicate Predicate, source, target ID) []Relation {
	result := relations[:0]
	for _, relation := range relations {
		if relation.Kind == predicate && relation.Source == source && relation.Target == target {
			continue
		}
		result = append(result, relation)
	}
	return result
}
func ensureEndpoint(model *Model, id ID, hintedKind Kind, fact Fact, subject bool) *Conflict {
	for _, node := range model.Nodes {
		if node.ID != id {
			continue
		}
		if hintedKind != "" && node.Kind != hintedKind {
			return &Conflict{Kind: ConflictKindMismatch, Fact: fact, Message: fmt.Sprintf("%s %q is %s, fact says %s", endpointLabel(subject), id, node.Kind, hintedKind)}
		}
		return nil
	}
	return &Conflict{Kind: ConflictUnknownEndpoint, Fact: fact, Message: fmt.Sprintf("%s %q is not registered in the base model", endpointLabel(subject), id)}
}
func endpointLabel(subject bool) string {
	if subject {
		return "subject"
	}
	return "object"
}
