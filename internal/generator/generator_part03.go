package generator

import (
	"fmt"
	"go/parser"
	"go/token"
)

func validateDeclaredSlots(ir SemanticIR, markers parsedMarkers, allowRemovedRegions bool) error {
	declared := make(map[string]struct{})
	active := make(map[string]struct{})
	for _, activity := range ir.Activities {
		for _, slot := range activity.Slots {
			declared[slot.ID] = struct{}{}
		}
	}
	for _, region := range markers.Regions {
		if allowRemovedRegions && !hasActivity(ir, region.ID) {
			continue
		}
		for _, slot := range region.Slots {
			active[slot.ID] = struct{}{}
		}
	}
	for id := range active {
		if _, exists := declared[id]; !exists {
			return fmt.Errorf("generator: stale slot identity %q", id)
		}
	}
	for _, slot := range markers.Slots {
		owner, declared := declaredSlotOwner(ir, slot.ID)
		if !declared {
			continue
		}
		if owner != slot.RegionID {
			return fmt.Errorf("generator: slot %q changes region owner from %q to %q", slot.ID, slot.RegionID, owner)
		}
		if slot.RegionKind != "activity" {
			return fmt.Errorf("generator: slot %q belongs to non-activity region kind %q", slot.ID, slot.RegionKind)
		}
	}
	return nil
}
func declaredSlotOwner(ir SemanticIR, slotID string) (string, bool) {
	for _, activity := range ir.Activities {
		for _, slot := range activity.Slots {
			if slot.ID == slotID {
				return activity.ID, true
			}
		}
	}
	return "", false
}
func hasActivity(ir SemanticIR, id string) bool {
	for _, activity := range ir.Activities {
		if activity.ID == id {
			return true
		}
	}
	return false
}
func validatePackage(source []byte, expected string) error {
	file, err := parser.ParseFile(token.NewFileSet(), "previous.go", source, parser.PackageClauseOnly)
	if err != nil {
		return fmt.Errorf("generator: previous source has no readable package clause: %w", err)
	}
	if file.Name.Name != expected {
		return fmt.Errorf("generator: previous package %q does not match semantic package %q", file.Name.Name, expected)
	}
	return nil
}
