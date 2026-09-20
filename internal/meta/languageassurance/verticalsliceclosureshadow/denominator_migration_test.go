package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCurrentDenominatorPinsExistingCapabilities(t *testing.T) {
	raw := activeDenominator()
	if digestBytes(raw) != DenominatorMigrationV35Digest || activeDenominatorDigest() != DenominatorMigrationV35Digest {
		t.Fatal("active denominator is not the pinned v35 migration")
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
	if observed.Version != 35 || len(observed.Boundaries) != 6 ||
		observed.Boundaries[0].ID != "syntax" || observed.Boundaries[0].Target != 66 {
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
	lowered := bytes.Replace(raw, []byte(`"target": 66`), []byte(`"target": 65`), 1)
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
		digestBytes(embeddedDenominatorV34) != DenominatorMigrationV34Digest {
		t.Fatal("historical denominator evidence was rewritten")
	}
	var previous, current denominator
	if err := json.Unmarshal(embeddedDenominatorV34, &previous); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(activeDenominator(), &current); err != nil {
		t.Fatal(err)
	}
	if previous.Version != 34 || current.Version != 35 ||
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
