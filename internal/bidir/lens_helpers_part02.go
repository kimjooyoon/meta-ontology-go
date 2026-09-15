package bidir

import (
	"fmt"
	"strings"
)

func resolveReference(reference Reference, namespace string, names map[string]ID, ids map[ID]struct{}) (ID, error) {
	if reference.ID != "" {
		if err := validateID(reference.ID); err != nil {
			return "", err
		}
		if _, exists := ids[reference.ID]; !exists {
			return "", fmt.Errorf("unknown semantic ID %q", reference.ID)
		}
		return reference.ID, nil
	}
	if strings.TrimSpace(reference.Name) == "" {
		return "", fmt.Errorf("reference has neither ID nor name")
	}
	refNamespace := reference.Namespace
	if refNamespace == "" {
		refNamespace = namespace
	}
	id, exists := names[referenceKey(refNamespace, reference.Name)]
	if !exists {
		return "", fmt.Errorf("unknown declaration %q in namespace %q", reference.Name, refNamespace)
	}
	return id, nil
}
func referenceKey(namespace, name string) string {
	return namespace + "\x00" + name
}
func appendUniqueRelation(relations []Relation, relation Relation) []Relation {
	key := relationKey(relation.Kind, relation.Source, relation.Target)
	for _, existing := range relations {
		if relationKey(existing.Kind, existing.Source, existing.Target) == key {
			return relations
		}
	}
	return append(relations, relation)
}
func appendCheckedRelation(relations []Relation, relation Relation) ([]Relation, error) {
	key := relationKey(relation.Kind, relation.Source, relation.Target)
	for _, existing := range relations {
		if relationKey(existing.Kind, existing.Source, existing.Target) != key {
			continue
		}
		if relationSemanticEqual(existing, relation) {
			return relations, nil
		}
		return nil, fmt.Errorf("duplicate relation %s with conflicting attributes", key)
	}
	return append(relations, relation), nil
}
