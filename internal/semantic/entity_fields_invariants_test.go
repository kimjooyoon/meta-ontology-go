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

func TestFieldDeclarationOrderParticipatesInCanonicalSemanticModel(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	first := testStringField(entityID, "first")
	second := testStringField(entityID, "second")
	left := NewGraph()
	right := NewGraph()
	if err := left.AddNode(testEntityWithFields(t, entityID, first, second)); err != nil {
		t.Fatal(err)
	}
	if err := right.AddNode(testEntityWithFields(t, entityID, second, first)); err != nil {
		t.Fatal(err)
	}
	if left.StableHash() == right.StableHash() {
		t.Fatal("field declaration order did not participate in the canonical semantic model")
	}
}

func TestInvalidFieldsDoNotMutateFactsOrEvidence(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	activityID := MustIdentity("billing://activity/pay")
	ir := NewIR("billing", Namespace("billing"))
	if err := ir.AddNode(mustEntity(t, entityID, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(mustActivity(t, activityID, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	fact := NewUsedFact(activityID, entityID)
	if err := ir.AddFact(fact); err != nil {
		t.Fatal(err)
	}
	evidence, err := NewEvidence(
		MustIdentity("billing://evidence/field-fixture"), GoVerifierID, VerificationEvidence,
		fact.Key(), StableHashString("field-fixture"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	beforeCanonical := ir.Canonical()
	beforeEvidence := ir.EvidenceHash()
	bad := testStringField(entityID, "malformed")
	bad.Parent = MustIdentity("billing://entity/other")
	invalid := testEntityWithFields(t, entityID, bad)
	if err := ir.AddNode(invalid); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("malformed field error = %v, want ErrInvalidField", err)
	}
	if ir.Canonical() != beforeCanonical || ir.EvidenceHash() != beforeEvidence {
		t.Fatal("rejected incomplete field mutated IR facts or evidence")
	}
}

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
