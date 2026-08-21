package generator

import (
	"fmt"
)

func normalizeActivity(activity *Activity, index int, entityTypes map[string]string, ids, names map[string]struct{}) error {
	if activity.ID == "" {
		return fmt.Errorf("generator: activity %d has no semantic ID", index)
	}
	if activity.Name == "" {
		return fmt.Errorf("generator: activity %q has no name", activity.ID)
	}
	if activity.GoName == "" {
		activity.GoName = activity.Name
	}
	if !isGoIdentifier(activity.GoName) {
		return fmt.Errorf("generator: activity %q has invalid Go name %q", activity.ID, activity.GoName)
	}
	if _, exists := ids[activity.ID]; exists {
		return fmt.Errorf("generator: duplicate activity ID %q", activity.ID)
	}
	if _, exists := names[activity.GoName]; exists {
		return fmt.Errorf("generator: duplicate activity Go name %q", activity.GoName)
	}
	ids[activity.ID] = struct{}{}
	names[activity.GoName] = struct{}{}
	if err := normalizePorts(activity, entityTypes); err != nil {
		return err
	}
	if len(activity.Slots) == 0 {
		activity.Slots = []Slot{{ID: activity.ID + "/implementation", Name: "implementation", Default: defaultActivityBody(activity, entityTypes)}}
	}
	return normalizeSlots(activity)
}
func normalizeSlots(activity *Activity) error {
	seen := make(map[string]struct{}, len(activity.Slots))
	for index := range activity.Slots {
		slot := &activity.Slots[index]
		if slot.ID == "" {
			return fmt.Errorf("generator: activity %q slot %d has no stable ID", activity.ID, index)
		}
		if _, exists := seen[slot.ID]; exists {
			return fmt.Errorf("generator: activity %q has duplicate slot ID %q", activity.ID, slot.ID)
		}
		seen[slot.ID] = struct{}{}
	}
	return nil
}
func validateTopLevelNames(ir SemanticIR) error {
	names := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		if previous, exists := names[entity.GoName]; exists {
			return fmt.Errorf("generator: Go name %q is used by %s and entity %q", entity.GoName, previous, entity.ID)
		}
		names[entity.GoName] = "entity " + entity.ID
	}
	for _, activity := range ir.Activities {
		if previous, exists := names[activity.GoName]; exists {
			return fmt.Errorf("generator: Go name %q is used by %s and activity %q", activity.GoName, previous, activity.ID)
		}
		names[activity.GoName] = "activity " + activity.ID
	}
	return nil
}
