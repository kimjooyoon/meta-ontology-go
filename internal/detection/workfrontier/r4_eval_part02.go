package workfrontier

import (
	"sort"
)

func classifyR4Paths(legacy Input, graph r4Graph, result R4Result) ([]RepairPath, R4Result) {
	indexes := buildIndexes(legacy)
	ready := make([]RepairPath, 0, len(graph.reachablePaths))
	if indexes.invalid {
		for _, path := range graph.reachablePaths {
			result.Unknown = append(result.Unknown, path.StableID)
		}
		return ready, result
	}
	for _, path := range graph.reachablePaths {
		switch classifyPath(legacy, indexes, path) {
		case pathReady:
			ready = append(ready, path)
		case pathUnknown:
			result.Unknown = append(result.Unknown, path.StableID)
		case pathBlocked:
			result.Blocked = append(result.Blocked, path.StableID)
		case pathShortfall:
			result.Shortfall = append(result.Shortfall, path.StableID)
		}
	}
	return ready, result
}
func selectR4Paths(input R4Input, legacy Input, ready []RepairPath, result R4Result) R4Result {
	sort.Slice(ready, func(i, j int) bool { return selectionKey(ready[i]) < selectionKey(ready[j]) })
	var usedCPU uint64
	selected := make([]RepairPath, 0, len(ready))
	for _, path := range ready {
		conflict := false
		for _, prior := range selected {
			result.WorkReceipt.ConflictChecks++
			if conflicts(path, prior) {
				conflict = true
				break
			}
		}
		if conflict || path.CPUCoreNSUpperBound > input.Capacity.CPUCoreNS-usedCPU {
			result.Blocked = append(result.Blocked, path.StableID)
			continue
		}
		selected = append(selected, path)
		usedCPU += path.CPUCoreNSUpperBound
		workID := WorkIDFor(legacy, path)
		result.Selected = append(result.Selected, workID)
		result.SelectedIDs = append(result.SelectedIDs, path.StableID)
		result.WorkIDs = append(result.WorkIDs, workID)
	}
	return result
}
func finishR4Result(result R4Result) R4Result {
	if len(result.Unknown) != 0 {
		return r4UnknownWithResult(result, R4ReasonRequiredInputMissing)
	}
	if len(result.Shortfall) != 0 {
		return r4UnknownWithResult(result, R4ReasonSelectionShortfall)
	}
	if len(result.Selected) == 0 && len(result.Blocked) != 0 {
		result.Status = R4StatusBlocked
		result.Quality = R4StatusBlocked
	} else {
		result.Status = R4StatusPass
		result.Quality = "MAXIMAL"
	}
	result.Reason = R4ReasonNone
	return normalizeR4Result(result)
}
