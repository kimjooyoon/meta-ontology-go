package semantic

import (
	"errors"
	"testing"
)

func TestIdentityIsAbsoluteAndCanonical(t *testing.T) {
	id, err := ParseIdentity(" Billing://Entity/Order ")
	if err != nil {
		t.Fatalf("ParseIdentity() error = %v", err)
	}
	if got, want := id.String(), "billing://entity/Order"; got != want {
		t.Fatalf("identity = %q, want %q", got, want)
	}
	if _, err := ParseIdentity("Order"); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("relative identity error = %v, want ErrInvalidIdentity", err)
	}
	if _, err := ParseIdentity("billing://"); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("empty authority error = %v, want ErrInvalidIdentity", err)
	}
}

func TestIdentityIsIndependentFromNamespaceAndDisplayName(t *testing.T) {
	ns, err := ParseNamespace("billing")
	if err != nil {
		t.Fatal(err)
	}
	id := MustIdentity("billing://entity/order")
	first, err := NewEntity(id, ns, "Order")
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewEntity(id, ns, "Purchase")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatal("renaming a declaration changed its semantic identity")
	}
	if first.Name == second.Name {
		t.Fatal("test did not exercise a display-name change")
	}
}

func TestGraphNodeLookupUsesCanonicalIdentity(t *testing.T) {
	graph := NewGraph()
	node := Node{
		ID:        ID(" BILLING://ENTITY/order "),
		Kind:      Entity,
		Namespace: Namespace("billing"),
		Name:      "Order",
	}
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err)
	}
	for _, lookup := range []ID{node.ID, MustIdentity("billing://entity/order")} {
		got, ok := graph.Node(lookup)
		if !ok {
			t.Fatalf("canonical identity lookup %q did not resolve", lookup)
		}
		if got.ID != MustIdentity("billing://entity/order") {
			t.Fatalf("lookup returned non-canonical identity %q", got.ID)
		}
	}
	if _, ok := graph.Node(ID("not an identity")); ok {
		t.Fatal("invalid identity lookup unexpectedly resolved")
	}
}
