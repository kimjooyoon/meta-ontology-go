package fullsoundness

func sealProjectionFailure(output Output) Output {
	output.Decision = DecisionUnknown
	output.Reason = ReasonResourceOverflow
	output.ResourceVector = nil
	output.FullFailureCommandIDs = nil
	output.SelectedFailureCommandIDs = nil
	output.OmittedCommandIDs = nil
	output.SemanticEvaluated = false
	output = normalizeOutput(output)
	output.DecisionDigest, _ = output.DecisionStableDigest()
	output.CanonicalDigest = output.StableDigest()
	return output
}
func missingRequiredInput(input Input) bool {
	return input.Obligations == nil || input.Commands == nil || input.ImpactedObligationIDs == nil || input.SelectedCommandIDs == nil || input.SelectionReceipt == nil || input.SelectionReceipt.CommandIDs == nil || input.FullOutcomes == nil || input.SelectedOutcomes == nil || input.FullResourceReceipts == nil || input.SelectedResourceReceipts == nil
}
func populateCounts(output *Output, input Input) {
	output.ObligationCount = uint64(len(input.Obligations))
	output.CommandCount = uint64(len(input.Commands))
	output.SelectedCommandCount = uint64(len(input.SelectedCommandIDs))
}
func populateCommandLists(output *Output, state evaluationState) {
	for id, outcome := range state.fullOutcomes {
		if outcome.Status == OutcomeFail {
			output.FullFailureCommandIDs = append(output.FullFailureCommandIDs, id)
		}
	}
	for id, outcome := range state.selectedOutcomes {
		if outcome.Status == OutcomeFail {
			output.SelectedFailureCommandIDs = append(output.SelectedFailureCommandIDs, id)
		}
	}
	for id := range state.commands {
		if _, selected := state.selected[id]; !selected {
			output.OmittedCommandIDs = append(output.OmittedCommandIDs, id)
		}
	}
	for id := range state.impacted {
		if state.obligations[id] == AuthorityAuthoritative {
			output.AuthoritativeImpactedObligationCount++
		}
	}
}
func decisionFor(reason Reason) Decision {
	if reason == ReasonUnprovableObligation {
		return DecisionUnknown
	}
	return DecisionUnsound
}
