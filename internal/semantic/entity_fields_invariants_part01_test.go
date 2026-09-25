package semantic

import (
	"errors"
	"testing"
)

func TestFieldIDsShareOneCollisionDomainWithEveryNodeKind(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	activityID := MustIdentity("billing://activity/pay")
	agentID := MustIdentity("billing://agent/owner")
	fieldID := MustIdentity("billing://field/order-number")

	for _, declaration := range []Node{
		mustEntity(t, entityID, Namespace("billing"), "Order"),
		mustActivity(t, activityID, Namespace("billing"), "Pay"),
		mustAgent(t, agentID, Namespace("billing"), "Owner"),
	} {
		graph := NewGraph()
		if err := graph.AddNode(declaration); err != nil {
			t.Fatal(err)
		}
		field := testStringField(entityID, "order-number")
		field.ID = declaration.ID
		before := graph.Canonical()
		if err := graph.AddNode(testEntityWithFields(t, entityID, field)); !errors.Is(err, ErrInvalidField) {
			t.Fatalf("field/declaration collision for %s = %v, want ErrInvalidField", declaration.Kind, err)
		}
		if graph.Canonical() != before {
			t.Fatalf("field/declaration collision for %s mutated graph", declaration.Kind)
		}
	}

	graph := NewGraph()
	owner := testEntityWithFields(t, entityID, func() Field {
		field := testStringField(entityID, "order-number")
		field.ID = fieldID
		return field
	}())
	if err := graph.AddNode(owner); err != nil {
		t.Fatal(err)
	}
	before := graph.Canonical()
	if err := graph.AddNode(mustActivity(t, fieldID, Namespace("billing"), "Collision")); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("declaration/field collision = %v, want ErrInvalidField", err)
	}
	if graph.Canonical() != before {
		t.Fatal("declaration/field collision mutated graph")
	}
}
func TestTypeRefPresentationDoesNotChangeStableFieldIdentity(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	plain := testStringField(entityID, "presentation")
	withPresentation := plain
	withPresentation.TypeRef = TypeRef{
		ID: BuiltinStringTypeID, Namespace: BuiltinTypeNamespace, Name: BuiltinStringTypeName,
	}
	if withPresentation.StableHash() != plain.StableHash() {
		t.Fatal("TypeRef presentation metadata changed the stable field identity")
	}
}
