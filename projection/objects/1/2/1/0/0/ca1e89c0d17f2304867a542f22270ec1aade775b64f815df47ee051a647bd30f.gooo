package generator

import (
	"bytes"
	"fmt"
)

func validateMarkerSlotV1(source []byte, region generatedRegion, observed parsedSlot, expected map[string]markerManifestSlotV1, seen map[string]markerManifestSlotV1) (markerManifestSlotV1, error) {
	want, exists := expected[observed.ID]
	if !exists {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 unknown slot ID %q", observed.ID)
	}
	if _, exists := seen[observed.ID]; exists {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 duplicate slot ID %q", observed.ID)
	}
	if observed.RegionID != region.ID || observed.RegionKind != region.Kind || observed.RegionID != want.RegionID || observed.RegionKind != want.RegionKind {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 slot %q has invalid owner", observed.ID)
	}
	if err := validateMarkerBoundsV1(source, observed.Start, observed.End, observed.StartLine, observed.EndLine, false); err != nil {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 slot %q: %w", observed.ID, err)
	}
	if !bytes.Equal(observed.Body, source[observed.Start:observed.End]) {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 slot %q body does not match source bounds", observed.ID)
	}
	if observed.Start < region.Start || observed.End > region.End || observed.End < observed.Start {
		return markerManifestSlotV1{}, fmt.Errorf("generator: marker manifest v1 slot %q is outside region bounds", observed.ID)
	}
	slot := markerManifestSlotV1{ID: observed.ID, RegionID: observed.RegionID, RegionKind: observed.RegionKind, Start: observed.Start, End: observed.End, StartLine: observed.StartLine, EndLine: observed.EndLine}
	seen[observed.ID] = slot
	return slot, nil
}
func validateMarkerBoundsV1(source []byte, start, end, startLine, endLine int, region bool) error {
	if start < 0 || end < start || end > len(source) || startLine < 0 || endLine < startLine {
		return fmt.Errorf("noncanonical byte or line bounds")
	}
	if region {
		if !markerLineStartV1(source, start) || !markerLineEndV1(source, end) || markerLineStartIndexV1(source, start) != startLine || markerLineEndIndexV1(source, end) != endLine {
			return fmt.Errorf("noncanonical generated-region boundary")
		}
		if end == start {
			return fmt.Errorf("empty generated region")
		}
		return nil
	}
	if !markerLineEndV1(source, start) || !markerLineStartV1(source, end) || markerLineEndIndexV1(source, start) != startLine || markerLineStartIndexV1(source, end) != endLine {
		return fmt.Errorf("noncanonical slot boundary")
	}
	return nil
}
func validateMarkerRegionOrderV1(regions []markerManifestRegionV1) error {
	for index := 1; index < len(regions); index++ {
		if regions[index-1].End > regions[index].Start {
			return fmt.Errorf("generator: marker manifest v1 generated regions overlap")
		}
	}
	return nil
}
func validateMarkerSlotOrderV1(region markerManifestRegionV1) error {
	for index := 1; index < len(region.Slots); index++ {
		if region.Slots[index-1].End > region.Slots[index].Start {
			return fmt.Errorf("generator: marker manifest v1 slots overlap in region %q", region.ID)
		}
	}
	return nil
}
