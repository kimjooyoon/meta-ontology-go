package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCurrentDenominatorPinsExistingCapabilities(t *testing.T) {
	raw := activeDenominator()
	if digestBytes(raw) != DenominatorMigrationV39Digest || activeDenominatorDigest() != DenominatorMigrationV39Digest {
		t.Fatal("active denominator is not the pinned v39 migration")
	}
	var observed struct {
		Version    int `json:"version"`
		Boundaries []struct {
			ID         string `json:"id"`
			Target     int    `json:"target"`
			LinkTarget int    `json:"link_target"`
		} `json:"boundaries"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Version != 39 || len(observed.Boundaries) != 6 ||
		observed.Boundaries[0].ID != "syntax" || observed.Boundaries[0].Target != 70 {
		t.Fatalf("migration changed the declared capability boundary: %#v", observed)
	}
	links := 0
	for _, boundary := range observed.Boundaries {
		links += boundary.LinkTarget
	}
	if links != 12 {
		t.Fatalf("migration changed the declared dependency count: %d", links)
	}
	if _, err := decodeDenominator(raw); err != nil {
		t.Fatalf("current pinned denominator was rejected: %v", err)
	}
}

func TestCurrentDenominatorRejectsLoweredTarget(t *testing.T) {
	raw := activeDenominator()
	lowered := bytes.Replace(raw, []byte(`"target": 70`), []byte(`"target": 69`), 1)
	if bytes.Equal(raw, lowered) {
		t.Fatal("counterexample did not change the syntax target")
	}
	if _, err := decodeDenominator(lowered); err == nil {
		t.Fatal("a lowered unapproved denominator was accepted")
	}
	if digestBytes(EmbeddedDenominator()) != DenominatorDigest {
		t.Fatal("the original denominator evidence was rewritten")
	}
}

func TestRecordMigrationPreservesPreviousBoundaryEvidence(t *testing.T) {
	if digestBytes(embeddedDenominatorV30) != DenominatorMigrationV30Digest ||
		digestBytes(embeddedDenominatorV31) != DenominatorMigrationV31Digest ||
		digestBytes(embeddedDenominatorV32) != DenominatorMigrationV32Digest ||
		digestBytes(embeddedDenominatorV33) != DenominatorMigrationV33Digest ||
		digestBytes(embeddedDenominatorV34) != DenominatorMigrationV34Digest ||
		digestBytes(embeddedDenominatorV35) != DenominatorMigrationV35Digest ||
		digestBytes(embeddedDenominatorV36) != DenominatorMigrationV36Digest ||
		digestBytes(embeddedDenominatorV37) != DenominatorMigrationV37Digest ||
		digestBytes(embeddedDenominatorV38) != DenominatorMigrationV38Digest {
		t.Fatal("historical denominator evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV38, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(activeDenominator(), &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 38 || current.Version != 39 ||
		len(previous.Boundaries) != 6 || len(current.Boundaries) != 6 {
		t.Fatal("migration changed the boundary inventory")
	}
	for index, expected := range previous.Boundaries {
		if index == 0 || index == 1 {
			expected.Target++
		}
		if current.Boundaries[index] != expected {
			t.Fatalf("migration changed unrelated boundary %d: %#v", index, current.Boundaries[index])
		}
	}
	if _, err := decodeDenominator(embeddedDenominatorV31); err == nil {
		t.Fatal("the historical denominator authorized the new capability")
	}
}
