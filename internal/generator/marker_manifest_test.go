package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
)

func TestCanonicalMarkerManifestV1HasCheckedInEncoding(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := canonicalMarkerManifestV1(result.Source, markers, ir)
	if err != nil {
		t.Fatal(err)
	}
	const header = "schema\tgooo.generator.marker-manifest\tprofile\tgenerated-regions-and-slots\tversion\t1\tencoding\tutf-8-byte-length-hex\tterminal-newline\tLF\n"
	if !bytes.HasPrefix(manifest, []byte(header)) || !bytes.HasSuffix(manifest, []byte{'\n'}) {
		t.Fatalf("manifest does not use the checked-in LF format:\n%s", manifest)
	}
	if strings.Contains(string(manifest), "\r") || !strings.Contains(string(manifest), "regions\t4\n") || !strings.Contains(string(manifest), "slot\t") {
		t.Fatalf("manifest omitted required typed records:\n%s", manifest)
	}
	digest := sha256.Sum256(manifest)
	const wantDigest = "f35f03e7bb130e1dadfb65ed40b5fda91eed61620f2d2f3fface06e08f971d23"
	if hex.EncodeToString(digest[:]) != wantDigest {
		t.Fatalf("canonical marker manifest digest changed: got %s want %s", hex.EncodeToString(digest[:]), wantDigest)
	}
}

func TestCanonicalMarkerManifestV1IsIdempotentAcrossObservationOrder(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	first, err := canonicalMarkerManifestV1(result.Source, markers, ir)
	if err != nil {
		t.Fatal(err)
	}
	permuted := cloneParsedMarkersV1(markers)
	for left, right := 0, len(permuted.Regions)-1; left < right; left, right = left+1, right-1 {
		permuted.Regions[left], permuted.Regions[right] = permuted.Regions[right], permuted.Regions[left]
	}
	for index := range permuted.Regions {
		for left, right := 0, len(permuted.Regions[index].Slots)-1; left < right; left, right = left+1, right-1 {
			permuted.Regions[index].Slots[left], permuted.Regions[index].Slots[right] = permuted.Regions[index].Slots[right], permuted.Regions[index].Slots[left]
		}
	}
	second, err := canonicalMarkerManifestV1(result.Source, permuted, ir)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("equivalent marker observations produced different canonical bytes")
	}
	changed := cloneParsedMarkersV1(markers)
	changed.Regions[0].ID, changed.Regions[1].ID = changed.Regions[1].ID, changed.Regions[0].ID
	third, err := canonicalMarkerManifestV1(result.Source, changed, ir)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, third) {
		t.Fatal("marker identity change did not change canonical bytes")
	}
}

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

func markerRegionIndexV1(markers parsedMarkers, id string) int {
	for index, region := range markers.Regions {
		if region.ID == id {
			return index
		}
	}
	return -1
}

func cloneParsedMarkersV1(markers parsedMarkers) parsedMarkers {
	clone := parsedMarkers{Regions: make([]generatedRegion, len(markers.Regions)), Slots: make(map[string]parsedSlot, len(markers.Slots))}
	for index, region := range markers.Regions {
		clone.Regions[index] = region
		clone.Regions[index].Slots = append([]parsedSlot(nil), region.Slots...)
		for slotIndex := range clone.Regions[index].Slots {
			clone.Regions[index].Slots[slotIndex].Body = append([]byte(nil), region.Slots[slotIndex].Body...)
		}
	}
	for id, slot := range markers.Slots {
		slot.Body = append([]byte(nil), slot.Body...)
		clone.Slots[id] = slot
	}
	return clone
}
