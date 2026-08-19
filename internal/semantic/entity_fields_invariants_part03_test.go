package semantic

import (
	"errors"
	"testing"
)

func TestUnknownFieldTypeNormalizationFailsWithoutPartialGraph(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	unknown := testStringField(entityID, "unknown")
	unknown.TypeRef = TypeRef{ID: MustIdentity("billing://type/unknown")}
	graph := NewGraph()
	if err := graph.AddNode(testEntityWithFields(t, entityID, unknown)); err != nil {
		t.Fatal(err)
	}
	before := graph.Canonical()
	if _, err := graph.NormalizedWithTypes(NewTypeRegistry()); !errors.Is(err, ErrUnknownType) {
		t.Fatalf("unknown type normalization error = %v, want ErrUnknownType", err)
	}
	if graph.Canonical() != before {
		t.Fatal("unknown type normalization mutated the source graph")
	}
}
