package generator

import (
	"fmt"
	"sort"
)

func normalizeIR(input SemanticIR) (SemanticIR, error) {
	if input.Package == "" {
		return SemanticIR{}, fmt.Errorf("generator: package is required")
	}
	if !isGoIdentifier(input.Package) {
		return SemanticIR{}, fmt.Errorf("generator: invalid Go package %q", input.Package)
	}
	result := copyIR(input)
	if err := normalizeImports(&result); err != nil {
		return SemanticIR{}, err
	}
	types, err := normalizeEntities(&result)
	if err != nil {
		return SemanticIR{}, err
	}
	if err := normalizeActivities(&result, types); err != nil {
		return SemanticIR{}, err
	}
	return result, nil
}

func copyIR(input SemanticIR) SemanticIR {
	result := input
	result.Imports = append([]Import(nil), input.Imports...)
	result.Entities = append([]Entity(nil), input.Entities...)
	result.Activities = append([]Activity(nil), input.Activities...)
	for index := range result.Entities {
		result.Entities[index].Fields = append([]Field(nil), input.Entities[index].Fields...)
	}
	for index := range result.Activities {
		result.Activities[index].Inputs = append([]Port(nil), input.Activities[index].Inputs...)
		result.Activities[index].Outputs = append([]Port(nil), input.Activities[index].Outputs...)
		result.Activities[index].Slots = append([]Slot(nil), input.Activities[index].Slots...)
	}
	return result
}

func normalizeImports(ir *SemanticIR) error {
	seen := make(map[string]struct{}, len(ir.Imports))
	for index := range ir.Imports {
		item := &ir.Imports[index]
		if item.Path == "" {
			return fmt.Errorf("generator: import %d has an empty path", index)
		}
		if item.Name != "" && !isGoIdentifier(item.Name) {
			return fmt.Errorf("generator: invalid import name %q", item.Name)
		}
		key := item.Name + "\x00" + item.Path
		if _, exists := seen[key]; exists {
			return fmt.Errorf("generator: duplicate import %q", item.Path)
		}
		seen[key] = struct{}{}
	}
	sort.Slice(ir.Imports, func(i, j int) bool {
		if ir.Imports[i].Path != ir.Imports[j].Path {
			return ir.Imports[i].Path < ir.Imports[j].Path
		}
		return ir.Imports[i].Name < ir.Imports[j].Name
	})
	return nil
}

func normalizeEntities(ir *SemanticIR) (map[string]string, error) {
	types := make(map[string]string, len(ir.Entities))
	names := make(map[string]struct{}, len(ir.Entities))
	for index := range ir.Entities {
		if err := normalizeEntity(&ir.Entities[index], index, types, names); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(ir.Entities, func(i, j int) bool {
		if ir.Entities[i].ID != ir.Entities[j].ID {
			return ir.Entities[i].ID < ir.Entities[j].ID
		}
		return ir.Entities[i].GoName < ir.Entities[j].GoName
	})
	return types, nil
}

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
		activity.Slots = []Slot{{ID: activity.ID + "/implementation", Name: "implementation", Default: defaultActivityBody(activity)}}
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
