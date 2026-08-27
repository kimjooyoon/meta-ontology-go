package proofchoicejudge

func splitValues(values []value) (map[string]value, []value, string, value) {
	observations := map[string]value{}
	subjects := []value{}
	seen := map[string]bool{}
	for _, current := range values {
		if seen[current.ID] {
			return observations, subjects, "PROOF_ROUTE_CONTRADICTION", current
		}
		seen[current.ID] = true
	}
	for _, current := range values {
		switch current.Kind {
		case "observation":
			if current.EvidenceKind == "" || current.Predicate == "" || current.Value == "" || len(current.Provenance) == 0 {
				return nil, nil, "EVIDENCE_UNKNOWN", value{}
			}
			observations[current.ID] = current
		case "claim", "metric":
			subjects = append(subjects, current)
		default:
			return nil, nil, "SEMANTIC_VALUE_UNKNOWN", value{}
		}
	}
	if len(subjects) == 0 {
		return nil, nil, "NO_PROOF_SUBJECTS", value{}
	}
	return observations, subjects, "", value{}
}

func validateSubject(subject value, observations map[string]value) string {
	for _, dependency := range subject.Dependencies {
		if _, exists := observations[dependency]; !exists {
			return "EVIDENCE_UNKNOWN"
		}
	}
	for _, observation := range subject.Observations {
		if _, exists := observations[observation]; !exists {
			return "EVIDENCE_UNKNOWN"
		}
	}
	if subject.Kind == "claim" && (subject.PriorState != "OPEN" || subject.Statement == "") {
		return "CLAIM_PRIOR_STATE_UNKNOWN"
	}
	return ""
}
