package generator

import (
	"fmt"
	"sort"
)

func normalizeMarkerManifestV1(source []byte, markers parsedMarkers, ir SemanticIR) ([]markerManifestRegionV1, error) {
	expectedRegions, expectedSlots, err := expectedMarkerManifestOwnersV1(ir)
	if err != nil {
		return nil, err
	}
	regions := make([]markerManifestRegionV1, 0, len(markers.Regions))
	seenRegions := make(map[string]struct{}, len(markers.Regions))
	seenSlots := make(map[string]markerManifestSlotV1, len(markers.Slots))
	for _, observed := range markers.Regions {
		if err := validateMarkerRegionV1(source, observed, expectedRegions, seenRegions); err != nil {
			return nil, err
		}
		region := markerManifestRegionV1{ID: observed.ID, Kind: observed.Kind, Start: observed.Start, End: observed.End, StartLine: observed.StartLine, EndLine: observed.EndLine}
		for _, observedSlot := range observed.Slots {
			slot, err := validateMarkerSlotV1(source, observed, observedSlot, expectedSlots, seenSlots)
			if err != nil {
				return nil, err
			}
			region.Slots = append(region.Slots, slot)
		}
		sort.SliceStable(region.Slots, func(left, right int) bool { return markerManifestSlotLessV1(region.Slots[left], region.Slots[right]) })
		if err := validateMarkerSlotOrderV1(region); err != nil {
			return nil, err
		}
		regions = append(regions, region)
	}
	if len(regions) != len(expectedRegions) {
		return nil, fmt.Errorf("generator: marker manifest v1 region count %d does not match declared count %d", len(regions), len(expectedRegions))
	}
	if len(seenSlots) != len(expectedSlots) || len(markers.Slots) != len(expectedSlots) {
		return nil, fmt.Errorf("generator: marker manifest v1 slot count does not match declared count")
	}
	for slotID, expected := range seenSlots {
		observed, exists := markers.Slots[slotID]
		if !exists || !sameParsedSlotV1(source, observed, expected) {
			return nil, fmt.Errorf("generator: marker manifest v1 slot map disagrees for %q", slotID)
		}
	}
	sort.SliceStable(regions, func(left, right int) bool { return markerManifestRegionLessV1(regions[left], regions[right]) })
	if err := validateMarkerRegionOrderV1(regions); err != nil {
		return nil, err
	}
	return regions, nil
}
