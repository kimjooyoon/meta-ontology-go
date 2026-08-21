package resourcevector

func baselineAffected(input Input, commands map[string]CommandRecord, ids []string) (uint64, validationFailure) {
	if input.AffectedStableIDs == nil {
		return 0, validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
	}
	registry := map[string]struct{}{}
	for _, stableID := range input.AffectedStableIDs {
		if !validAffectedID(stableID) {
			return 0, validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
		}
		if _, exists := registry[stableID]; exists {
			return 0, validationFailure{DecisionFailClosed, ReasonDuplicateAffectedBinding}
		}
		registry[stableID] = struct{}{}
	}
	selected := map[string]struct{}{}
	for _, id := range ids {
		selected[id] = struct{}{}
	}
	union := map[string]struct{}{}
	bound := map[string]struct{}{}
	for _, id := range sortedCommandIDs(commands) {
		command := commands[id]
		if command.AffectedStableIDs == nil {
			return 0, validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
		}
		local := map[string]struct{}{}
		for _, stableID := range command.AffectedStableIDs {
			if !validAffectedID(stableID) {
				return 0, validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
			}
			if _, exists := local[stableID]; exists {
				return 0, validationFailure{DecisionFailClosed, ReasonDuplicateAffectedBinding}
			}
			if _, exists := registry[stableID]; !exists {
				return 0, validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
			}
			local[stableID], bound[stableID] = struct{}{}, struct{}{}
			if _, exists := selected[id]; exists {
				union[stableID] = struct{}{}
			}
		}
	}
	for _, stableID := range input.AffectedStableIDs {
		if _, exists := bound[stableID]; !exists {
			return 0, validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
		}
	}
	return uint64(len(union)), validationFailure{}
}
func baselinePressures(command CommandRecord, applicable *uint64, groups map[string]struct{}) bool {
	if command.Pressures == nil {
		return false
	}
	known := true
	for _, pressure := range command.Pressures {
		if pressure.Applicable == nil || pressure.IndependenceGroupID == "" {
			known = false
			continue
		}
		if *pressure.Applicable {
			var ok bool
			*applicable, ok = add(*applicable, 1)
			if !ok {
				known = false
			}
			groups[pressure.IndependenceGroupID] = struct{}{}
		}
	}
	return known
}
