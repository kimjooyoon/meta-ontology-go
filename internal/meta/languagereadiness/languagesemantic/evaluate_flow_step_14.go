package languagesemantic

import "path/filepath"

func evaluateFlowStep13(flow *evaluateFlowState) {
	flow.slot09 = semanticSourcePaths(flow.slot05)
}

func semanticSourcePaths(receipt syntaxReceipt) []string {
	kinds := make(map[string]string, len(receipt.Cases))
	for _, item := range receipt.Cases {
		path := filepath.ToSlash(filepath.Clean(item.Definition.Path))
		kinds[path] = item.Definition.Kind
	}
	paths := make([]string, 0, len(receipt.Source.GoooFiles))
	for _, file := range receipt.Source.GoooFiles {
		path := filepath.ToSlash(filepath.Clean(file.Path))
		kind, registered := kinds[path]
		if !registered || kind == "VALID" {
			paths = append(paths, path)
		}
	}
	return paths
}
