package resourcevector

func validatePressure(pressure PressureRecord, local, global map[string]struct{}) validationFailure {
	if !validToken(pressure.ID) || !validToken(pressure.IndependenceGroupID) {
		return validationFailure{DecisionFailClosed, ReasonInvalidPressure}
	}
	if pressure.Applicable == nil {
		return validationFailure{DecisionUnknown, ReasonMissingResource}
	}
	if _, exists := local[pressure.ID]; exists {
		return validationFailure{DecisionFailClosed, ReasonDuplicateID}
	}
	if _, exists := global[pressure.ID]; exists {
		return validationFailure{DecisionFailClosed, ReasonDuplicateID}
	}
	local[pressure.ID], global[pressure.ID] = struct{}{}, struct{}{}
	return validationFailure{}
}
func validateSelections(input Input, commands map[string]CommandRecord) validationFailure {
	if failure := validateSelection(input.SelectedCommandIDs, commands, nil); failure.reason != "" {
		return failure
	}
	if failure := validateSelection(input.FullCommandIDs, commands, input.SelectedCommandIDs); failure.reason != "" {
		return failure
	}
	full := map[string]struct{}{}
	for _, id := range input.FullCommandIDs {
		full[id] = struct{}{}
	}
	if len(full) != len(commands) {
		return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
	}
	for id := range commands {
		if _, exists := full[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
	}
	return validationFailure{}
}
func validateSelection(values []string, commands map[string]CommandRecord, subset []string) validationFailure {
	seen := map[string]struct{}{}
	for _, id := range values {
		if !validToken(id) {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
		if _, exists := seen[id]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateID}
		}
		if _, exists := commands[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonDanglingID}
		}
		seen[id] = struct{}{}
	}
	if subset == nil {
		return validationFailure{}
	}
	for _, id := range subset {
		if _, exists := seen[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
	}
	return validationFailure{}
}
