package workfrontier

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
