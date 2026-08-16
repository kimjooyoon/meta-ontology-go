package semantic

import (
	"errors"
	"strings"
	"testing"
)

func TestGraphValidationRequiresDeclaredNodesAndPROVKinds(t *testing.T) {
	g := NewGraph()
	activity := MustIdentity("billing://activity/pay")
	entity := MustIdentity("billing://entity/order")
	before := g.Canonical()
	if err := g.AddFact(NewUsedFact(activity, entity)); !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("missing node add error = %v, want ErrNodeNotFound", err)
	}
	if g.Canonical() != before {
		t.Fatal("rejected missing-node fact mutated the graph")
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("empty graph after rejected fact is invalid: %v", err)
	}

	g = NewGraph()
	if err := g.AddNode(mustEntity(t, activity, Namespace("billing"), "Pay")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddNode(mustEntity(t, entity, Namespace("billing"), "Order")); err != nil {
		t.Fatal(err)
	}
	if err := g.AddFact(NewUsedFact(activity, entity)); !errors.Is(err, ErrInvalidFact) {
		t.Fatalf("reversed used edge error = %v, want ErrInvalidFact", err)
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("rejected edge mutated graph: %v", err)
	}
}

func TestGraphValidationRejectsUnknownRelationsAndKeepsCandidateStatus(t *testing.T) {
	g := NewGraph()
	activity := mustActivity(t, MustIdentity("billing://activity/pay"), Namespace("billing"), "Pay")
	entity := mustEntity(t, MustIdentity("billing://entity/order"), Namespace("billing"), "Order")
	if err := g.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := g.AddNode(entity); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCandidate(Fact{Subject: activity.ID, Predicate: Relation("calls"), Object: entity.ID}); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("unknown relation error = %v, want ErrUnknownRelation", err)
	}
	if err := g.AddCandidate(NewCandidateFact(activity.ID, Used, entity.ID, "needs explicit domain assertion")); err != nil {
		t.Fatal(err)
	}
	if err := g.Validate(); err != nil {
		t.Fatalf("valid candidate graph failed validation: %v", err)
	}
	if got := g.Candidates()[0].Status; got != FactCandidate {
		t.Fatalf("candidate status = %v, want %v", got, FactCandidate)
	}
}

func TestIRValidationAndHashDelegateToNormalizedGraph(t *testing.T) {
	ir := NewIR(" billing ", Namespace(" billing "))
	activity := mustActivity(t, MustIdentity("billing://activity/pay"), Namespace("billing"), "Pay")
	entity := mustEntity(t, MustIdentity("billing://entity/order"), Namespace("billing"), "Order")
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(entity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(NewUsedFact(activity.ID, entity.ID)); err != nil {
		t.Fatal(err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Package != "billing" || normalized.Namespace != Namespace("billing") {
		t.Fatalf("IR metadata was not normalized: %#v", normalized)
	}
	if normalized.StableHash() == "" || normalized.StableHash() != normalized.StableHash() {
		t.Fatal("IR stable hash is not deterministic")
	}
	if err := normalized.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestIRValidationRejectsUnknownSchemaVersion(t *testing.T) {
	ir := NewIR("billing", Namespace("billing"))
	ir.Version = "semantic-ir/v2"
	if err := ir.Validate(); !errors.Is(err, ErrGraphInvalid) {
		t.Fatalf("unknown IR version error = %v, want ErrGraphInvalid", err)
	}
	if _, err := ir.Normalized(); !errors.Is(err, ErrGraphInvalid) {
		t.Fatalf("unknown IR version normalization error = %v, want ErrGraphInvalid", err)
	}

	ir.Version = " semantic-ir/v1 "
	if err := ir.Validate(); err != nil {
		t.Fatalf("trimmed current IR version was rejected: %v", err)
	}
}

func TestGraphValidationRejectsStaleIdentityIndexes(t *testing.T) {
	graph := NewGraph()
	node := mustEntity(t, MustIdentity("billing://entity/order"), Namespace("billing"), "Order")
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err)
	}
	delete(graph.names, node.NameRef())
	if err := graph.Validate(); err == nil || !strings.Contains(err.Error(), "name-index-missing") {
		t.Fatalf("missing name index error = %v, want deterministic index diagnostic", err)
	}

	graph = NewGraph()
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err)
	}
	delete(graph.nodes, node.ID)
	graph.nodes[ID(" BILLING://ENTITY/ORDER ")] = node
	if err := graph.Validate(); err == nil || !strings.Contains(err.Error(), "node-key") {
		t.Fatalf("non-canonical node key error = %v, want deterministic key diagnostic", err)
	}
}
