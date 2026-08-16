package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type pathClass uint8

const (
	pathReady pathClass = iota
	pathUnknown
	pathBlocked
	pathShortfall
)

type frontierIndexes struct {
	pressures map[string]struct{}
	states    map[string]string
	invalid   bool
}

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

// WorkID returns SHA-256(snapshot || obligation || path || policy) as hex.
func WorkID(snapshotDigest, obligationID, pathID, policyDigest string) string {
	digest := sha256.Sum256([]byte(snapshotDigest + obligationID + pathID + policyDigest))
	return hex.EncodeToString(digest[:])
}

// WorkIDFor computes the identity of a path in an input snapshot.
func WorkIDFor(input Input, path RepairPath) string {
	if path.WorkID != "" {
		return path.WorkID
	}
	return WorkID(input.SnapshotDigest, path.ObligationID, path.stableID(), input.PolicyDigest)
}

func inputShapeKnown(input Input) bool {
	if input.SchemaVersion != SchemaVersion || input.SnapshotDigest == "" ||
		input.PolicyDigest == "" || input.RegistryDigest == "" ||
		input.MinimumSelectedPressures < 2 {
		return false
	}
	if input.fromJSON {
		p := input.present
		return p.schemaVersion && p.snapshotDigest && p.policyDigest && p.registryDigest &&
			p.minimumSelectedPressures && p.capacity && p.pressures && p.states && p.paths &&
			input.Capacity.cpuCoreNSPresent
	}
	return true
}

func buildIndexes(input Input) frontierIndexes {
	indexes := frontierIndexes{pressures: make(map[string]struct{}), states: make(map[string]string)}
	for _, pressure := range input.Pressures {
		id := pressure.stableID()
		if !pressureKnown(pressure) || id == "" {
			indexes.invalid = true
			continue
		}
		if _, exists := indexes.pressures[id]; exists {
			indexes.invalid = true
		}
		indexes.pressures[id] = struct{}{}
	}
	for _, state := range input.States {
		id := state.obligationID()
		if !stateKnown(state) || id == "" {
			indexes.invalid = true
			continue
		}
		if _, exists := indexes.states[id]; exists {
			indexes.invalid = true
		}
		indexes.states[id] = state.Status
	}
	return indexes
}

func pressureKnown(pressure Pressure) bool {
	return pressure.stableID() != "" && (!pressure.fromJSON || pressure.stableIDPresent)
}

func stateKnown(state ObligationState) bool {
	if state.obligationID() == "" || state.Status == "" {
		return false
	}
	return !state.fromJSON || (state.obligationIDPresent && state.statusPresent)
}

func pathKnown(path RepairPath) bool {
	if path.stableID() == "" || path.ObligationID == "" || path.CPUCoreNSUpperBound == 0 {
		return false
	}
	if path.fromJSON && (!path.stableIDPresent || !path.obligationIDPresent ||
		!path.prerequisiteObligationIDsPresent || !path.readSetPresent ||
		!path.writeSetPresent || !path.requiredPressureIDsPresent ||
		!path.policyPriorityPresent || !path.cpuCoreNSUpperBoundPresent) {
		return false
	}
	return (len(path.ReadSet) != 0 || len(path.WriteSet) != 0) &&
		stringsKnown(path.PrerequisiteObligationIDs) && stringsKnown(path.ReadSet) &&
		stringsKnown(path.WriteSet) && stringsKnown(path.RequiredPressureIDs) &&
		!hasDuplicate(path.PrerequisiteObligationIDs) && !hasDuplicate(path.ReadSet) &&
		!hasDuplicate(path.WriteSet) && !hasDuplicate(path.RequiredPressureIDs)
}

func classifyPath(input Input, indexes frontierIndexes, path RepairPath) pathClass {
	if !pathKnown(path) {
		return pathUnknown
	}
	state, exists := indexes.states[path.ObligationID]
	if !exists || !pressuresResolve(indexes, path.RequiredPressureIDs) ||
		!obligationsResolve(indexes, path.PrerequisiteObligationIDs) ||
		!pressuresResolve(indexes, path.ReadSet) || !pressuresResolve(indexes, path.WriteSet) {
		return pathUnknown
	}
	if state == "PASS" || prerequisitesIncomplete(indexes, path.PrerequisiteObligationIDs) ||
		path.CPUCoreNSUpperBound > input.Capacity.CPUCoreNS {
		return pathBlocked
	}
	if uint64(len(path.RequiredPressureIDs)) < uint64(input.MinimumSelectedPressures) {
		return pathShortfall
	}
	return pathReady
}

func pressuresResolve(indexes frontierIndexes, ids []string) bool {
	for _, id := range ids {
		if _, ok := indexes.pressures[id]; !ok {
			return false
		}
	}
	return true
}

func obligationsResolve(indexes frontierIndexes, ids []string) bool {
	for _, id := range ids {
		if _, ok := indexes.states[id]; !ok {
			return false
		}
	}
	return true
}

func prerequisitesIncomplete(indexes frontierIndexes, ids []string) bool {
	for _, id := range ids {
		if indexes.states[id] != "PASS" {
			return true
		}
	}
	return false
}

func stringsKnown(values []string) bool {
	for _, value := range values {
		if value == "" {
			return false
		}
	}
	return true
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func conflictsWithAnyPath(path RepairPath, selected []RepairPath) bool {
	for _, other := range selected {
		if conflicts(path, other) {
			return true
		}
	}
	return false
}
