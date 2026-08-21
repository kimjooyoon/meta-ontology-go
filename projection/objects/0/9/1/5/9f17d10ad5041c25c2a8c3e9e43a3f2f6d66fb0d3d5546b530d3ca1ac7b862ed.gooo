package semantic

import (
	"errors"
	"testing"
)

func TestFieldIDsAndOwnerNamesAreValidatedWithoutCrossParentNameCollisions(t *testing.T) {
	firstID := MustIdentity("billing://entity/first")
	secondID := MustIdentity("billing://entity/second")

	duplicateID := testEntityWithFields(t, firstID,
		testStringField(firstID, "same-id"),
		testStringField(firstID, "same-id"),
	)
	if err := (&Graph{}).AddNode(duplicateID); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("duplicate field ID error = %v, want ErrInvalidField", err)
	}

	nameCollision := testEntityWithFields(t, firstID,
		testStringField(firstID, "first"),
		func() Field {
			field := testStringField(firstID, "second")
			field.Name = "first"
			return field
		}(),
	)
	if err := (&Graph{}).AddNode(nameCollision); !errors.Is(err, ErrNameCollision) {
		t.Fatalf("same-owner name collision error = %v, want ErrNameCollision", err)
	}

	aliasCollision := testEntityWithFields(t, firstID,
		testStringField(firstID, "alias-first"),
		func() Field {
			field := testStringField(firstID, "alias-second")
			field.Aliases = []string{"alias-first"}
			return field
		}(),
	)
	if err := (&Graph{}).AddNode(aliasCollision); !errors.Is(err, ErrNameCollision) {
		t.Fatalf("same-owner alias collision error = %v, want ErrNameCollision", err)
	}

	graph := NewGraph()
	firstField := testStringField(firstID, "name")
	secondField := testStringField(secondID, "other-name")
	secondField.Name = "name"
	if err := graph.AddNode(testEntityWithFields(t, firstID, firstField)); err != nil {
		t.Fatal(err)
	}
	secondNode := testEntityWithFields(t, secondID, secondField)
	secondNode.Name = "Receipt"
	if err := graph.AddNode(secondNode); err != nil {
		t.Fatalf("cross-parent equal field names rejected: %v", err)
	}
}
