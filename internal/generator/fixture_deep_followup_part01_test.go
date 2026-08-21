package generator

import (
	"bytes"
	"strings"
	"testing"
)

func TestFixtureMultipleHandwrittenSlotsSurviveIRDefaults(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := strings.Replace(string(first.Source), "return Artifact{}", "return Artifact{}\n\t// source digest preserved", 1)
	previous = strings.Replace(previous, "return artifact", "return artifact\n\t// artifact digest preserved", 1)
	changed := acceptanceFixture()
	changed.Activities[0].Slots[0].Default = "panic(\"new compile default\")"
	changed.Activities[1].Slots[0].Default = "panic(\"new inspect default\")"
	second := mustAcceptanceResult(t, changed, []byte(previous))
	if !strings.Contains(string(second.Source), "// source digest preserved") || !strings.Contains(string(second.Source), "// artifact digest preserved") {
		t.Fatal("handwritten slot content was replaced by a new default")
	}
	if !bytes.Equal(testGeneratedBlock(t, first.Source, "gooo://entity/source"), testGeneratedBlock(t, second.Source, "gooo://entity/source")) {
		t.Fatal("unrelated generated region changed")
	}
}
func TestFixtureHandwrittenSlotBytesRemainExact(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	handwritten := "return Artifact{}\n\t// preserve spacing  \n"
	previous := strings.Replace(string(first.Source), "return Artifact{}", handwritten, 1)
	beforeMarkers, err := parseMarkers([]byte(previous))
	if err != nil {
		t.Fatal(err)
	}
	expected := beforeMarkers.Slots["gooo://slot/compile-implementation"].Body
	changed := acceptanceFixture()
	changed.Activities[0].Slots[0].Default = "panic(\"replacement default\")"
	second := mustAcceptanceResult(t, changed, []byte(previous))
	afterMarkers, err := parseMarkers(second.Source)
	if err != nil {
		t.Fatal(err)
	}
	actual := afterMarkers.Slots["gooo://slot/compile-implementation"].Body
	if !bytes.Equal(actual, expected) {
		t.Fatalf("handwritten slot bytes changed: got %q want %q", actual, expected)
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
