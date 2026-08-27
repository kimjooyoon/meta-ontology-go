package proofchoicealgebra

func transitionFor(sequence int, value Value, result routeResult, observations map[string]Value) Transition {
	to := "OPEN"
	if result.Resolution == Exact {
		to = "DISCHARGED"
	}
	if result.Resolution == FailClosed {
		to = "REFUTED"
	}
	return Transition{
		Sequence: sequence, ClaimID: value.ID, From: value.PriorState, To: to,
		Choice: result.Route, Resolution: result.Resolution,
		Stage: "semantic-resolution", Step: "evidence-route", Reason: result.Reason,
		EvidenceDigest: evidenceDigest(result.Observations, observations),
		Provenance:     evidenceProvenance(result.Observations, observations), Persistent: true,
	}
}
