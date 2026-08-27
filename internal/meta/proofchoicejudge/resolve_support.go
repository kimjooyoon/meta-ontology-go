package proofchoicejudge

func splitValues(values []value) ([]value, string) {
	seen := map[string]bool{}
	result := make([]value, 0, len(values))
	for _, current := range values {
		if seen[current.ID] {
			return nil, "PROOF_VALUE_DUPLICATE"
		}
		seen[current.ID] = true
		result = append(result, current)
	}
	if len(result) == 0 {
		return nil, "NO_PROOF_SUBJECTS"
	}
	return result, ""
}

func validateValue(current value) string {
	if current.Kind == "claim" && (current.PriorState != "OPEN" || current.Statement == "") {
		return "CLAIM_PROPOSITION_UNKNOWN"
	}
	if current.Kind == "composition" && len(current.Members) == 0 {
		return "COMPOSITION_MEMBERS_UNKNOWN"
	}
	if current.Kind != "composition" && current.Subject == "" {
		return "SUBJECT_UNKNOWN"
	}
	return ""
}

func itemFor(current value, route routeResult, bundle evidenceBundle) item {
	result := item{Kind: current.Kind, ID: current.ID, Statement: current.Statement, PriorState: current.PriorState, Choice: route.Route, Resolution: route.Resolution, ObservationState: route.ObservationState, Observations: route.Observations, EvidenceDigest: route.EvidenceDigest, Provenance: route.Provenance}
	if current.Kind == "metric" {
		result.Denominator = 3
		for _, currentEvidence := range evidenceFor(current, bundle) {
			for _, slot := range currentEvidence.ObservationSlots {
				if slot.Observed {
					result.Numerator++
				}
			}
			if len(currentEvidence.ObservationSlots) > 0 {
				break
			}
		}
	}
	return result
}

func transitionFor(sequence int, current value, route routeResult) transition {
	to := "OPEN"
	if route.Resolution == "EXACT" && !route.FailClosed {
		to = "DISCHARGED"
	}
	if route.FailClosed {
		to = "REFUTED"
	}
	return transition{Sequence: sequence, ClaimID: current.ID, From: current.PriorState, To: to, Choice: route.Route, Resolution: route.Resolution, Stage: "semantic-resolution", Step: "evidence-route", Reason: route.Reason, EvidenceDigest: route.EvidenceDigest, Provenance: route.Provenance, Persistent: true}
}
