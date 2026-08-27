package proofchoicealgebra

func splitValues(values []Value) ([]Value, string) {
	seen := map[string]bool{}
	result := make([]Value, 0, len(values))
	for _, value := range values {
		if seen[value.ID] {
			return nil, "PROOF_VALUE_DUPLICATE"
		}
		seen[value.ID] = true
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil, "NO_PROOF_SUBJECTS"
	}
	return result, ""
}

func validateValue(value Value) string {
	if value.Kind == ClaimKind && (value.PriorState != "OPEN" || value.Statement == "") {
		return "CLAIM_PROPOSITION_UNKNOWN"
	}
	if value.Kind == CompositionKind && len(value.Members) == 0 {
		return "COMPOSITION_MEMBERS_UNKNOWN"
	}
	if value.Kind != CompositionKind && value.Subject == "" {
		return "SUBJECT_UNKNOWN"
	}
	return ""
}
