package languageproofartifactverifier

type observation struct {
	Decision              string
	Resolution            string
	Reason                string
	Coordinate            Coordinate
	ArtifactDigest        string
	SourceDigest          string
	SemanticDigest        string
	OperationDigest       string
	EvidenceLinkDigest    string
	ClaimTransitionDigest string
	Claims                []ClaimResult
}

func failure(resolution, reason, stage, step string) observation {
	status := "REFUTED"
	if resolution == "LOWER_RESOLUTION" {
		status = "OPEN"
	}
	return observation{Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Claims: unresolvedClaims(reason, resolution, status, stage, step)}
}

func unresolvedClaims(reason, resolution, status, stage, step string) []ClaimResult {
	return []ClaimResult{
		{ID: "source-bytes-bound", Status: status, Resolution: resolution, Reason: reason, ProofChoice: "FOUNDATION", MetaOperation: "recheck-source-digest", Coordinate: Coordinate{stage, step, reason}, Provenance: "consumer-observation"},
		{ID: "operation-receipt-bound", Status: status, Resolution: resolution, Reason: reason, ProofChoice: "COHERENCE", MetaOperation: "recheck-operation-receipt", Coordinate: Coordinate{stage, step, reason}, Provenance: "consumer-observation"},
		{ID: "no-byte-authority", Status: status, Resolution: resolution, Reason: reason, ProofChoice: "REGRESSION", MetaOperation: "recheck-no-byte-authority", Coordinate: Coordinate{stage, step, reason}, Provenance: "consumer-observation"},
	}
}

func exactClaims(evidence []Evidence) []ClaimResult {
	claims := make([]ClaimResult, 0, len(evidence))
	for _, item := range evidence {
		claims = append(claims, ClaimResult{ID: item.ClaimID, Status: "DISCHARGED", Resolution: "EXACT", Reason: "CLAIM_DISCHARGED",
			ProofChoice: item.ProofChoice, MetaOperation: item.MetaOperation, Coordinate: item.Coordinate,
			EvidenceDigest: item.EvidenceDigest, Provenance: "consumer-canonical-recipe-v1"})
	}
	return claims
}
