package valuecatalog

func buildOperationClaimTransitions(report Report, checks []operationSpecCheck) []ClaimTransition {
	ledger := make([]ClaimTransition, 0, OperationClaimEventTotal)
	for _, claim := range report.Claims {
		ledger = appendClaimTransition(ledger, ClaimTransition{
			ClaimID: claim.ClaimID, DeclarationDigest: claimDeclarationDigest(claim),
			Event: ClaimEventRegistered, Before: ClaimStatusUnrecorded, After: ClaimStatusOpen,
			Coordinate: claimCoordinate(claim, ReasonClaimDeclared),
		})
	}
	for index, check := range checks {
		claim := report.Claims[index]
		transition := ClaimTransition{
			ClaimID: claim.ClaimID, DeclarationDigest: claimDeclarationDigest(claim),
			Event: ClaimEventEvidenceUnavailable, Before: ClaimStatusOpen, After: ClaimStatusOpen,
			Coordinate: report.ProcessCoordinate,
		}
		if check.satisfied {
			transition.Event = ClaimEventEvidenceAccepted
			transition.After = ClaimStatusDischarged
			transition.Coordinate = claimCoordinate(claim, ReasonClaimEvidenceAccepted)
			transition.EvidenceDigest = check.evidence
		}
		ledger = appendClaimTransition(ledger, transition)
	}
	return ledger
}

func countClaimTransitionEvents(ledger []ClaimTransition) (registered, accepted, unavailable int) {
	for _, transition := range ledger {
		switch transition.Event {
		case ClaimEventRegistered:
			registered++
		case ClaimEventEvidenceAccepted:
			accepted++
		case ClaimEventEvidenceUnavailable:
			unavailable++
		}
	}
	return registered, accepted, unavailable
}
