package verticalsliceclosureshadow

import (
	"encoding/json"
	"testing"
)

func TestCollectionDenominatorMigration8553(t *testing.T) {
	if digestBytes(embeddedDenominatorV48) != DenominatorMigrationV48Digest ||
		digestBytes(embeddedDenominatorV49) != DenominatorMigrationV49Digest {
		t.Fatal("versioned source bytes lost their identity")
	}
	next, err := decodeDenominator(embeddedDenominatorV49)
	if err != nil || next.Version != 49 || next.Boundaries[0].Target != 82 {
		t.Fatal("ordered collection scope", next, err)
	}
	var prior denominator
	if err := json.Unmarshal(embeddedDenominatorV48, &prior); err != nil {
		t.Fatal(err)
	}
	if len(prior.Boundaries) != 6 || len(next.Boundaries) != 6 {
		t.Fatal("boundary inventory changed")
	}
	for index, expected := range prior.Boundaries {
		if index == 0 {
			expected.Target += 2
		}
		if next.Boundaries[index] != expected {
			t.Fatalf("unrelated boundary changed: %d", index)
		}
	}
	t.Log("8553: v49 records two new syntax capabilities; prior bytes, other boundaries and all dependency counts preserved")
}
