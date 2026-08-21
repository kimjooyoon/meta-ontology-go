package generator

import (
	"bytes"
	"strings"
	"testing"
)

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
