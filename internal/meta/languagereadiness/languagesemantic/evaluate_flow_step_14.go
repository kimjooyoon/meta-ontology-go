package languagesemantic

import "path/filepath"

func evaluateFlowStep13(flow *evaluateFlowState) {
	flow.slot09 = semanticSourcePaths(flow.slot01, flow.slot05)
}

func semanticSourcePaths(registry Registry, receipt syntaxReceipt) []string {
	present := make(map[string]bool, len(receipt.Source.GoooFiles))
	for _, file := range receipt.Source.GoooFiles {
		path := filepath.ToSlash(filepath.Clean(file.Path))
		present[path] = true
	}
	packageMembers := map[string]bool{}
	for _, unit := range receipt.Source.PackageUnits {
		for _, member := range unit.Members {
			packageMembers[filepath.ToSlash(filepath.Clean(member))] = true
		}
	}
	paths := make([]string, 0, expectedSources)
	for _, definition := range registry.Cases {
		if definition.Kind != CaseSource {
			continue
		}
		path := filepath.ToSlash(filepath.Clean(definition.Path))
		if !packageMembers[path] && present[path] {
			paths = append(paths, path)
		}
	}
	return paths
}
