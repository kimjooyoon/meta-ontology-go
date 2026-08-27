package proofchoicealgebra

func splitValues(values []Value) (map[string]Value, []Value, string) {
	observations := map[string]Value{}
	subjects := []Value{}
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value.ID] {
			return observations, []Value{value}, "PROOF_ROUTE_CONTRADICTION"
		}
		seen[value.ID] = true
	}
	for _, value := range values {
		switch value.Kind {
		case ObservationKind:
			if value.EvidenceKind == "" || value.Predicate == "" || value.Value == "" || len(value.Provenance) == 0 {
				return nil, nil, "EVIDENCE_UNKNOWN"
			}
			observations[value.ID] = value
		case ClaimKind, MetricKind:
			subjects = append(subjects, value)
		default:
			return nil, nil, "SEMANTIC_VALUE_UNKNOWN"
		}
	}
	if len(subjects) == 0 {
		return nil, nil, "NO_PROOF_SUBJECTS"
	}
	return observations, subjects, ""
}

func validateDependencies(value Value, observations map[string]Value) string {
	for _, dependency := range value.Dependencies {
		if _, exists := observations[dependency]; !exists {
			return "EVIDENCE_UNKNOWN"
		}
	}
	for _, observation := range value.Observations {
		if _, exists := observations[observation]; !exists {
			return "EVIDENCE_UNKNOWN"
		}
	}
	if value.Kind == ClaimKind && len(value.Dependencies) == 0 {
		return "EVIDENCE_UNKNOWN"
	}
	if value.Kind == ClaimKind && value.PriorState != "OPEN" {
		return "CLAIM_PRIOR_STATE_UNKNOWN"
	}
	if value.Kind == ClaimKind && value.Statement == "" {
		return "CLAIM_STATEMENT_UNKNOWN"
	}
	return ""
}
