package languagesemantic

import (
	"path/filepath"
)

func sourceDefinitions(registry Registry) []string {
	paths := make([]string, 0, expectedSources)
	for _, definition := range registry.Cases {
		if definition.Kind == CaseSource {
			paths = append(paths, filepath.ToSlash(filepath.Clean(definition.Path)))
		}
	}
	return paths
}
