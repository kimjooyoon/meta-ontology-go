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
	if err := graph.AddNode(testEntityWithFields(t, firstID, field)); err != nil {
		t.Fatal(err)
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

	parentMismatch := field
	parentMismatch.Parent = secondID
	if err := (&Graph{}).AddNode(testEntityWithFields(t, firstID, parentMismatch)); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("parent mismatch error = %v, want ErrInvalidField", err)
	}
}
