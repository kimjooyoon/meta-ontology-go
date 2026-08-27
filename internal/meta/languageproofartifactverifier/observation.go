package languageproofartifactverifier

type observation struct {
	Decision                  string
	Resolution                string
	Reason                    string
	Coordinate                Coordinate
	ArtifactDigest            string
	SourceDigest              string
	SemanticDigest            string
	OperationDigest           string
	OperationAttachmentDigest string
	RecipeAttachmentDigest    string
	EvidenceLinkDigest        string
	ClaimTransitionDigest     string
	ConsumerTargetDigest      string
	ConsumerOutputDigest      string
	ConsumerOutputExists      bool
	ConsumerErrorClass        string
	ConsumerErrorDigest       string
	Claims                    []ClaimResult
	Policy                    CaseEnvelopePolicyObservation
}

// ClaimAdjudication is the independent, claim-local result of checking one
// proposition. Case-level coordinates are only an aggregate failure view;
// they must never replace this evidence and causal coordinate.
type ClaimAdjudication struct {
	ClaimID         string
	Status          string
	Resolution      string
	Reason          string
	Coordinate      Coordinate
	EvidenceDigests []string
	Provenance      string
}

func failure(resolution, reason, stage, step string) observation {
	return failureWithAdjudications(nil, defaultClaimAdjudications(resolution, reason, Coordinate{Stage: stage, Step: step, Reason: reason}), resolution, reason, stage, step)
}

func failureWithAdjudications(statements []ClaimStatement, adjudications []ClaimAdjudication, resolution, reason, stage, step string) observation {
	return observation{Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}, Claims: claimsFromAdjudications(statements, adjudications)}
}

func defaultClaimAdjudications(resolution, reason string, coordinate Coordinate) []ClaimAdjudication {
	adjudications := make([]ClaimAdjudication, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		adjudications = append(adjudications, ClaimAdjudication{ClaimID: spec.ID, Status: "OPEN", Resolution: "LOWER_RESOLUTION", Reason: "CLAIM_PENDING", Coordinate: coordinate, Provenance: "consumer-observation"})
	}
	return adjudications
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
		{ID: "case-envelope-policy-bound", Proposition: "case-envelope-policy-matches-source", TargetDigest: "", Dependencies: []string{"source-bytes-bound"}, ProofChoice: "FOUNDATION", MetaOperation: "bind-case-envelope-policy", Coordinate: Coordinate{"CONSUME_POLICY", "policy-evidence", "CASE_ENVELOPE_POLICY_RECONSTRUCTED"}},
		{ID: "consumer-authority", Proposition: "verified-consumer-may-read-only-consume", TargetDigest: "READ_ONLY_CONSUMPTION", Dependencies: []string{"source-bytes-bound", "operation-receipt-bound", "no-byte-authority", "recipe-match", "case-envelope-policy-bound"}, ProofChoice: "COHERENCE", MetaOperation: "grant-read-only-consumption", Coordinate: Coordinate{"CONSUME_AUTHORITY", "authority-evidence", "AUTHORITY_REQUIRES_ATTESTATION"}},
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

func claimsFromAdjudications(statements []ClaimStatement, adjudications []ClaimAdjudication) []ClaimResult {
	claims := make([]ClaimResult, 0, ClaimTemplateTotal)
	for index, spec := range claimSpecs() {
		if index < len(statements) && statements[index].ID == spec.ID {
			statement := statements[index]
			spec = claimSpec{ID: statement.ID, Proposition: statement.Proposition, TargetDigest: statement.TargetDigest, Dependencies: statement.Dependencies, ProofChoice: statement.ProofChoice, MetaOperation: statement.MetaOperation, Coordinate: statement.Coordinate}
		}
		adjudication, ok := findClaimAdjudication(adjudications, spec.ID)
		if !ok {
			adjudication = ClaimAdjudication{ClaimID: spec.ID, Status: "OPEN", Resolution: "LOWER_RESOLUTION", Reason: "CLAIM_PENDING", Coordinate: Coordinate{"CONSUME", "claim-adjudication", "CLAIM_ADJUDICATION_NOT_OBSERVED"}, Provenance: "consumer-observation"}
		}
		claims = append(claims, makeClaimResult(spec, adjudication.Status, adjudication.Resolution, adjudication.Reason, adjudication.Coordinate, adjudication.EvidenceDigests, adjudication.Provenance))
	}
	return claims
}

func findClaimAdjudication(adjudications []ClaimAdjudication, claimID string) (ClaimAdjudication, bool) {
	for _, adjudication := range adjudications {
		if adjudication.ClaimID == claimID {
			return adjudication, true
		}
	}
	return ClaimAdjudication{}, false
}

func claimAdjudication(statements []ClaimStatement, claimID, status, resolution, reason string, coordinate Coordinate, provenance string) ClaimAdjudication {
	result := ClaimAdjudication{ClaimID: claimID, Status: status, Resolution: resolution, Reason: reason, Coordinate: coordinate, Provenance: provenance}
	for _, statement := range statements {
		if statement.ID == claimID {
			result.EvidenceDigests = append([]string(nil), statement.EvidenceDigest...)
			break
		}
	}
	return result
}

func claimStatus(adjudications []ClaimAdjudication, claimID string) string {
	if adjudication, ok := findClaimAdjudication(adjudications, claimID); ok {
		return adjudication.Status
	}
	return "OPEN"
}

func claimCoordinate(claimID string) Coordinate {
	for _, spec := range claimSpecs() {
		if spec.ID == claimID {
			return spec.Coordinate
		}
	}
	return Coordinate{"CONSUME", "claim-adjudication", "CLAIM_ADJUDICATION_NOT_OBSERVED"}
}

func claimStatementCoordinate(statements []ClaimStatement, claimID string) Coordinate {
	for _, statement := range statements {
		if statement.ID == claimID && statement.Coordinate.Stage != "" && statement.Coordinate.Step != "" && statement.Coordinate.Reason != "" {
			return statement.Coordinate
		}
	}
	return claimCoordinate(claimID)
}

func authorityFailureAdjudications(statements []ClaimStatement, status, resolution, reason string, coordinate Coordinate) []ClaimAdjudication {
	adjudications := make([]ClaimAdjudication, 0, ClaimTemplateTotal)
	for _, spec := range claimSpecs() {
		if spec.ID == "consumer-authority" {
			adjudications = append(adjudications, claimAdjudication(statements, spec.ID, status, resolution, reason, coordinate, "consumer-observation"))
			continue
		}
		adjudications = append(adjudications, claimAdjudication(statements, spec.ID, "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(statements, spec.ID), "consumer-canonical-recipe-v2"))
	}
	return adjudications
}

func exactClaims(statements []ClaimStatement) []ClaimResult {
	claims := make([]ClaimResult, 0, len(statements))
	for _, statement := range statements {
		claims = append(claims, makeClaimResult(claimSpec{ID: statement.ID, Proposition: statement.Proposition, TargetDigest: statement.TargetDigest, Dependencies: statement.Dependencies, ProofChoice: statement.ProofChoice, MetaOperation: statement.MetaOperation, Coordinate: statement.Coordinate}, "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", statement.Coordinate, statement.EvidenceDigest, "consumer-canonical-recipe-v2"))
	}
	return claims
}
