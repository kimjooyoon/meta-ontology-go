package generator

import (
	"fmt"
)

func validateStableIDs(ir SemanticIR) error {
	seen := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		if err := recordStableID(seen, entity.ID, "entity"); err != nil {
			return err
		}
		for _, field := range entity.Fields {
			if err := recordStableID(seen, field.ID, "field"); err != nil {
				return err
			}
		}
	}
	for _, activity := range ir.Activities {
		if err := recordStableID(seen, activity.ID, "activity"); err != nil {
			return err
		}
		for _, slot := range activity.Slots {
			if err := recordStableID(seen, slot.ID, "slot"); err != nil {
				return err
			}
		}
	}
	return nil
}
func recordStableID(seen map[string]string, id, kind string) error {
	if previous, exists := seen[id]; exists {
		return fmt.Errorf("generator: stable ID %q is used by %s and %s", id, previous, kind)
	}
	seen[id] = kind
	return nil
}
