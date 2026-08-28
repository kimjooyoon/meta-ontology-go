package main

import "strings"

const claimResolutionProgramPrefix = "claim.resolve:v1"

type claimTupleFailure struct {
	reason string
	next   string
}

func parseClaimValueProgram(program string) (claimResolutionClaim, int, *claimTupleFailure) {
	fields := make(map[string]string, 6)
	index := 0
	for part := range strings.SplitSeq(program, ";") {
		if index == 0 {
			if part != claimResolutionProgramPrefix {
				return claimResolutionClaim{}, 0, newClaimTupleFailure("CLAIM_RESOLUTION_PREFIX_INVALID", "RESTORE_CLAIM_RESOLUTION_PREFIX")
			}
			index++
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || !claimResolutionFieldAllowed(key) || !claimResolutionAtomValid(value) {
			return claimResolutionClaim{}, len(fields), newClaimTupleFailure("CLAIM_RESOLUTION_FIELD_INVALID", "RESTORE_CLAIM_RESOLUTION_FIELDS")
		}
		if _, duplicate := fields[key]; duplicate {
			return claimResolutionClaim{}, len(fields), newClaimTupleFailure("CLAIM_RESOLUTION_FIELD_DUPLICATE", "REMOVE_DUPLICATE_CLAIM_RESOLUTION_FIELD")
		}
		fields[key] = value
		index++
	}
	if index != 7 || len(fields) != 6 {
		return claimResolutionClaim{}, len(fields), newClaimTupleFailure("CLAIM_RESOLUTION_FIELD_MISSING", "PROVIDE_SIX_CLAIM_RESOLUTION_FIELDS")
	}
	claim := claimResolutionClaim{
		State: fields["state"], Stage: claimOptional(fields["stage"]), Step: claimOptional(fields["step"]),
		Reason: fields["reason"], UnknownClass: claimOptional(fields["unknown_class"]),
		NextOperation: fields["next_operation"],
	}
	if failed := validateResolutionClaim(claim); failed != nil {
		return claimResolutionClaim{}, len(fields), failed
	}
	return claim, len(fields), nil
}

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
