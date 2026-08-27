package proofchoicealgebra

func coherenceEvidence(value Value, values []Value, lowered lowered) Evidence {
	result := Evidence{ClaimID: value.ID, Subject: value.Subject, Route: Coherence}
	if value.Subject == "" || value.Statement == "" {
		result.State = UnknownState
		result.Reason = "COHERENCE_PROPOSITION_UNKNOWN"
		return finishEvidence(result)
	}
	for _, peer := range values {
		if peer.ID == value.ID || peer.Subject != value.Subject || peer.Statement != value.Statement {
			continue
		}
		result = compareProjections(value, peer, lowered)
		return finishEvidence(result)
	}
	result.State = InsufficientState
	result.Reason = "COHERENCE_SECOND_OBSERVATION_MISSING"
	return finishEvidence(result)
}

func compareProjections(value, peer Value, lowered lowered) Evidence {
	projectionA := digestBytes([]byte(value.Subject + "\x00" + value.Statement))
	projectionB := digestBytes([]byte(peer.Subject + "\x00" + peer.Statement))
	result := Evidence{
		ClaimID: value.ID, Subject: value.Subject, Route: Coherence,
		State: InsufficientState, ProjectionA: projectionA, ProjectionB: projectionB,
		Agreement: projectionA == projectionB,
	}
	result.ObservationSlots = []ObservationSlot{
		{ID: value.ID + ":projection-a", Observed: true, Provenance: []string{"ir://node/" + lowered.Bindings[value.ID]}},
		{ID: value.ID + ":projection-b", Observed: true, Provenance: []string{"ir://node/" + lowered.Bindings[peer.ID]}},
		{ID: value.ID + ":agreement", Observed: result.Agreement, Provenance: []string{"relation://" + value.Subject}},
	}
	if result.Agreement {
		result.State = "OBSERVED"
	} else {
		result.Reason = "COHERENCE_PROPOSITION_MISMATCH"
	}
	result.Provenance = provenanceOf(result.ObservationSlots)
	return result
}
