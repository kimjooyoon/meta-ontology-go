package generator

import (
	"bytes"
	"strings"
	"testing"
)

func TestFixtureSourceMapPreservesDeclaredSlotOrder(t *testing.T) {
	ir := acceptanceFixture()
	ir.Activities[0].Slots = []Slot{
		{ID: "gooo://slot/compile-implementation", Default: "return Artifact{}"},
		{ID: "gooo://slot/compile-audit", Default: "return Artifact{}"},
	}
	result := mustAcceptanceResult(t, ir, nil)
	mappings := result.SourceMap.Lookup("gooo://slot/compile-implementation")
	if len(mappings) != 1 || mappings[0].Ordinal != 0 {
		t.Fatalf("first declared slot lost its ordinal: %#v", mappings)
	}
	mappings = result.SourceMap.Lookup("gooo://slot/compile-audit")
	if len(mappings) != 1 || mappings[0].Ordinal != 1 {
		t.Fatalf("second declared slot lost its ordinal: %#v", mappings)
	}
}
func TestFixtureDuplicateRegionRollbackIsAtomic(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	duplicate := append(append([]byte(nil), result.Source...), testGeneratedBlock(t, result.Source, "gooo://entity/source")...)
	previous := append([]byte(nil), duplicate...)
	_, err := Generate(acceptanceFixture(), previous)
	if err == nil || !strings.Contains(err.Error(), "duplicate generated region ID") {
		t.Fatalf("expected duplicate region rejection, got %v", err)
	}
	if !bytes.Equal(previous, duplicate) {
		t.Fatal("duplicate-region rejection mutated previous source")
	}
}
func fixtureMapping(t *testing.T, sourceMap SourceMap, id string) SourceMapping {
	t.Helper()
	mappings := sourceMap.Lookup(id)
	if len(mappings) != 1 {
		t.Fatalf("expected one mapping for %q, got %#v", id, mappings)
	}
	return mappings[0]
}
