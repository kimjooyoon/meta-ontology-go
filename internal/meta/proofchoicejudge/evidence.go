package proofchoicejudge

type evidence struct {
	ClaimID              string            `json:"claim_id"`
	Subject              string            `json:"subject"`
	Route                string            `json:"route"`
	Producer             string            `json:"producer"`
	Consumer             string            `json:"consumer"`
	State                string            `json:"state"`
	Reason               string            `json:"reason,omitempty"`
	ObservationIDs       []string          `json:"observation_ids"`
	ObservationSlots     []observationSlot `json:"observation_slots"`
	StableIdentity       string            `json:"stable_identity,omitempty"`
	OriginDigest         string            `json:"origin_digest,omitempty"`
	SubjectBinding       string            `json:"subject_binding,omitempty"`
	ProjectionA          string            `json:"projection_a,omitempty"`
	ProjectionB          string            `json:"projection_b,omitempty"`
	Agreement            bool              `json:"agreement,omitempty"`
	FirstArtifactDigest  string            `json:"first_artifact_digest,omitempty"`
	SecondArtifactDigest string            `json:"second_artifact_digest,omitempty"`
	ByteEqual            bool              `json:"byte_equal,omitempty"`
	SemanticEqual        bool              `json:"semantic_equal,omitempty"`
	EvidenceDigest       string            `json:"evidence_digest"`
	Provenance           []string          `json:"provenance"`
}

type observationSlot struct {
	ID         string   `json:"id"`
	Observed   bool     `json:"observed"`
	Provenance []string `json:"provenance"`
}

type evidenceBundle struct {
	ByValue map[string][]evidence
	All     []evidence
}

func buildEvidence(values []value, source lowered, baseline []byte) evidenceBundle {
	bundle := evidenceBundle{ByValue: map[string][]evidence{}}
	for _, current := range values {
		for _, route := range current.EvidenceCapabilities {
			var result evidence
			switch route {
			case "FOUNDATION":
				result = foundationEvidence(current, source)
			case "COHERENCE":
				result = coherenceEvidence(current, values, source)
			case "REGRESSION":
				result = regressionEvidence(current, source, baseline)
			}
			bundle.ByValue[current.ID] = append(bundle.ByValue[current.ID], result)
			bundle.All = append(bundle.All, result)
		}
	}
	return bundle
}
func foundationEvidence(current value, source lowered) evidence {
	nodeID := source.Bindings[current.ID]
	result := evidence{ClaimID: current.ID, Subject: current.Subject, Route: "FOUNDATION"}
	if !stableSubject(current.Subject) || nodeID == "" {
		result.State, result.Reason = "INSUFFICIENT", "FOUNDATION_IDENTITY_UNSTABLE"
		return finishEvidence(result)
	}
	result.State = "OBSERVED"
	result.StableIdentity = digestBytes([]byte(nodeID))
	result.OriginDigest = source.SemanticDigest
	result.SubjectBinding = digestBytes([]byte(current.Subject + "|" + nodeID))
	result.ObservationSlots = []observationSlot{
		{ID: current.ID + ":stable-identity", Observed: true, Provenance: []string{"ir://node/" + nodeID}},
		{ID: current.ID + ":origin-digest", Observed: true, Provenance: []string{"ir://semantic/" + source.SemanticDigest}},
		{ID: current.ID + ":subject-binding", Observed: true, Provenance: []string{"subject://" + current.Subject}},
	}
	result.Provenance = provenanceOf(result.ObservationSlots)
	return finishEvidence(result)
}
