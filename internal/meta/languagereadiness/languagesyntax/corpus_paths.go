package languagesyntax

import "strings"

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
	for _, unit := range registry.ProjectionUnits {
		paths = append(paths, unit.Path)
	}
	return paths
}
