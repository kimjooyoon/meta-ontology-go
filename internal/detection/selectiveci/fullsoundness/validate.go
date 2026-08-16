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

func commandMap(values []Command, obligations map[string]ObligationAuthority) (map[string]Command, Reason) {
	result := make(map[string]Command, len(values))
	for _, value := range values {
		if !validID(value.ID) || value.ObligationIDs == nil {
			return nil, ReasonFullSuiteRequired
		}
		if _, exists := result[value.ID]; exists {
			return nil, ReasonFullSuiteRequired
		}
		ids, unique := stringSet(value.ObligationIDs)
		if !unique {
			return nil, ReasonFullSuiteRequired
		}
		for id := range ids {
			if !validID(id) {
				return nil, ReasonFullSuiteRequired
			}
			if _, registered := obligations[id]; !registered {
				return nil, ReasonUnregisteredObligation
			}
		}
		result[value.ID] = value
	}
	return result, ""
}

func impactedSet(values []string, obligations map[string]ObligationAuthority) (map[string]struct{}, Reason) {
	result, unique := stringSet(values)
	if !unique {
		return nil, ReasonFullSuiteRequired
	}
	for id := range result {
		authority, registered := obligations[id]
		if !validID(id) {
			return nil, ReasonFullSuiteRequired
		}
		if !registered {
			return nil, ReasonUnregisteredObligation
		}
		if authority != AuthorityAuthoritative {
			return nil, ReasonUnprovableObligation
		}
	}
	return result, ""
}

func knownSet(values []string, commands map[string]Command) (map[string]struct{}, Reason) {
	result, unique := stringSet(values)
	if !unique {
		return nil, ReasonFullSuiteRequired
	}
	for id := range result {
		if !validID(id) || commands[id].ID == "" {
			return nil, ReasonFullSuiteRequired
		}
	}
	return result, ""
}

func validateReceipt(input Input, state evaluationState) Reason {
	receipt := input.SelectionReceipt
	if !validDigest(receipt.SnapshotDigest) || !validDigest(receipt.PolicyDigest) || !validDigest(receipt.RegistryDigest) || !validDigest(receipt.SelectionDigest) {
		return ReasonFullSuiteRequired
	}
	if receipt.SnapshotDigest != input.SnapshotDigest || receipt.PolicyDigest != input.PolicyDigest || receipt.RegistryDigest != input.RegistryDigest || receipt.SelectionDigest != input.SelectionDigest {
		return ReasonDigestBindingMismatch
	}
	_, reason := knownSet(receipt.CommandIDs, state.commands)
	return reason
}

func validateOutcomes(values []Outcome, commands map[string]Command) Reason {
	if len(values) != len(commands) {
		return ReasonFullSuiteRequired
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validID(value.CommandID) || commands[value.CommandID].ID == "" || !validOutcome(value) {
			return ReasonInvalidOutcome
		}
		if _, exists := seen[value.CommandID]; exists {
			return ReasonFullSuiteRequired
		}
		seen[value.CommandID] = struct{}{}
	}
	return ""
}

func outcomeMap(values []Outcome) map[string]Outcome {
	result := make(map[string]Outcome, len(values))
	for _, value := range values {
		result[value.CommandID] = value
	}
	return result
}

func (state evaluationState) selectedCommands() map[string]Command {
	result := make(map[string]Command, len(state.selected))
	for id := range state.selected {
		result[id] = state.commands[id]
	}
	return result
}

func validInputDigests(input Input) bool {
	return validDigest(input.SnapshotDigest) && validDigest(input.PolicyDigest) && validDigest(input.RegistryDigest) && validDigest(input.SelectionDigest) && validDigest(input.ToolchainDigest) && validDigest(input.RunnerDigest)
}

func validID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	if value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		char := value[index]
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func validAuthority(value ObligationAuthority) bool {
	return value == AuthorityAuthoritative || value == AuthorityCandidate || value == AuthorityDerived
}

func validOutcome(value Outcome) bool {
	if (value.Status != OutcomePass && value.Status != OutcomeFail) || !validDigest(value.OutputDigest) {
		return false
	}
	return value.Status != OutcomePass || value.FailureCode == ""
}
