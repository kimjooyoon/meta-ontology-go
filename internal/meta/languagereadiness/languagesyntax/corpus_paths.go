package languagesyntax

import "strings"

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
	return paths
}
