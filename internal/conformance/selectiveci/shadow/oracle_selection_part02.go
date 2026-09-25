package shadow

import (
	"encoding/json"
	"sort"
)

func normalizeFiles(values []manifestFile) []manifestFile {
	result := append([]manifestFile(nil), values...)
	for i := range result {
		result[i].SemanticIDs = sortedCopy(result[i].SemanticIDs)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
func normalizeCommands(values []command) []command {
	result := append([]command(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	for i := range result {
		result[i].Argv = append([]string(nil), result[i].Argv...)
	}
	return result
}
func fileIndex(values []manifestFile) map[string]manifestFile {
	result := map[string]manifestFile{}
	for _, value := range values {
		result[value.Path] = value
	}
	return result
}
func equalFiles(left, right []manifestFile) bool {
	return stringJSON(normalizeFiles(left)) == stringJSON(normalizeFiles(right))
}
func stringJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
func equalStrings(left, right []string) bool {
	return stringJSON(sortedCopy(left)) == stringJSON(sortedCopy(right))
}
func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
