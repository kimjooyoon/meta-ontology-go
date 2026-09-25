package generator

import (
	"reflect"
	"testing"
)

func TestCanonicalMarkerManifestV1RejectsMalformedEvidenceWithoutBytes(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	base, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*parsedMarkers)
	}{
		{name: "unknown region kind", mutate: func(markers *parsedMarkers) { markers.Regions[0].Kind = "unknown" }},
		{name: "unknown region id", mutate: func(markers *parsedMarkers) { markers.Regions[0].ID = "gooo://entity/unknown" }},
		{name: "duplicate region id", mutate: func(markers *parsedMarkers) { markers.Regions[1].ID = markers.Regions[0].ID }},
		{name: "region overlap", mutate: func(markers *parsedMarkers) { markers.Regions[1].Start = markers.Regions[0].Start }},
		{name: "malformed region boundary", mutate: func(markers *parsedMarkers) { markers.Regions[0].Start++ }},
		{name: "unknown slot id", mutate: func(markers *parsedMarkers) { markers.Regions[2].Slots[0].ID = "gooo://slot/unknown" }},
		{name: "duplicate slot id", mutate: func(markers *parsedMarkers) { markers.Regions[3].Slots[0].ID = markers.Regions[2].Slots[0].ID }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneParsedMarkersV1(base)
			test.mutate(&mutated)
			before := cloneParsedMarkersV1(mutated)
			manifest, err := canonicalMarkerManifestV1(result.Source, mutated, ir)
			if err == nil || manifest != nil {
				t.Fatalf("malformed marker evidence was accepted: manifest=%q err=%v", manifest, err)
			}
			if !reflect.DeepEqual(mutated, before) {
				t.Fatal("manifest rejection mutated marker evidence")
			}
		})
	}
}
func TestCanonicalMarkerManifestV1RejectsSlotOverlapWithoutBytes(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	overlapIR := acceptanceFixture()
	secondSlotID := overlapIR.Activities[1].Slots[0].ID
	overlapIR.Activities[0].Slots = append(overlapIR.Activities[0].Slots, Slot{ID: secondSlotID})
	overlapIR.Activities[1].Slots = nil
	overlapMarkers := cloneParsedMarkersV1(markers)
	firstRegion := markerRegionIndexV1(overlapMarkers, overlapIR.Activities[0].ID)
	secondRegion := markerRegionIndexV1(overlapMarkers, overlapIR.Activities[1].ID)
	overlapSlot := overlapMarkers.Regions[firstRegion].Slots[0]
	overlapSlot.ID = secondSlotID
	overlapSlot.RegionID = overlapMarkers.Regions[firstRegion].ID
	overlapSlot.RegionKind = overlapMarkers.Regions[firstRegion].Kind
	overlapMarkers.Regions[firstRegion].Slots = append(overlapMarkers.Regions[firstRegion].Slots, overlapSlot)
	overlapMarkers.Regions[secondRegion].Slots = nil
	overlapMarkers.Slots[secondSlotID] = overlapSlot
	manifest, err := canonicalMarkerManifestV1(result.Source, overlapMarkers, overlapIR)
	if err == nil || manifest != nil {
		t.Fatalf("overlapping slot evidence was accepted: manifest=%q err=%v", manifest, err)
	}
}
