package languageproofartifactverifier

type observation struct {
	Decision        string
	Resolution      string
	Reason          string
	Coordinate      Coordinate
	ArtifactDigest  string
	SourceDigest    string
	OperationDigest string
	Claims          []ClaimResult
}

func failure(resolution, reason, stage, step string) observation {
	return observation{Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Claims: unresolvedClaims(reason, stage, step)}
}

func unresolvedClaims(reason, stage, step string) []ClaimResult {
	return []ClaimResult{
		{ID: "source-bytes-bound", Status: "NOT_DISCHARGED", Reason: reason, ProofChoice: "FOUNDATION", MetaOperation: "recheck-source-digest", Coordinate: Coordinate{stage, step, reason}},
		{ID: "operation-receipt-bound", Status: "NOT_DISCHARGED", Reason: reason, ProofChoice: "COHERENCE", MetaOperation: "recheck-operation-receipt", Coordinate: Coordinate{stage, step, reason}},
		{ID: "no-byte-authority", Status: "NOT_DISCHARGED", Reason: reason, ProofChoice: "REGRESSION", MetaOperation: "recheck-no-byte-authority", Coordinate: Coordinate{stage, step, reason}},
	}
}

func exactClaims(evidence []Evidence) []ClaimResult {
	claims := make([]ClaimResult, 0, len(evidence))
	for _, item := range evidence {
		claims = append(claims, ClaimResult{ID: item.ClaimID, Status: "DISCHARGED", Reason: "CLAIM_PRESERVED",
			ProofChoice: item.ProofChoice, MetaOperation: item.MetaOperation, Coordinate: item.Coordinate,
			EvidenceDigest: item.EvidenceDigest})
	}
	return claims
}
