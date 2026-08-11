package generator

import (
	"bytes"
	"strings"
	"testing"
)

func TestFixtureMultipleHandwrittenSlotsSurviveIRDefaults(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := strings.Replace(string(first.Source), "return Artifact{}", "return Artifact{Digest: source.Digest}", 1)
	previous = strings.Replace(previous, "return artifact", "return Artifact{Digest: artifact.Digest}", 1)
	changed := acceptanceFixture()
	changed.Activities[0].Slots[0].Default = "panic(\"new compile default\")"
	changed.Activities[1].Slots[0].Default = "panic(\"new inspect default\")"
	second := mustAcceptanceResult(t, changed, []byte(previous))
	if !strings.Contains(string(second.Source), "return Artifact{Digest: source.Digest}") || !strings.Contains(string(second.Source), "return Artifact{Digest: artifact.Digest}") {
		t.Fatal("handwritten slot content was replaced by a new default")
	}
	if !bytes.Equal(testGeneratedBlock(t, first.Source, "gooo://entity/source"), testGeneratedBlock(t, second.Source, "gooo://entity/source")) {
		t.Fatal("unrelated generated region changed")
	}
}

func TestFixtureSourceMapMatchesProtectedMarkerBounds(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	for _, region := range markers.Regions {
		mapping := fixtureMapping(t, result.SourceMap, region.ID)
		if mapping.Generated.Start.Offset != region.Start || mapping.Generated.End.Offset != region.End {
			t.Fatalf("region %q source map does not match marker bounds: %#v", region.ID, mapping)
		}
		for _, slot := range region.Slots {
			slotMapping := fixtureMapping(t, result.SourceMap, slot.ID)
			if slotMapping.Generated.Start.Offset != slot.Start || slotMapping.Generated.End.Offset != slot.End {
				t.Fatalf("slot %q source map does not match marker bounds: %#v", slot.ID, slotMapping)
			}
		}
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
