package semantic

import (
	"errors"
	"testing"
)

func TestFieldTypeRegistryResolvesStableIDsAndRejectsUnknownAmbiguousOrInvalidRefs(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	registry := NewTypeRegistry()
	valid := testEntityWithFields(t, entityID, testStringField(entityID, "name"))
	graph := NewGraph()
	if err := graph.AddNode(valid); err != nil {
		t.Fatal(err)
	}
	if err := graph.ValidateWithTypes(registry); err != nil {
		t.Fatalf("built-in string type rejected: %v", err)
	}

	unknown := testStringField(entityID, "unknown")
	unknown.TypeRef = TypeRef{ID: MustIdentity("billing://type/unknown")}
	unknownGraph := NewGraph()
	if err := unknownGraph.AddNode(testEntityWithFields(t, entityID, unknown)); err != nil {
		t.Fatal(err)
	}
	if err := unknownGraph.ValidateWithTypes(registry); !errors.Is(err, ErrUnknownType) {
		t.Fatalf("unknown type error = %v, want ErrUnknownType", err)
	}

	if err := registry.Register(TypeDef{ID: MustIdentity("alt://type/string"), Namespace: Namespace("alt"), Name: "string"}); err != nil {
		t.Fatal(err)
	}
	ambiguous := testStringField(entityID, "ambiguous")
	ambiguous.TypeRef = TypeRef{Name: "string"}
	ambiguousGraph := NewGraph()
	if err := ambiguousGraph.AddNode(testEntityWithFields(t, entityID, ambiguous)); err != nil {
		t.Fatal(err)
	}
	if err := ambiguousGraph.ValidateWithTypes(registry); !errors.Is(err, ErrAmbiguousType) {
		t.Fatalf("ambiguous type error = %v, want ErrAmbiguousType", err)
	}

	invalid := testStringField(entityID, "invalid")
	invalid.TypeRef = TypeRef{ID: "not an identity"}
	if err := (&Graph{}).AddNode(testEntityWithFields(t, entityID, invalid)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid type ref error = %v, want ErrInvalidField", err)
	}

	lookup := testStringField(entityID, "lookup")
	lookup.TypeRef = TypeRef{Name: BuiltinStringTypeName, Namespace: BuiltinTypeNamespace}
	lookupGraph := NewGraph()
	if err := lookupGraph.AddNode(testEntityWithFields(t, entityID, lookup)); err != nil {
		t.Fatal(err)
	}
	normalized, err := lookupGraph.NormalizedWithTypes(NewTypeRegistry())
	if err != nil {
		t.Fatalf("lookup TypeRef normalization failed: %v", err)
	}
	got := normalized.Nodes()[0].Fields[0].TypeRef
	if got.ID != BuiltinStringTypeID || got.Name != "" || got.Namespace != "" {
		t.Fatalf("lookup metadata was not reduced to stable ID: %#v", got)
	}
}
