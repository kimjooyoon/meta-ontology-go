package languagesyntax

import (
	"slices"
	"strings"
)

// DeclaredSourceCount counts distinct registered paths before observing files.
// It is an inventory boundary, not a language capability denominator.
func DeclaredSourceCount() int {
	paths := registryPaths(expectedRegistry())
	declared := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		declared[path] = struct{}{}
	}
	return len(declared)
}

func registryPaths(registry Registry) []string {
	paths := []string{}
	for _, definition := range registry.Cases {
		if strings.HasSuffix(definition.Path, ".gooo") {
			paths = append(paths, definition.Path)
		}
	}
	for _, unit := range registry.PackageUnits {
		paths = append(paths, unit.Members...)
	}
	paths = append(paths, registry.MetaSources...)
	for _, group := range registry.CaseSources {
		for _, member := range group.Members {
			if !slices.Contains(paths, member) {
				paths = append(paths, member)
			}
		}
	}
	for _, unit := range registry.ProjectionUnits {
		paths = append(paths, unit.Path)
	}
	return paths
}
