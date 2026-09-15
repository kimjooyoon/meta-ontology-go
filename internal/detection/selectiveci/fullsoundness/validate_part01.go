package fullsoundness

func validateInput(input Input) (evaluationState, Reason) {
	if input.SchemaVersion != SchemaVersion || !validInputDigests(input) {
		return evaluationState{}, ReasonFullSuiteRequired
	}
	state, reason := validateRegistries(input)
	if reason != "" {
		return evaluationState{}, reason
	}
	if reason := validateReceipt(input, state); reason != "" {
		return evaluationState{}, reason
	}
	if reason := validateOutcomes(input.FullOutcomes, state.commands); reason != "" {
		return evaluationState{}, reason
	}
	if reason := validateOutcomes(input.SelectedOutcomes, state.selectedCommands()); reason != "" {
		return evaluationState{}, reason
	}
	state.fullOutcomes = outcomeMap(input.FullOutcomes)
	state.selectedOutcomes = outcomeMap(input.SelectedOutcomes)
	if reason := validateReceipts(input, &state); reason != "" {
		return evaluationState{}, reason
	}
	return state, ""
}
func validateRegistries(input Input) (evaluationState, Reason) {
	obligations, reason := obligationMap(input.Obligations)
	if reason != "" {
		return evaluationState{}, reason
	}
	commands, reason := commandMap(input.Commands, obligations)
	if reason != "" {
		return evaluationState{}, reason
	}
	impacted, reason := impactedSet(input.ImpactedObligationIDs, obligations)
	if reason != "" {
		return evaluationState{}, reason
	}
	selected, reason := knownSet(input.SelectedCommandIDs, commands)
	if reason != "" {
		return evaluationState{}, reason
	}
	return evaluationState{obligations: obligations, commands: commands, impacted: impacted, selected: selected}, ""
}
func obligationMap(values []Obligation) (map[string]ObligationAuthority, Reason) {
	result := make(map[string]ObligationAuthority, len(values))
	for _, value := range values {
		if !validID(value.ID) || !validAuthority(value.Authority) {
			return nil, ReasonFullSuiteRequired
		}
		if _, exists := result[value.ID]; exists {
			return nil, ReasonFullSuiteRequired
		}
		result[value.ID] = value.Authority
	}
	return result, ""
}
