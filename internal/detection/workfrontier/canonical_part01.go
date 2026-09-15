package workfrontier

import (
	"sort"
)

func normalizeInput(input Input) Input {
	input.Pressures = append(make([]Pressure, 0, len(input.Pressures)), input.Pressures...)
	input.States = append(make([]ObligationState, 0, len(input.States)), input.States...)
	input.Paths = append(make([]RepairPath, 0, len(input.Paths)), input.Paths...)
	for i := range input.Pressures {
		if input.Pressures[i].StableID == "" {
			input.Pressures[i].StableID = input.Pressures[i].ID
		}
	}
	for i := range input.States {
		if input.States[i].ObligationID == "" {
			input.States[i].ObligationID = input.States[i].ID
		}
	}
	for i := range input.Paths {
		path := &input.Paths[i]
		if path.StableID == "" {
			path.StableID = path.ID
		}
		path.PrerequisiteObligationIDs = sortedCopy(path.PrerequisiteObligationIDs)
		path.ReadSet = sortedCopy(path.ReadSet)
		path.WriteSet = sortedCopy(path.WriteSet)
		path.RequiredPressureIDs = sortedCopy(path.RequiredPressureIDs)
	}
	sort.Slice(input.Pressures, func(i, j int) bool {
		return input.Pressures[i].stableID() < input.Pressures[j].stableID()
	})
	sort.Slice(input.States, func(i, j int) bool {
		left, right := input.States[i], input.States[j]
		return stateKey(left) < stateKey(right)
	})
	sort.Slice(input.Paths, func(i, j int) bool {
		return pathKey(input.Paths[i]) < pathKey(input.Paths[j])
	})
	return input
}
func normalizeResult(result Result) Result {
	result.Selected = uniqueInOrder(result.Selected)
	result.SelectedIDs = uniqueInOrder(result.SelectedIDs)
	result.WorkIDs = uniqueInOrder(result.WorkIDs)
	result.Unknown = sortedUnique(result.Unknown)
	result.Blocked = sortedUnique(result.Blocked)
	result.Shortfall = sortedUnique(result.Shortfall)
	return result
}
func uniqueInOrder(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
func sortedCopy(values []string) []string {
	copyOfValues := append(make([]string, 0, len(values)), values...)
	sort.Strings(copyOfValues)
	return copyOfValues
}
