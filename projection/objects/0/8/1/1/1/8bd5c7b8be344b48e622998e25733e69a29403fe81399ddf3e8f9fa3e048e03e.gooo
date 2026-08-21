package semantic

import (
	"errors"
	"strings"
	"testing"
)

func TestFieldChangesAreSemanticAndFactsRemainSeparate(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	activityID := MustIdentity("billing://activity/pay")
	base := testEntityWithFields(t, entityID, testStringField(entityID, "name"))
	changed := base
	changed.Fields = copyFields(base.Fields)
	changed.Fields[0].Cardinality = Many
	if base.StableHash() == changed.StableHash() {
		t.Fatal("field cardinality did not participate in semantic identity")
	}

	graph := NewGraph()
	if err := graph.AddNode(base); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddNode(mustActivity(t, activityID, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	fact := NewCandidateFact(activityID, Used, entityID, "candidate evidence")
	if err := graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	if len(graph.Facts()) != 0 || len(graph.Candidates()) != 1 {
		t.Fatalf("field addition crossed fact boundary: facts=%d candidates=%d", len(graph.Facts()), len(graph.Candidates()))
	}
	if strings.Contains(graph.SemanticCanonical(), "candidate evidence") {
		t.Fatal("candidate explanation leaked into semantic canonical form")
	}
}
func TestInvalidFieldAdditionDoesNotMutateOriginalIR(t *testing.T) {
	entityID := MustIdentity("billing://entity/order")
	ir := NewIR("billing", Namespace("billing"))
	if err := ir.AddNode(mustEntity(t, entityID, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	beforeCanonical := ir.Canonical()
	beforeHash := ir.StableHash()
	bad := testStringField(entityID, "bad")
	bad.Parent = MustIdentity("billing://entity/other")
	if err := ir.AddNode(testEntityWithFields(t, entityID, bad)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("invalid field error = %v, want ErrInvalidField", err)
	}
	if ir.Canonical() != beforeCanonical || ir.StableHash() != beforeHash {
		t.Fatal("rejected field addition mutated original IR")
	}
}
func TestFieldsAreLatentUntilDirectSemanticConstruction(t *testing.T) {

	node, err := NewEntity(MustIdentity("billing://entity/order"), Namespace("billing"), "Order")
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Fields) != 0 {
		t.Fatal("field semantics became reachable from the existing node constructor")
	}
}
