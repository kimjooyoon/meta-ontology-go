package claimresolution

import "strings"

const programPrefix = "claim.resolve:v1"

type tupleFailure struct {
	reason string
	next   string
}

func parseValueProgram(program string) (Claim, int, *tupleFailure) {
	fields := make(map[string]string, 6)
	index := 0
	for part := range strings.SplitSeq(program, ";") {
		if index == 0 {
			if part != programPrefix {
				return Claim{}, 0, failure("CLAIM_RESOLUTION_PREFIX_INVALID", "RESTORE_CLAIM_RESOLUTION_PREFIX")
			}
			index++
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || !allowedField(key) || !validAtom(value) {
			return Claim{}, len(fields), failure("CLAIM_RESOLUTION_FIELD_INVALID", "RESTORE_CLAIM_RESOLUTION_FIELDS")
		}
		if _, duplicate := fields[key]; duplicate {
			return Claim{}, len(fields), failure("CLAIM_RESOLUTION_FIELD_DUPLICATE", "REMOVE_DUPLICATE_CLAIM_RESOLUTION_FIELD")
		}
		fields[key] = value
		index++
	}
	if index != 7 || len(fields) != 6 {
		return Claim{}, len(fields), failure("CLAIM_RESOLUTION_FIELD_MISSING", "PROVIDE_SIX_CLAIM_RESOLUTION_FIELDS")
	}
	claim := Claim{
		State: fields["state"], Stage: optional(fields["stage"]), Step: optional(fields["step"]),
		Reason: fields["reason"], UnknownClass: optional(fields["unknown_class"]),
		NextOperation: fields["next_operation"],
	}
	if failed := validateClaim(claim); failed != nil {
		return Claim{}, len(fields), failed
	}
	return claim, len(fields), nil
}

func validateClaim(claim Claim) *tupleFailure {
	if claim.Reason == None {
		return failure("CLAIM_REASON_REQUIRED", "PROVIDE_CLAIM_REASON")
	}
	switch claim.State {
	case StateClosed:
		if claim.Stage != nil || claim.Step != nil || claim.UnknownClass != nil || claim.NextOperation != None {
			return failure("CLOSED_CLAIM_BOUNDARY_INVALID", "RESTORE_CLOSED_CLAIM_BOUNDARY")
		}
	case StateUnknown:
		if claim.Stage == nil || claim.Step == nil || claim.UnknownClass == nil || claim.NextOperation == None {
			return failure("UNKNOWN_TUPLE_INCOMPLETE", "PROVIDE_COMPLETE_UNKNOWN_TUPLE")
		}
	case StateRefuted:
		if claim.Stage == nil || claim.Step == nil || claim.UnknownClass != nil || claim.NextOperation == None {
			return failure("REFUTED_CLAIM_BOUNDARY_INVALID", "RESTORE_REFUTED_CLAIM_BOUNDARY")
		}
	default:
		return failure("CLAIM_STATE_UNKNOWN", "RESTORE_CLOSED_UNKNOWN_OR_REFUTED_STATE")
	}
	return nil
}

func allowedField(key string) bool {
	switch key {
	case "state", "stage", "step", "reason", "unknown_class", "next_operation":
		return true
	default:
		return false
	}
}

func validAtom(value string) bool {
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

func optional(value string) *string {
	if value == None {
		return nil
	}
	return &value
}

func failure(reason, next string) *tupleFailure {
	return &tupleFailure{reason: reason, next: next}
}
