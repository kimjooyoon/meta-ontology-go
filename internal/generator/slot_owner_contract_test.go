package generator

import (
	"bytes"
	"strings"
	"testing"
)

func TestSlotOwnerChangeFailsClosedWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	changed := acceptanceFixture()
	changed.Activities[0].Slots = nil
	changed.Activities[1].Slots = []Slot{
		{ID: "gooo://slot/inspect-implementation", Default: "return artifact"},
		{ID: "gooo://slot/compile-implementation", Default: "return artifact"},
	}
	previous := append([]byte(nil), first.Source...)
	if _, err := Generate(changed, previous); err == nil || !strings.Contains(err.Error(), "changes region owner") {
		t.Fatalf("expected slot owner rejection, got %v", err)
	}
	if !bytes.Equal(previous, first.Source) {
		t.Fatal("slot owner rejection mutated previous source")
	}
}

func TestSameSlotOwnerPreservesBodyAndSourceMap(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := bytes.Replace(first.Source, []byte("return Artifact{}"), []byte("return Artifact{}\n\t// same owner preserved"), 1)
	second := mustAcceptanceResult(t, acceptanceFixture(), previous)
	if !bytes.Contains(second.Source, []byte("// same owner preserved")) {
		t.Fatal("same-owner slot body was not preserved")
	}
	if len(second.SourceMap.Lookup("gooo://slot/compile-implementation")) != 1 {
		t.Fatal("same-owner slot source-map identity was lost")
	}
}

func TestSlotOwnerKindMismatchFailsClosedWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corrupted := strings.Replace(string(first.Source), `//gooo:generated:start id="gooo://activity/compile" kind="activity"`, `//gooo:generated:start id="gooo://activity/compile" kind="entity"`, 1)
	corrupted = strings.Replace(corrupted, `//gooo:generated:end id="gooo://activity/compile" kind="activity"`, `//gooo:generated:end id="gooo://activity/compile" kind="entity"`, 1)
	previous := []byte(corrupted)
	if _, err := Generate(acceptanceFixture(), previous); err == nil || !strings.Contains(err.Error(), "changes kind") {
		t.Fatalf("expected slot kind rejection, got %v", err)
	}
	if !bytes.Equal(previous, []byte(corrupted)) {
		t.Fatal("slot kind rejection mutated previous source")
	}
}

func TestRegionKindChangeFailsClosedWithoutMutation(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	changed := acceptanceFixture()
	activity := changed.Activities[0]
	changed.Activities = changed.Activities[1:]
	changed.Entities = append(changed.Entities, Entity{
		ID: activity.ID, Name: activity.Name, GoName: activity.GoName,
	})
	previous := append([]byte(nil), first.Source...)
	if _, err := Generate(changed, previous); err == nil || !strings.Contains(err.Error(), "changes kind") {
		t.Fatalf("expected generated-region kind rejection, got %v", err)
	}
	if !bytes.Equal(previous, first.Source) {
		t.Fatal("generated-region kind rejection mutated previous source")
	}
}

func TestRemovedRegionSlotCannotMoveToNewOwner(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	changed := acceptanceFixture()
	changed.Activities = changed.Activities[1:]
	changed.Activities[0].Slots = []Slot{
		{ID: "gooo://slot/inspect-implementation", Default: "return artifact"},
		{ID: "gooo://slot/compile-implementation", Default: "return artifact"},
	}
	previous := append([]byte(nil), first.Source...)
	if _, err := Generate(changed, previous); err == nil || !strings.Contains(err.Error(), "changes region owner") {
		t.Fatalf("expected removed-region owner rejection, got %v", err)
	}
	if !bytes.Equal(previous, first.Source) {
		t.Fatal("removed-region owner rejection mutated previous source")
	}
}
