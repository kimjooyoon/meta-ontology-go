package generator

import (
	"bytes"
	"strings"
	"testing"
)

func TestGeneratorRoundTripPreservesHandwrittenSlotBytes(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	handwritten := "return Artifact{}\n\t// preserve spacing  \n"
	previous := bytes.Replace(first.Source, []byte("return Artifact{}"), []byte(handwritten), 1)
	if bytes.Equal(previous, first.Source) {
		t.Fatal("fixture did not produce a handwritten replay input")
	}

	second := mustAcceptanceResult(t, acceptanceFixture(), previous)
	if !bytes.Equal(second.Source, previous) {
		t.Fatalf("round-trip rewrote caller-owned slot bytes:\nwant:\n%s\ngot:\n%s", previous, second.Source)
	}
	markers, err := parseMarkers(second.Source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(markers.Slots["gooo://slot/compile-implementation"].Body, []byte("\t"+handwritten+"\n")) {
		t.Fatalf("round-trip slot body changed: %q", markers.Slots["gooo://slot/compile-implementation"].Body)
	}
}

func TestGeneratorRejectsStaleSlotWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	changed := acceptanceFixture()
	changed.Activities[0].Slots = []Slot{{ID: "gooo://slot/compile-v2", Default: "return Artifact{}"}}
	previous := append([]byte(nil), first.Source...)

	_, err := Generate(changed, previous)
	if err == nil || !strings.Contains(err.Error(), "stale slot identity") {
		t.Fatalf("expected stale-slot rejection, got %v", err)
	}
	if !bytes.Equal(previous, first.Source) {
		t.Fatal("stale-slot rejection mutated previous source")
	}
}

func TestGeneratorRejectsSlotReparentingWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	changed := acceptanceFixture()
	changed.Activities[0].Slots = []Slot{{ID: "gooo://slot/compile-v2", Default: "return Artifact{}"}}
	changed.Activities[1].Slots = []Slot{
		{ID: "gooo://slot/inspect-implementation", Default: "return artifact"},
		{ID: "gooo://slot/compile-implementation", Default: "return artifact"},
	}
	previous := append([]byte(nil), first.Source...)

	_, err := Generate(changed, previous)
	if err == nil || !strings.Contains(err.Error(), "changes region owner") {
		t.Fatalf("expected slot-owner rejection, got %v", err)
	}
	if !bytes.Equal(previous, first.Source) {
		t.Fatal("slot-owner rejection mutated previous source")
	}
}

func TestGeneratorRejectsCrossOwnerStableIDCollisions(t *testing.T) {
	cases := []struct {
		name string
		edit func(*SemanticIR)
	}{
		{name: "entity and activity", edit: func(ir *SemanticIR) {
			ir.Activities[0].ID = ir.Entities[0].ID
		}},
		{name: "entity and slot", edit: func(ir *SemanticIR) {
			ir.Activities[0].Slots[0].ID = ir.Entities[0].ID
		}},
		{name: "two slots", edit: func(ir *SemanticIR) {
			ir.Activities[1].Slots[0].ID = ir.Activities[0].Slots[0].ID
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ir := acceptanceFixture()
			testCase.edit(&ir)
			if _, err := Generate(ir, nil); err == nil || !strings.Contains(err.Error(), "stable ID") {
				t.Fatalf("expected stable-ID rejection, got %v", err)
			}
		})
	}
}

func TestGeneratorRejectsRegionKindMutationWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corrupted := strings.Replace(string(first.Source),
		`//gooo:generated:start id="gooo://activity/compile" kind="activity"`,
		`//gooo:generated:start id="gooo://activity/compile" kind="entity"`, 1)
	corrupted = strings.Replace(corrupted,
		`//gooo:generated:end id="gooo://activity/compile" kind="activity"`,
		`//gooo:generated:end id="gooo://activity/compile" kind="entity"`, 1)
	previous := []byte(corrupted)

	_, err := Generate(acceptanceFixture(), previous)
	if err == nil || !strings.Contains(err.Error(), "changes kind") {
		t.Fatalf("expected region-kind rejection, got %v", err)
	}
	if !bytes.Equal(previous, []byte(corrupted)) {
		t.Fatal("region-kind rejection mutated previous source")
	}
}

func TestGeneratorRejectsUnknownMarkerAttributesWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corrupted := strings.Replace(string(first.Source),
		`//gooo:generated:start id="gooo://entity/source" kind="entity"`,
		`//gooo:generated:start id="gooo://entity/source" kind="entity" owner="user"`, 1)
	previous := []byte(corrupted)

	_, err := Generate(acceptanceFixture(), previous)
	if err == nil || !strings.Contains(err.Error(), "unknown generated-start attribute") {
		t.Fatalf("expected marker-attribute rejection, got %v", err)
	}
	if !bytes.Equal(previous, []byte(corrupted)) {
		t.Fatal("marker-attribute rejection mutated previous source")
	}
}
