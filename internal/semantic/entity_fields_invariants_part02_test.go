package semantic

import (
	"errors"
	"testing"
)

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
