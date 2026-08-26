package cycles

import (
	"fmt"
	"strings"
)

func normalizedName(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}
func normalizedNamespace(raw string) string {
	return strings.TrimSpace(raw)
}
func relationName(edge Edge) Relation {
	return edge.predicate()
}
func directionMessage(edge Edge, subject, object Node, known bool) string {
	predicate := relationName(edge)
	if !known {
		return fmt.Sprintf("unknown relation %q", predicate)
	}
	expectedSubject, expectedObject, _ := expectedKinds(predicate)
	return fmt.Sprintf("%s expects %s -> %s, got %s -> %s", predicate,
		expectedSubject, expectedObject, subject.Kind, object.Kind)
}
