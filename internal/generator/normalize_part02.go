package generator

import (
	"fmt"
	"sort"
)

func normalizeImports(ir *SemanticIR) error {
	seenPaths := make(map[string]struct{}, len(ir.Imports))
	for index := range ir.Imports {
		item := &ir.Imports[index]
		if item.Path == "" {
			return fmt.Errorf("generator: import %d has an empty path", index)
		}
		if item.Name != "" && !isGoIdentifier(item.Name) {
			return fmt.Errorf("generator: invalid import name %q", item.Name)
		}
		if _, exists := seenPaths[item.Path]; exists {
			return fmt.Errorf("generator: duplicate import path %q", item.Path)
		}
		seenPaths[item.Path] = struct{}{}
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
