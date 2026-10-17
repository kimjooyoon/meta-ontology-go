package semantic

import (
	"errors"
	"testing"
)

func TestFieldParentMismatchAndMoveAreRejectedAtomically(t *testing.T) {
	firstID := MustIdentity("billing://entity/first")
	secondID := MustIdentity("billing://entity/second")
	field := testStringField(firstID, "name")

	graph := NewGraph()
	shared := graph
	if err := graph.AddNode(testEntityWithFields(t, firstID, field)); err != nil {
		t.Fatal(err)
	}
	if err := shared.Validate(); err != nil {
		t.Fatalf("initialized graph copy lost the field index: %v", err)
	}
	before := graph.Canonical()
	mismatched := testStringField(secondID, "mismatch")
	mismatched.ID = field.ID
	if err := graph.AddNode(testEntityWithFields(t, secondID, mismatched)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("moved field error = %v, want ErrInvalidField", err)
	}
	if graph.Canonical() != before {
		t.Fatal("rejected moved field mutated graph")
	}
	if err := graph.Validate(); err != nil {
		t.Fatalf("rejected move changed the derived index: %v", err)
	}
	if err := graph.AddNode(testEntityWithFields(t, firstID)); err != nil {
		t.Fatal(err)
	}
	if err := graph.Validate(); err != nil {
		t.Fatalf("replacing a field owner left a stale index: %v", err)
	}
	second := testEntityWithFields(t, secondID, mismatched)
	second.Name = "Second"
	if err := graph.AddNode(second); err != nil {
		t.Fatalf("removed field remained reserved by its previous owner: %v", err)
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}

	parentMismatch := field
	parentMismatch.Parent = secondID
	if err := (&Graph{}).AddNode(testEntityWithFields(t, firstID, parentMismatch)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("parent mismatch error = %v, want ErrInvalidField", err)
	}
}
