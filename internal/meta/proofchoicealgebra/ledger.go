package proofchoicealgebra

func transitionFor(sequence int, value Value, result routeResult) Transition {
	to := "OPEN"
	if result.Resolution == Exact && !result.FailClosed {
		to = "DISCHARGED"
	}
	if result.FailClosed {
		to = "REFUTED"
	}
	return Transition{
		Sequence: sequence, ClaimID: value.ID, From: value.PriorState, To: to,
		Choice: string(result.Route), Resolution: result.Resolution,
		Stage: "semantic-resolution", Step: "evidence-route", Reason: result.Reason,
		EvidenceDigest: result.EvidenceDigest, Provenance: result.Provenance,
		Persistent: true,
	}
}
