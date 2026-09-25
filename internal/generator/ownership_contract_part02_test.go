package generator

import (
	"bytes"
	"strings"
	"testing"
)

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
