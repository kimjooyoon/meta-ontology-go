package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"sort"
	"unicode"
	"unicode/utf8"
)

func canonicalSelector(input workfrontier.Input) workfrontier.Input {
	input.Pressures = append([]workfrontier.Pressure{}, input.Pressures...)
	input.States = append([]workfrontier.ObligationState{}, input.States...)
	input.Paths = append([]workfrontier.RepairPath{}, input.Paths...)
	sort.Slice(input.Pressures, func(left, right int) bool {
		return pressureID(input.Pressures[left]) < pressureID(input.Pressures[right])
	})
	sort.Slice(input.States, func(left, right int) bool {
		return input.States[left].ObligationID < input.States[right].ObligationID
	})
	for index := range input.Paths {
		path := &input.Paths[index]
		path.PrerequisiteObligationIDs = sortedStrings(path.PrerequisiteObligationIDs)
		path.ReadSet, path.WriteSet = sortedStrings(path.ReadSet), sortedStrings(path.WriteSet)
		path.RequiredPressureIDs = sortedStrings(path.RequiredPressureIDs)
	}
	sort.Slice(input.Paths, func(left, right int) bool {
		return pathID(input.Paths[left]) < pathID(input.Paths[right])
	})
	return input
}
func validIDs(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validID(value) {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
func validID(value string) bool {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
func sortedStrings(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}
func pressureID(pressure workfrontier.Pressure) string {
	return stableID(pressure.StableID, pressure.ID)
}
func pathID(path workfrontier.RepairPath) string {
	return stableID(path.StableID, path.ID)
}
func stableID(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}
