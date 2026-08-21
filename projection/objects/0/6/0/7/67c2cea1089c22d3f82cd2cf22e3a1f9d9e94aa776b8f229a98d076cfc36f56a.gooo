package generator

import (
	"fmt"
)

func expectedMarkerManifestOwnersV1(ir SemanticIR) (map[string]string, map[string]markerManifestSlotV1, error) {
	expectedRegions := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	expectedSlots := make(map[string]markerManifestSlotV1)
	for _, entity := range ir.Entities {
		if err := addMarkerManifestRegionOwnerV1(expectedRegions, entity.ID, "entity"); err != nil {
			return nil, nil, err
		}
	}
	for _, activity := range ir.Activities {
		if err := addMarkerManifestRegionOwnerV1(expectedRegions, activity.ID, "activity"); err != nil {
			return nil, nil, err
		}
		for _, slot := range activity.Slots {
			if slot.ID == "" {
				return nil, nil, fmt.Errorf("generator: marker manifest v1 declared slot has empty ID")
			}
			if _, exists := expectedSlots[slot.ID]; exists {
				return nil, nil, fmt.Errorf("generator: marker manifest v1 duplicate declared slot ID %q", slot.ID)
			}
			expectedSlots[slot.ID] = markerManifestSlotV1{ID: slot.ID, RegionID: activity.ID, RegionKind: "activity"}
		}
	}
	return expectedRegions, expectedSlots, nil
}
func addMarkerManifestRegionOwnerV1(owners map[string]string, id, kind string) error {
	if id == "" {
		return fmt.Errorf("generator: marker manifest v1 declared %s has empty ID", kind)
	}
	if _, exists := owners[id]; exists {
		return fmt.Errorf("generator: marker manifest v1 duplicate declared region ID %q", id)
	}
	owners[id] = kind
	return nil
}
func validateMarkerRegionV1(source []byte, observed generatedRegion, expected map[string]string, seen map[string]struct{}) error {
	if observed.ID == "" {
		return fmt.Errorf("generator: marker manifest v1 region has empty ID")
	}
	if _, exists := seen[observed.ID]; exists {
		return fmt.Errorf("generator: marker manifest v1 duplicate region ID %q", observed.ID)
	}
	wantKind, exists := expected[observed.ID]
	if !exists {
		return fmt.Errorf("generator: marker manifest v1 unknown region ID %q", observed.ID)
	}
	if observed.Kind != wantKind {
		return fmt.Errorf("generator: marker manifest v1 region %q has kind %q, want %q", observed.ID, observed.Kind, wantKind)
	}
	if err := validateMarkerBoundsV1(source, observed.Start, observed.End, observed.StartLine, observed.EndLine, true); err != nil {
		return fmt.Errorf("generator: marker manifest v1 region %q: %w", observed.ID, err)
	}
	seen[observed.ID] = struct{}{}
	return nil
}
