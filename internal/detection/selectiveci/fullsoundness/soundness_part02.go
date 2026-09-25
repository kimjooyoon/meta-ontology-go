package fullsoundness

func compareSelectedOutcomes(state evaluationState) Reason {
	for id := range state.selected {
		full := state.fullOutcomes[id]
		selected := state.selectedOutcomes[id]
		if full.Status == OutcomePass && selected.Status == OutcomeFail {
			return ReasonSelectedExtraFailure
		}
		if full.Status != selected.Status {
			return ReasonSelectedFullStatusMismatch
		}
		if full.FailureCode != selected.FailureCode {
			return ReasonFailureCodeMismatch
		}
		if full.OutputDigest != selected.OutputDigest {
			return ReasonOutputDigestMismatch
		}
	}
	return ""
}
func fullFailureOmitted(state evaluationState) bool {
	for id, outcome := range state.fullOutcomes {
		if outcome.Status == OutcomeFail {
			if _, selected := state.selected[id]; !selected {
				return true
			}
		}
	}
	return false
}
func impactedCommandOmitted(state evaluationState) bool {
	for id, command := range state.commands {
		if _, selected := state.selected[id]; selected {
			continue
		}
		for _, obligationID := range command.ObligationIDs {
			if _, impacted := state.impacted[obligationID]; impacted {
				return true
			}
		}
	}
	return false
}
func unprovableCommandOmitted(state evaluationState) bool {
	for id, command := range state.commands {
		if _, selected := state.selected[id]; selected {
			continue
		}
		for _, obligationID := range command.ObligationIDs {
			if state.obligations[obligationID] != AuthorityAuthoritative {
				return true
			}
		}
	}
	return false
}
