package proofchoicejudge

func itemFor(subject value, route routeResult, observations map[string]value) item {
	result := item{Kind: subject.Kind, ID: subject.ID, Statement: subject.Statement, PriorState: subject.PriorState, Choice: route.Route, Resolution: route.Resolution, Observations: route.Observations, EvidenceDigest: evidenceDigest(route.Observations, observations), Provenance: evidenceProvenance(route.Observations, observations)}
	if subject.Kind == "metric" {
		result.Denominator = len(subject.Slots)
		for _, slot := range subject.Slots {
			if slot.Observed {
				result.Numerator++
			}
		}
	}
	return result
}

func transitionFor(sequence int, subject value, route routeResult, observations map[string]value) transition {
	to := "OPEN"
	if route.Resolution == "EXACT" {
		to = "DISCHARGED"
	}
	if route.Resolution == "FAIL_CLOSED" {
		to = "REFUTED"
	}
	return transition{Sequence: sequence, ClaimID: subject.ID, From: subject.PriorState, To: to, Choice: route.Route, Resolution: route.Resolution, Stage: "semantic-resolution", Step: "evidence-route", Reason: route.Reason, EvidenceDigest: evidenceDigest(route.Observations, observations), Provenance: evidenceProvenance(route.Observations, observations), Persistent: true}
}
