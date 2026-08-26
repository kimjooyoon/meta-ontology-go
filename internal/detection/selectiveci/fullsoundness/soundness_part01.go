package fullsoundness

func validateReceipts(input Input, state *evaluationState) Reason {
	full, reason := receiptMap(input.FullResourceReceipts, state.commands, input)
	if reason != "" {
		return reason
	}
	selected, reason := receiptMap(input.SelectedResourceReceipts, state.selectedCommands(), input)
	if reason != "" {
		return reason
	}
	state.fullReceipts = full
	state.selectedReceipts = selected
	return ""
}
func receiptMap(values []ResourceReceipt, commands map[string]Command, input Input) (map[string]ResourceReceipt, Reason) {
	if len(values) != len(commands) {
		return nil, ReasonFullSuiteRequired
	}
	result := make(map[string]ResourceReceipt, len(values))
	for _, value := range values {
		if !validID(value.CommandID) || commands[value.CommandID].ID == "" {
			return nil, ReasonFullSuiteRequired
		}
		if _, exists := result[value.CommandID]; exists {
			return nil, ReasonFullSuiteRequired
		}
		if !validReceiptDigests(value) {
			return nil, ReasonFullSuiteRequired
		}
		if value.SnapshotDigest != input.SnapshotDigest || value.ToolchainDigest != input.ToolchainDigest || value.RunnerDigest != input.RunnerDigest {
			return nil, ReasonFullSuiteRequired
		}
		if !validResourceNumbers(value) {
			return nil, ReasonFullSuiteRequired
		}
		result[value.CommandID] = value
	}
	return result, ""
}
func validReceiptDigests(value ResourceReceipt) bool {
	return validDigest(value.SnapshotDigest) && validDigest(value.ToolchainDigest) && validDigest(value.RunnerDigest)
}
func checkSoundness(state evaluationState) Reason {
	if reason := compareSelectedOutcomes(state); reason != "" {
		return reason
	}
	if fullFailureOmitted(state) {
		return ReasonOmittedFullFailure
	}
	if impactedCommandOmitted(state) {
		return ReasonImpactedCommandOmitted
	}
	if unprovableCommandOmitted(state) {
		return ReasonUnprovableObligation
	}
	return ""
}
func globalGuardOmitted(state evaluationState) bool {
	for id, command := range state.commands {
		if command.GlobalGuard {
			if _, selected := state.selected[id]; !selected {
				return true
			}
		}
	}
	return false
}
