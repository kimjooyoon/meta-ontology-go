package proofchoicealgebra

type evidenceBundle struct {
	ByValue map[string][]Evidence
	All     []Evidence
}

func buildEvidence(values []Value, lowered lowered, baseline []byte) evidenceBundle {
	bundle := evidenceBundle{ByValue: map[string][]Evidence{}}
	for _, value := range values {
		for _, route := range value.EvidenceCapabilities {
			var result Evidence
			switch route {
			case Foundation:
				result = foundationEvidence(value, lowered)
			case Coherence:
				result = coherenceEvidence(value, values, lowered)
			case Regression:
				result = regressionEvidence(value, lowered, baseline)
			}
			bundle.ByValue[value.ID] = append(bundle.ByValue[value.ID], result)
			bundle.All = append(bundle.All, result)
		}
	}
	return bundle
}

func foundationEvidence(value Value, lowered lowered) Evidence {
	nodeID := lowered.Bindings[value.ID]
	result := Evidence{ClaimID: value.ID, Subject: value.Subject, Route: Foundation}
	if !stableSubject(value.Subject) || nodeID == "" {
		result.State = InsufficientState
		result.Reason = "FOUNDATION_IDENTITY_UNSTABLE"
		return finishEvidence(result)
	}
	result.State = "OBSERVED"
	result.StableIdentity = digestBytes([]byte(nodeID))
	result.OriginDigest = lowered.SemanticDigest
	result.SubjectBinding = digestBytes([]byte(value.Subject + "|" + nodeID))
	result.ObservationSlots = []ObservationSlot{
		{ID: value.ID + ":stable-identity", Observed: true, Provenance: []string{"ir://node/" + nodeID}},
		{ID: value.ID + ":origin-digest", Observed: true, Provenance: []string{"ir://semantic/" + lowered.SemanticDigest}},
		{ID: value.ID + ":subject-binding", Observed: true, Provenance: []string{"subject://" + value.Subject}},
	}
	result.Provenance = provenanceOf(result.ObservationSlots)
	return finishEvidence(result)
}
