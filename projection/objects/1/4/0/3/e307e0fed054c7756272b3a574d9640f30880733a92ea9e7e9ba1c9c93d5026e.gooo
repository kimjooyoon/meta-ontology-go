package fullsoundness

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
