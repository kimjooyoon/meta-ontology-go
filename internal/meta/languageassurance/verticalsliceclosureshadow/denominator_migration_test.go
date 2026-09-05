package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCurrentDenominatorPinsExistingCapabilities(t *testing.T) {
	raw := activeDenominator()
	if digestBytes(raw) != DenominatorMigrationV28Digest || activeDenominatorDigest() != DenominatorMigrationV28Digest {
		t.Fatal("active denominator is not the pinned v28 migration")
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
	if observed.Version != 28 || len(observed.Boundaries) != 6 ||
		observed.Boundaries[0].ID != "syntax" || observed.Boundaries[0].Target != 56 {
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
	lowered := bytes.Replace(raw, []byte(`"target": 56`), []byte(`"target": 55`), 1)
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
