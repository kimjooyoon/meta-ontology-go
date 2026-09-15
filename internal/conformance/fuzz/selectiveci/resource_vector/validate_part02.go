package resourcevector

import (
	"sort"
	"strings"
)

func validateAffected(input Input, commands map[string]CommandRecord) validationFailure {
	if input.AffectedStableIDs == nil {
		return validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
	}
	registry := map[string]struct{}{}
	for _, stableID := range input.AffectedStableIDs {
		if !validAffectedID(stableID) {
			return validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
		}
		if _, exists := registry[stableID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateAffectedBinding}
		}
		registry[stableID] = struct{}{}
	}
	bound := map[string]struct{}{}
	for _, id := range sortedCommandIDs(commands) {
		command := commands[id]
		if command.AffectedStableIDs == nil {
			return validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
		}
		local := map[string]struct{}{}
		for _, stableID := range command.AffectedStableIDs {
			if !validAffectedID(stableID) {
				return validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
			}
			if _, exists := local[stableID]; exists {
				return validationFailure{DecisionFailClosed, ReasonDuplicateAffectedBinding}
			}
			if _, exists := registry[stableID]; !exists {
				return validationFailure{DecisionFailClosed, ReasonDanglingAffectedBinding}
			}
			local[stableID], bound[stableID] = struct{}{}, struct{}{}
		}
	}
	for _, stableID := range input.AffectedStableIDs {
		if _, exists := bound[stableID]; !exists {
			return validationFailure{DecisionUnknown, ReasonMissingAffectedBinding}
		}
	}
	return validationFailure{}
}
func validAffectedID(value string) bool {
	return strings.HasPrefix(value, "s-") && validToken(value)
}
func sortedCommandIDs(commands map[string]CommandRecord) []string {
	ids := make([]string, 0, len(commands))
	for id := range commands {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
