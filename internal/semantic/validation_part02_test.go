package semantic

import (
	"errors"
	"strings"
	"testing"
)

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
	for _, code := range []string{"field-index-missing", "field-index-owner", "field-index-stale"} {
		t.Run(code, func(t *testing.T) {
			current := NewGraph()
			field := testStringField(node.ID, "name")
			if err := current.AddNode(testEntityWithFields(t, node.ID, field)); err != nil {
				t.Fatal(err)
			}
			canonical, hash := current.Canonical(), current.StableHash()
			switch code {
			case "field-index-missing":
				delete(current.fieldOwners, field.ID)
			case "field-index-owner":
				current.fieldOwners[field.ID] = MustIdentity("billing://entity/other")
			case "field-index-stale":
				current.fieldOwners[MustIdentity("billing://field/stale")] = node.ID
			}
			if err := current.Validate(); err == nil || !strings.Contains(err.Error(), code) {
				t.Fatalf("corrupt derived index error = %v, want %s", err, code)
			}
			if current.Canonical() != canonical || current.StableHash() != hash {
				t.Fatal("derived index changed authoritative source identity")
			}
		})
	}
}
