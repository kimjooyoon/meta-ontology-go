package valuecatalog

import "fmt"

func validateClaimTransitionLedger(report Report) error {
	if len(report.Claims) != OperationSpecAxisTotal || len(report.ClaimTransitions) != OperationClaimEventTotal {
		return fmt.Errorf("operation claim transition denominator changed")
	}
	previous := ""
	for index, transition := range report.ClaimTransitions {
		claimIndex := index
		if index >= OperationSpecAxisTotal {
			claimIndex -= OperationSpecAxisTotal
		}
		claim := report.Claims[claimIndex]
		if transition.Sequence != index+1 || transition.ClaimID != claim.ClaimID ||
			transition.DeclarationDigest != claimDeclarationDigest(claim) {
			return fmt.Errorf("operation claim transition %d identity changed", index+1)
		}
		if transition.PreviousTransitionDigest != previous ||
			transition.TransitionDigest != claimTransitionDigest(transition) || !validDigest(transition.TransitionDigest) {
			return fmt.Errorf("operation claim transition %d chain changed", index+1)
		}
		if index < OperationSpecAxisTotal {
			if err := validateRegistrationTransition(transition, claim); err != nil {
				return err
			}
		} else if err := validateResolutionTransition(transition, claim, report.ProcessCoordinate); err != nil {
			return err
		}
		previous = transition.TransitionDigest
	}
	if report.ClaimTransitionHead != previous || !validDigest(report.ClaimTransitionHead) {
		return fmt.Errorf("operation claim transition head changed")
	}
	registered, accepted, unavailable := countClaimTransitionEvents(report.ClaimTransitions)
	metrics := report.OperationSpecMetrics
	if metrics.TransitionEventTotal != len(report.ClaimTransitions) || metrics.RegistrationEventTotal != registered ||
		metrics.EvidenceAcceptedTotal != accepted || metrics.EvidenceUnavailableTotal != unavailable {
		return fmt.Errorf("operation claim transition metrics changed")
	}
	return nil
}

func validateRegistrationTransition(transition ClaimTransition, claim Claim) error {
	if transition.Event != ClaimEventRegistered || transition.Before != ClaimStatusUnrecorded ||
		transition.After != ClaimStatusOpen || transition.EvidenceDigest != "" ||
		transition.Coordinate != claimCoordinate(claim, ReasonClaimDeclared) {
		return fmt.Errorf("claim %s registration changed", claim.ClaimID)
	}
	return nil
}

func validateResolutionTransition(transition ClaimTransition, claim Claim, unknown ProcessCoordinate) error {
	if transition.Before != ClaimStatusOpen {
		return fmt.Errorf("claim %s resolution origin changed", claim.ClaimID)
	}
	if claim.Status == ClaimStatusDischarged && transition.Event == ClaimEventEvidenceAccepted &&
		transition.After == ClaimStatusDischarged && transition.EvidenceDigest == claim.EvidenceDigest &&
		validDigest(transition.EvidenceDigest) && transition.Coordinate == claimCoordinate(claim, ReasonClaimEvidenceAccepted) {
		return nil
	}
	if claim.Status == ClaimStatusOpen && transition.Event == ClaimEventEvidenceUnavailable &&
		transition.After == ClaimStatusOpen && transition.EvidenceDigest == "" && transition.Coordinate == unknown {
		return nil
	}
	return fmt.Errorf("claim %s resolution changed", claim.ClaimID)
}
