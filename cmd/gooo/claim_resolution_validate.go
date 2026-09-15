package main

func validateResolutionClaim(claim claimResolutionClaim) *claimTupleFailure {
	if claim.Reason == claimResolutionNone {
		return newClaimTupleFailure("CLAIM_REASON_REQUIRED", "PROVIDE_CLAIM_REASON")
	}
	switch claim.State {
	case claimStateClosed:
		if claim.Stage != nil || claim.Step != nil || claim.UnknownClass != nil || claim.NextOperation != claimResolutionNone {
			return newClaimTupleFailure("CLOSED_CLAIM_BOUNDARY_INVALID", "RESTORE_CLOSED_CLAIM_BOUNDARY")
		}
	case claimStateUnknown:
		if claim.Stage == nil || claim.Step == nil || claim.UnknownClass == nil || claim.NextOperation == claimResolutionNone {
			return newClaimTupleFailure("UNKNOWN_TUPLE_INCOMPLETE", "PROVIDE_COMPLETE_UNKNOWN_TUPLE")
		}
	case claimStateRefuted:
		if claim.Stage == nil || claim.Step == nil || claim.UnknownClass != nil || claim.NextOperation == claimResolutionNone {
			return newClaimTupleFailure("REFUTED_CLAIM_BOUNDARY_INVALID", "RESTORE_REFUTED_CLAIM_BOUNDARY")
		}
	default:
		return newClaimTupleFailure("CLAIM_STATE_UNKNOWN", "RESTORE_CLOSED_UNKNOWN_OR_REFUTED_STATE")
	}
	return nil
}

func claimResolutionFieldAllowed(key string) bool {
	switch key {
	case "state", "stage", "step", "reason", "unknown_class", "next_operation":
		return true
	default:
		return false
	}
}

func claimResolutionAtomValid(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
}

func claimOptional(value string) *string {
	if value == claimResolutionNone {
		return nil
	}
	return &value
}

func newClaimTupleFailure(reason, next string) *claimTupleFailure {
	return &claimTupleFailure{reason: reason, next: next}
}
