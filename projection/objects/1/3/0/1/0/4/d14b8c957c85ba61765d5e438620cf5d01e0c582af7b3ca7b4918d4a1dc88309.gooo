package generator

import (
	"fmt"
)

func validateSupportedEntityFields(ir SemanticIR) error {
	used := make(map[string]string, len(ir.Entities)+len(ir.Activities))
	for _, entity := range ir.Entities {
		if previous, exists := used[entity.ID]; exists {
			return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", entity.ID, previous))
		}
		used[entity.ID] = "entity"
	}
	for _, activity := range ir.Activities {
		if previous, exists := used[activity.ID]; exists {
			return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", activity.ID, previous))
		}
		used[activity.ID] = "activity"
		for _, slot := range activity.Slots {
			if previous, exists := used[slot.ID]; exists {
				return entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("identity %q is already used by %s", slot.ID, previous))
			}
			used[slot.ID] = "slot"
		}
	}

	for _, entity := range ir.Entities {
		if len(entity.Fields) == 0 {
			continue
		}
		seenNames := make(map[string]string, len(entity.Fields))
		var sourceURI string
		var previousStart int
		var hasPrevious bool
		for index, field := range entity.Fields {
			if err := validateSupportedField(entity, index, field, used, seenNames, sourceURI, previousStart, hasPrevious); err != nil {
				return err
			}
			if sourceURI == "" {
				sourceURI = field.Source.URI
			}
			previousStart = field.Source.Start.Offset
			hasPrevious = true
		}
	}
	return nil
}
