package proofchoicejudge

func coherenceEvidence(current value, values []value, source lowered) evidence {
	result := evidence{ClaimID: current.ID, Subject: current.Subject, Route: "COHERENCE"}
	if current.Subject == "" || current.Statement == "" {
		result.State, result.Reason = "UNKNOWN", "COHERENCE_PROPOSITION_UNKNOWN"
		return finishEvidence(result)
	}
	for _, peer := range values {
		if peer.ID == current.ID || peer.Subject != current.Subject || peer.Statement != current.Statement {
			continue
		}
		result = compareProjections(current, peer, source)
		return finishEvidence(result)
	}
	result.State, result.Reason = "INSUFFICIENT", "COHERENCE_SECOND_OBSERVATION_MISSING"
	return finishEvidence(result)
}

func compareProjections(current, peer value, source lowered) evidence {
	projectionA := digestBytes([]byte(current.Subject + "\x00" + current.Statement))
	projectionB := digestBytes([]byte(peer.Subject + "\x00" + peer.Statement))
	result := evidence{ClaimID: current.ID, Subject: current.Subject, Route: "COHERENCE", State: "INSUFFICIENT", ProjectionA: projectionA, ProjectionB: projectionB, Agreement: projectionA == projectionB}
	result.ObservationSlots = []observationSlot{
		{ID: current.ID + ":projection-a", Observed: true, Provenance: []string{"ir://node/" + source.Bindings[current.ID]}},
		{ID: current.ID + ":projection-b", Observed: true, Provenance: []string{"ir://node/" + source.Bindings[peer.ID]}},
		{ID: current.ID + ":agreement", Observed: result.Agreement, Provenance: []string{"relation://" + current.Subject}},
	}
	if result.Agreement {
		result.State = "OBSERVED"
	} else {
		result.Reason = "COHERENCE_PROPOSITION_MISMATCH"
	}
	result.Provenance = provenanceOf(result.ObservationSlots)
	return result
}
