package workfrontier

import (
	"sort"
	"strconv"
	"strings"
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

func sortedUnique(values []string) []string {
	values = sortedCopy(values)
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func stateKey(state ObligationState) string {
	return state.obligationID() + "\x00" + state.Status
}

func pathKey(path RepairPath) string {
	parts := []string{
		path.stableID(), path.ObligationID,
		strings.Join(path.PrerequisiteObligationIDs, "\x00"),
		strings.Join(path.ReadSet, "\x00"), strings.Join(path.WriteSet, "\x00"),
		strings.Join(path.RequiredPressureIDs, "\x00"),
		strconv.FormatUint(uint64(path.PolicyPriority), 10),
		strconv.FormatUint(path.CPUCoreNSUpperBound, 10),
	}
	return strings.Join(parts, "\x00")
}
