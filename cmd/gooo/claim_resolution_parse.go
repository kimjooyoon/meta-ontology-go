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
