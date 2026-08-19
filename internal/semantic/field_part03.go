package semantic

import (
	"fmt"
)

func normalizeFields(raw []Field, parent ID, kind Kind) ([]Field, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if kind != Entity {
		return nil, fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
	}
	fields := make([]Field, 0, len(raw))
	fieldIDs := make(map[ID]struct{}, len(raw))
	nameOwners := make(map[string]ID, len(raw)*2)
	for _, field := range raw {
		normalized, err := field.Normalized()
		if err != nil {
			return nil, err
		}
		if normalized.Parent != parent {
			return nil, fmt.Errorf("%w: field %s parent is %s, want %s", ErrInvalidField, normalized.ID, normalized.Parent, parent)
		}
		if _, exists := fieldIDs[normalized.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate field ID %s", ErrInvalidField, normalized.ID)
		}
		fieldIDs[normalized.ID] = struct{}{}
		for _, name := range append([]string{normalized.Name}, normalized.Aliases...) {
			if owner, exists := nameOwners[name]; exists && owner != normalized.ID {
				return nil, fmt.Errorf("%w: field name %q is shared by %s and %s", ErrNameCollision, name, owner, normalized.ID)
			}
			nameOwners[name] = normalized.ID
		}
		fields = append(fields, normalized)
	}
	return fields, nil
}
func copyFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	cloned := make([]Field, len(fields))
	for i, field := range fields {
		cloned[i] = field
		cloned[i].Aliases = append([]string(nil), field.Aliases...)
	}
	return cloned
}
