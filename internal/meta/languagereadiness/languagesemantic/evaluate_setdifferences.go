package languagesemantic

import (
	"path/filepath"
)

func setDifferences(observed, registered []string) ([]string, []string) {
	observedSet, registeredSet := make(map[string]struct{}), make(map[string]struct{})
	for _, value := range observed {
		observedSet[filepath.ToSlash(filepath.Clean(value))] = struct{}{}
	}
	for _, value := range registered {
		registeredSet[filepath.ToSlash(filepath.Clean(value))] = struct{}{}
	}
	var unregistered, missing []string
	for value := range observedSet {
		if _, ok := registeredSet[value]; !ok {
			unregistered = append(unregistered, value)
		}
	}
	for value := range registeredSet {
		if _, ok := observedSet[value]; !ok {
			missing = append(missing, value)
		}
	}
	return unregistered, missing
}
