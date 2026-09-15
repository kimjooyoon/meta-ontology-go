package main

func mergeImportSet(base map[string]struct{}, extra map[string]struct{}) map[string]struct{} {
	merged := make(map[string]struct{}, len(base)+len(extra))
	for value := range base {
		merged[value] = struct{}{}
	}
	for value := range extra {
		merged[value] = struct{}{}
	}
	return merged
}
func estimatedLines(preamble []byte, packageName string, chunk declChunk, allImports map[string]importSpec) int {
	formatted, err := renderChunk(preamble, packageName, chunk, allImports)
	if err != nil {
		return int(^uint(0) >> 1)
	}
	return lineCount(formatted)
}
