package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestFixtureRemovalPreservesRetainedRegionAndSlot(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := bytes.Replace(first.Source, []byte("package bootstrapgen\n"), []byte("package bootstrapgen\n\nvar Keep = 7\n"), 1)
	previous = bytes.Replace(previous, []byte("return Artifact{}"), []byte("return Artifact{Digest: source.Digest}"), 1)
	changed := acceptanceFixture()
	changed.Activities = changed.Activities[:1]
	second := mustAcceptanceResult(t, changed, previous)
	if !strings.Contains(string(second.Source), "var Keep = 7") {
		t.Fatal("marker-outside handwritten text was not preserved")
	}
	if !bytes.Equal(testGeneratedBlock(t, first.Source, "gooo://entity/source"), testGeneratedBlock(t, second.Source, "gooo://entity/source")) {
		t.Fatal("retained entity region changed")
	}
	if !strings.Contains(string(second.Source), "return Artifact{Digest: source.Digest}") {
		t.Fatal("handwritten slot was not preserved")
	}
	if strings.Contains(string(second.Source), `id="gooo://activity/inspect"`) || len(second.SourceMap.Lookup("gooo://activity/inspect")) != 0 {
		t.Fatal("removed activity region or source-map entry remained")
	}
}

func TestFixtureRollbackRejectsOrphanMarkerWithoutMutation(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corrupted := append(append([]byte(nil), result.Source...), []byte("\n//gooo:slot:start id=\"orphan\"\n")...)
	previous := append([]byte(nil), corrupted...)
	_, err := Generate(acceptanceFixture(), previous)
	if err == nil || !strings.Contains(err.Error(), "slot outside generated region") {
		t.Fatalf("expected orphan marker rejection, got %v", err)
	}
	if !bytes.Equal(previous, corrupted) {
		t.Fatal("rejected source was mutated")
	}
}

func TestFixtureDeclarationPermutationIsReproducible(t *testing.T) {
	firstIR := acceptanceFixture()
	secondIR := acceptanceFixture()
	secondIR.Entities[0], secondIR.Entities[1] = secondIR.Entities[1], secondIR.Entities[0]
	secondIR.Activities[0], secondIR.Activities[1] = secondIR.Activities[1], secondIR.Activities[0]
	first := mustAcceptanceResult(t, firstIR, nil)
	second := mustAcceptanceResult(t, secondIR, nil)
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.SourceMap, second.SourceMap) {
		t.Fatal("declaration permutation changed generated source or source map")
	}
}
