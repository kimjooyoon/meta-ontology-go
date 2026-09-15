package workfrontier

import "sort"

// FairBaseline is a deliberately simpler, non-authorizing baseline. It uses
// stable path order and the same declared resource/conflict constraints, but
// does not claim a stronger frontier quality or bypass the R4 evaluator.
func FairBaseline(input R4Input) []string {
	input = normalizeR4Input(input)
	graph, reason := buildR4Graph(input)
	if reason != "" {
		return nil
	}
	legacy := Input{
		SchemaVersion:            SchemaVersion,
		SnapshotDigest:           input.SnapshotDigest,
		PolicyDigest:             input.PolicyDigest,
		RegistryDigest:           input.RegistryDigest,
		MinimumSelectedPressures: input.MinimumSelectedPressures,
		Capacity:                 input.Capacity,
		Pressures:                append([]Pressure(nil), input.Pressures...),
		States:                   append([]ObligationState(nil), input.States...),
		Paths:                    append([]RepairPath(nil), graph.reachablePaths...),
	}
	indexes := buildIndexes(legacy)
	if indexes.invalid {
		return nil
	}
	paths := append([]RepairPath(nil), graph.reachablePaths...)
	sort.Slice(paths, func(i, j int) bool { return paths[i].StableID < paths[j].StableID })
	selected := make([]RepairPath, 0, len(paths))
	var usedCPU uint64
	for _, path := range paths {
		if classifyPath(legacy, indexes, path) != pathReady || path.CPUCoreNSUpperBound > input.Capacity.CPUCoreNS-usedCPU {
			continue
		}
		conflict := false
		for _, prior := range selected {
			if conflicts(path, prior) {
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		selected = append(selected, path)
		usedCPU += path.CPUCoreNSUpperBound
	}
	result := make([]string, 0, len(selected))
	for _, path := range selected {
		result = append(result, path.StableID)
	}
	return result
}
