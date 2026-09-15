package semantic

import (
	"errors"
	"testing"
)

func testStringField(parent ID, suffix string) Field {
	return Field{
		ID:          MustIdentity("billing://field/" + suffix),
		Parent:      parent,
		Name:        suffix,
		TypeRef:     TypeRef{ID: BuiltinStringTypeID},
		Presence:    Required,
		Cardinality: One,
		Span:        Span{File: "entity-fields.gooo", Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}},
	}
}
func testEntityWithFields(t *testing.T, id ID, fields ...Field) Node {
	t.Helper()
	node := mustEntity(t, id, Namespace("billing"), "Order")
	node.Fields = fields
	return node
}
func TestFieldRequiresExplicitIdentityAndEntityParent(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	field := testStringField(entityID, "name")
	graph := NewGraph()
	if err := graph.AddNode(testEntityWithFields(t, entityID, field)); err != nil {
		t.Fatalf("valid entity field rejected: %v", err)
	}
	node, ok := graph.Node(entityID)
	if !ok || len(node.Fields) != 1 {
		t.Fatalf("stored fields = %#v, ok=%v", node.Fields, ok)
	}
	if node.Fields[0].ID != field.ID || node.Fields[0].Parent != entityID {
		t.Fatalf("stored field identity = %#v", node.Fields[0])
	}
	if err := graph.ValidateWithTypes(NewTypeRegistry()); err != nil {
		t.Fatalf("valid typed field graph rejected: %v", err)
	}

	badID := field
	badID.ID = "not an absolute identity"
	if err := (&Graph{}).AddNode(testEntityWithFields(t, entityID, badID)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid field ID error = %v, want ErrInvalidField", err)
	}

	activity := mustActivity(t, MustIdentity("billing://activity/pay"), Namespace("billing"), "Pay")
	activity.Fields = []Field{field}
	if err := (&Graph{}).AddNode(activity); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("activity field error = %v, want ErrInvalidField", err)
	}
}
