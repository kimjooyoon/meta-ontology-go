package resourcevector

func validatePaths(input Input, commands map[string]CommandRecord) validationFailure {
	paths := map[string]PathRecord{}
	pathNames, recordIDs := map[string]struct{}{}, map[string]struct{}{}
	byCommand := map[string]int{}
	for _, path := range input.Paths {
		canonical, failure := validatePath(input.Root, path, commands, paths, pathNames, recordIDs)
		if failure.reason != "" {
			return failure
		}
		paths[path.ID], pathNames[canonical] = path, struct{}{}
		byCommand[path.CommandID]++
	}
	for id := range commands {
		if byCommand[id] == 0 {
			return validationFailure{DecisionUnknown, ReasonMissingPROV}
		}
	}
	return validationFailure{}
}
func validatePath(root string, path PathRecord, commands map[string]CommandRecord, paths map[string]PathRecord, names, records map[string]struct{}) (string, validationFailure) {
	if !validToken(path.ID) || !validToken(path.Path) {
		return "", validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	canonical, ok := canonicalRelativePath(root, path.Path)
	if !ok {
		return "", validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	if _, exists := paths[path.ID]; exists {
		return "", validationFailure{DecisionFailClosed, ReasonDuplicatePath}
	}
	if _, exists := names[canonical]; exists {
		return "", validationFailure{DecisionFailClosed, ReasonDuplicatePath}
	}
	if _, exists := commands[path.CommandID]; !exists {
		return "", validationFailure{DecisionFailClosed, ReasonDanglingID}
	}
	if failure := validatePathFields(path); failure.reason != "" {
		return "", failure
	}
	if failure := validateRecordIDs(path.RecordIDs, records); failure.reason != "" {
		return "", failure
	}
	return canonical, validationFailure{}
}
func validatePathFields(path PathRecord) validationFailure {
	if path.RecordIDs == nil || len(path.RecordIDs) == 0 || path.Finite == nil ||
		path.ClosureNumerator == nil || path.ClosureDenominator == nil {
		return validationFailure{DecisionUnknown, ReasonMissingPROV}
	}
	if *path.ClosureNumerator > *path.ClosureDenominator {
		return validationFailure{DecisionFailClosed, ReasonClosureInvalid}
	}
	return validationFailure{}
}
func validateRecordIDs(values []string, global map[string]struct{}) validationFailure {
	local := map[string]struct{}{}
	for _, recordID := range values {
		if !validToken(recordID) {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		if _, exists := local[recordID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
		}
		if _, exists := global[recordID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
		}
		local[recordID], global[recordID] = struct{}{}, struct{}{}
	}
	return validationFailure{}
}
