package workfrontier

import (
	"sort"
)

// Select computes a maximal compatible frontier. It never guesses around an
// incomplete registry, state, digest, or path declaration.
func Select(input Input) Result {
	input = normalizeInput(input)
	result := Result{Status: DecisionPass}
	if !inputShapeKnown(input) {
		return unknownAll(input)
	}
	indexes := buildIndexes(input)
	if indexes.invalid || duplicatePathIDs(input.Paths) {
		return unknownAll(input)
	}
	ready := make([]RepairPath, 0, len(input.Paths))
	for _, path := range input.Paths {
		switch classifyPath(input, indexes, path) {
		case pathReady:
			ready = append(ready, path)
		case pathUnknown:
			result.Unknown = append(result.Unknown, path.stableID())
		case pathBlocked:
			result.Blocked = append(result.Blocked, path.stableID())
		case pathShortfall:
			result.Shortfall = append(result.Shortfall, path.stableID())
		}
	}
	selected := make([]RepairPath, 0, len(ready))
	sort.Slice(ready, func(i, j int) bool { return selectionKey(ready[i]) < selectionKey(ready[j]) })
	var usedCPU uint64
	for _, path := range ready {
		if conflictsWithAnyPath(path, selected) || path.CPUCoreNSUpperBound > input.Capacity.CPUCoreNS-usedCPU {
			result.Blocked = append(result.Blocked, path.stableID())
			continue
		}
		selected = append(selected, path)
		usedCPU += path.CPUCoreNSUpperBound
		workID := WorkIDFor(input, path)
		result.Selected = append(result.Selected, workID)
		result.SelectedIDs = append(result.SelectedIDs, path.stableID())
		result.WorkIDs = append(result.WorkIDs, workID)
	}
	if len(result.Unknown) != 0 || len(result.Shortfall) != 0 {
		result.Status = DecisionUnknown
		result.Quality = "UNKNOWN"
		result.FullSuiteRequired = true
		result.Selected = nil
		result.SelectedIDs = nil
		result.WorkIDs = nil
	} else if len(result.Selected) == 0 && len(result.Blocked) != 0 {
		result.Status = DecisionBlocked
		result.Quality = "BLOCKED"
	} else {
		result.Quality = "MAXIMAL"
	}
	return normalizeResult(result)
}

// Run is an alias for Select for callers that model the selector as a pure run.
func Run(input Input) Result { return Select(input) }

// Evaluate is an alias for Select.
func Evaluate(input Input) Result { return Select(input) }

// SelectWork is an alias for Select.
func SelectWork(input Input) Result { return Select(input) }
