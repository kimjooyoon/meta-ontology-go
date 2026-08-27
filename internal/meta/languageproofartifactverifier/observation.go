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
	return failureWithStates(resolution, reason, stage, step, map[string]string{})
}

func failureWithStates(resolution, reason, stage, step string, states map[string]string) observation {
	return observation{Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Claims: failureClaims(states, resolution, reason, stage, step)}
}

func failureClaims(states map[string]string, resolution, reason, stage, step string) []ClaimResult {
	claims := make([]ClaimResult, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		status := states[spec.ID]
		if status == "" {
			status = "OPEN"
		}
		claimResolution := resolution
		if status == "DISCHARGED" {
			claimResolution = "EXACT"
		}
		claimReason := reason
		if status == "DISCHARGED" {
			claimReason = "CLAIM_DISCHARGED"
		}
		claims = append(claims, makeClaimResult(spec, status, claimResolution, claimReason, Coordinate{stage, step, claimReason}, nil, "consumer-observation"))
	}
	return claims
}

type claimSpec struct {
	ID            string
	Proposition   string
	TargetDigest  string
	Dependencies  []string
	ProofChoice   string
	MetaOperation string
	Coordinate    Coordinate
}

func claimSpecs() []claimSpec {
	return []claimSpec{
		{ID: "source-bytes-bound", Proposition: "source-bytes-match", TargetDigest: "", ProofChoice: "FOUNDATION", MetaOperation: "bind-source-bytes", Coordinate: Coordinate{"CONSUME_SOURCE", "source-evidence", "SOURCE_RECONSTRUCTED"}},
		{ID: "operation-receipt-bound", Proposition: "operation-receipt-match", TargetDigest: "", Dependencies: []string{"source-bytes-bound"}, ProofChoice: "COHERENCE", MetaOperation: "bind-operation-receipt", Coordinate: Coordinate{"CONSUME_OPERATION", "operation-evidence", "OPERATION_RECONSTRUCTED"}},
		{ID: "no-byte-authority", Proposition: "generated-bytes-do-not-grant-authority", TargetDigest: "READ_ONLY_CONSUMPTION", ProofChoice: "REGRESSION", MetaOperation: "preserve-no-byte-authority", Coordinate: Coordinate{"CONSUME_INVARIANT", "invariant-evidence", "NO_BYTE_AUTHORITY"}},
		{ID: "recipe-match", Proposition: "consumer-recipe-matches-source-recipe", TargetDigest: "", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority"}, ProofChoice: "COHERENCE", MetaOperation: "match-independent-recipe", Coordinate: Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_RECONSTRUCTED"}},
		{ID: "consumer-authority", Proposition: "verified-consumer-may-read-only-consume", TargetDigest: "READ_ONLY_CONSUMPTION", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match"}, ProofChoice: "COHERENCE", MetaOperation: "grant-read-only-consumption", Coordinate: Coordinate{"CONSUME_AUTHORITY", "authority-evidence", "AUTHORITY_REQUIRES_ATTESTATION"}},
	}
}

func makeClaimResult(spec claimSpec, status, resolution, reason string, coordinate Coordinate, evidence []string, provenance string) ClaimResult {
	result := ClaimResult{ID: spec.ID, Proposition: spec.Proposition, TargetDigest: spec.TargetDigest, Dependencies: append([]string(nil), spec.Dependencies...), Status: status, Resolution: resolution, Reason: reason, ProofChoice: spec.ProofChoice, MetaOperation: spec.MetaOperation, Coordinate: coordinate, Provenance: provenance}
	if len(evidence) > 0 {
		result.EvidenceDigest = evidence[0]
		result.EvidenceDigests = append([]string(nil), evidence...)
	}
	result.StateDigest = claimStateDigest(result)
	return result
}

func exactClaims(statements []ClaimStatement) []ClaimResult {
	claims := make([]ClaimResult, 0, len(statements))
	for _, statement := range statements {
		claims = append(claims, makeClaimResult(claimSpec{ID: statement.ID, Proposition: statement.Proposition, TargetDigest: statement.TargetDigest, Dependencies: statement.Dependencies, ProofChoice: statement.ProofChoice, MetaOperation: statement.MetaOperation, Coordinate: statement.Coordinate}, "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", statement.Coordinate, statement.EvidenceDigest, "consumer-canonical-recipe-v2"))
	}
	return claims
}
