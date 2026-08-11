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
