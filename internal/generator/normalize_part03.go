package generator

import (
	"fmt"
	"sort"
)

func normalizeEntity(entity *Entity, index int, types map[string]string, names map[string]struct{}) error {
	if entity.ID == "" {
		return fmt.Errorf("generator: entity %d has no semantic ID", index)
	}
	if entity.Name == "" {
		return fmt.Errorf("generator: entity %q has no name", entity.ID)
	}
	if entity.GoName == "" {
		entity.GoName = entity.Name
	}
	if !isGoIdentifier(entity.GoName) {
		return fmt.Errorf("generator: entity %q has invalid Go name %q", entity.ID, entity.GoName)
	}
	if _, exists := types[entity.ID]; exists {
		return fmt.Errorf("generator: duplicate entity ID %q", entity.ID)
	}
	if _, exists := names[entity.GoName]; exists {
		return fmt.Errorf("generator: duplicate entity Go name %q", entity.GoName)
	}
	types[entity.ID] = entity.GoName
	names[entity.GoName] = struct{}{}
	return normalizeFields(entity)
}
func normalizeFields(entity *Entity) error {
	seen := make(map[string]struct{}, len(entity.Fields))
	for index := range entity.Fields {
		field := &entity.Fields[index]
		if field.Name == "" && field.GoName == "" {
			return fmt.Errorf("generator: entity %q field %d has no name", entity.ID, index)
		}
		if field.GoName == "" {
			field.GoName = field.Name
		}
		if !isGoIdentifier(field.GoName) {
			return fmt.Errorf("generator: entity %q has invalid field name %q", entity.ID, field.GoName)
		}
		if field.GoType == "" {
			return fmt.Errorf("generator: entity %q field %q has no Go type", entity.ID, field.GoName)
		}
		if _, exists := seen[field.GoName]; exists {
			return fmt.Errorf("generator: entity %q has duplicate field %q", entity.ID, field.GoName)
		}
		seen[field.GoName] = struct{}{}
	}
	return nil
}
func normalizeActivities(ir *SemanticIR, entityTypes map[string]string) error {
	names := make(map[string]struct{}, len(ir.Activities))
	ids := make(map[string]struct{}, len(ir.Activities))
	for index := range ir.Activities {
		if err := normalizeActivity(&ir.Activities[index], index, entityTypes, ids, names); err != nil {
			return err
		}
	}
	sort.SliceStable(ir.Activities, func(i, j int) bool {
		if ir.Activities[i].ID != ir.Activities[j].ID {
			return ir.Activities[i].ID < ir.Activities[j].ID
		}
		return ir.Activities[i].GoName < ir.Activities[j].GoName
	})
	return nil
}
