package languageproofartifactverifier

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// verifyArtifact is the independent consumer kernel. It evaluates each
// proposition separately and propagates failure only over declared edges.
func verifyArtifact(raw, source, operation, recipe []byte, head, phase string) observation {
	artifact, err := decodeStrict[Artifact](raw)
	if err != nil {
		return failure("INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", "CONSUME_DECODE", "artifact")
	}
	artifactEvidenceDigest := artifact.Digest
	if artifact.Digest != artifactDigest(artifact) {
		result := failure("INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_DIGEST_MISMATCH", "CONSUME_DIGEST", "artifact")
		result.ArtifactDigest = artifactEvidenceDigest
		return result
	}
	if artifact.HeadSHA != head {
		adjudications := make([]ClaimAdjudication, 0, ClaimTemplateTotal)
		for _, spec := range claimSpecs() {
			if spec.ID == "consumer-authority" {
				adjudications = append(adjudications, claimAdjudication(artifact.Claims, spec.ID, "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_IDENTITY", "head", "HEAD_BINDING_MISMATCH"}, "consumer-observation"))
				continue
			}
			adjudications = append(adjudications, claimAdjudication(artifact.Claims, spec.ID, "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", spec.Coordinate, "consumer-canonical-recipe-v2"))
		}
		return observedFailure(artifact, "INVARIANT_ONLY", "HEAD_BINDING_MISMATCH", "CONSUME_IDENTITY", "head", adjudications, artifactEvidenceDigest, "", "", "")
	}

	identityOK := artifact.Schema == ArtifactSchema && artifact.HeadSHA == head && artifact.Producer == ProducerID && artifact.Consumer == ConsumerID && artifact.MetaOperation == "emit-proof-carrying-artifact" &&
		artifact.Decision == "CARRIED" && artifact.Resolution == "EVIDENCE_ATTACHED" && artifact.Reason == "PROOF_CARRYING_ARTIFACT_EMITTED" &&
		artifact.Authority.ArtifactUseAuthority == "NONE" && !artifact.Authority.CapabilityMutationGranted && !artifact.Authority.PromotionAuthority && !artifact.Authority.SemanticAuthority &&
		artifact.Authority.Basis == "INDEPENDENT_CONSUMER_VERIFICATION_REQUIRED" && artifact.SourcePath != "" && validDigest(artifact.SourceDigest) && validDigest(artifact.SemanticDigest) && validDigest(artifact.OperationDigest) &&
		artifact.Recipe.Version == 2 && artifact.RecipeDigest == digestValue(artifact.Recipe) && validateWriteSet(artifact.WriteSet) == nil
	if !identityOK {
		return observedFailure(artifact, "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", "CONSUME_IDENTITY", "artifact", defaultClaimAdjudications("INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", Coordinate{"CONSUME_IDENTITY", "artifact", "PROOF_CARRYING_ARTIFACT_INVALID"}), artifactEvidenceDigest, "", "", "")
	}

	byKind := map[string]Evidence{}
	evidenceOK := map[string]bool{}
	evidenceDigestMismatch := false
	mismatchedEvidenceKinds := map[string]bool{}
	duplicateEvidence := false
	for _, item := range artifact.Evidence {
		if item.Kind == "" || byKind[item.Kind].Kind != "" {
			duplicateEvidence = true
			continue
		}
		byKind[item.Kind] = item
		evidenceOK[item.Kind] = item.EvidenceDigest == evidenceDigest(item)
		if !evidenceOK[item.Kind] {
			evidenceDigestMismatch = true
			mismatchedEvidenceKinds[item.Kind] = true
		}
	}
	claimsStructureOK := validateClaimStatements(artifact.Claims, artifact.Evidence, artifact)
	priorLedgerOK := validatePriorLedger(artifact.PriorLedger, artifact.Claims) == nil

	sourceDigest := digestBytes(source)
	projection, projectionErr := projectSource(source, activityFrom(artifact))
	sourceClaimOK := claimOK(artifact.Claims, artifact.Evidence, artifact, "source-bytes-bound")
	sourceStatementOK := claimStatementOK(artifact.Claims, artifact, "source-bytes-bound")
	sourceBindingOK := len(source) > 0 && projectionErr == nil && artifact.SourceDigest == sourceDigest && artifact.SemanticDigest == projection.SemanticDigest
	sourceGood := sourceBindingOK && evidenceOK["SOURCE"] && sourceClaimOK
	adjudications := make([]ClaimAdjudication, 0, ClaimTemplateTotal)
	if sourceGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "source-bytes-bound", "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(artifact.Claims, "source-bytes-bound"), "consumer-canonical-recipe-v2"))
	} else if len(source) == 0 || projectionErr != nil {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "source-bytes-bound", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", Coordinate{"CONSUME_SOURCE", "reconstruct", "SOURCE_RECONSTRUCTION_NOT_OBSERVED"}, "consumer-observation"))
	} else if !sourceStatementOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "source-bytes-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"}, "consumer-observation"))
	} else if !evidenceOK["SOURCE"] || !sourceClaimOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "source-bytes-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_EVIDENCE", "source-evidence", "PROOF_EVIDENCE_DIGEST_MISMATCH"}, "consumer-observation"))
	} else {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "source-bytes-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_SOURCE", "reconstruct", "SOURCE_RECONSTRUCTION_MISMATCH"}, "consumer-observation"))
	}

	var receipt operationReceipt
	operationMissing := len(operation) == 0
	operationDecoded := !operationMissing && decodeInto(operation, &receipt) == nil
	operationDigestOK := operationDecoded && receipt.Digest == receiptDigest(receipt)
	operationClaimOK := claimOK(artifact.Claims, artifact.Evidence, artifact, "operation-receipt-bound")
	operationStatementOK := claimStatementOK(artifact.Claims, artifact, "operation-receipt-bound")
	operationEvidenceMissing := len(byKind["OPERATION"].Kind) == 0
	operationGood := operationDecoded && operationDigestOK && sourceGood && verifyOperation(receipt, sourceDigest, artifact.SourcePath, projection) && artifact.OperationDigest == receipt.Digest &&
		evidenceOK["OPERATION"] && operationClaimOK
	if operationGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(artifact.Claims, "operation-receipt-bound"), "consumer-canonical-recipe-v2"))
	} else if operationMissing || !sourceGood {
		coordinate := Coordinate{"CONSUME_OPERATION", "receipt", "OPERATION_RECONSTRUCTION_NOT_OBSERVED"}
		if operationMissing {
			coordinate = Coordinate{"CONSUME_INPUT", "operation-attachment", "ARTIFACT_ATTACHMENT_MISSING"}
		}
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", coordinate, "consumer-observation"))
	} else if operationEvidenceMissing {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_EVIDENCE", "operation-evidence", "PROOF_EVIDENCE_MISSING"}, "consumer-observation"))
	} else if !operationStatementOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"}, "consumer-observation"))
	} else if !evidenceOK["OPERATION"] || !operationClaimOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_EVIDENCE", "operation-evidence", "PROOF_EVIDENCE_DIGEST_MISMATCH"}, "consumer-observation"))
	} else if operationDecoded && !operationDigestOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_OPERATION", "attachment-digest", "OPERATION_ATTACHMENT_DIGEST_MISMATCH"}, "consumer-observation"))
	} else {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "operation-receipt-bound", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_OPERATION", "receipt", "OPERATION_RECONSTRUCTION_MISMATCH"}, "consumer-observation"))
	}

	invariantClaimOK := claimOK(artifact.Claims, artifact.Evidence, artifact, "no-byte-authority")
	invariantStatementOK := claimStatementOK(artifact.Claims, artifact, "no-byte-authority")
	invariantGood := evidenceOK["INVARIANT"] && invariantClaimOK && artifact.WriteSet.NetChangedPaths == 0 &&
		!artifact.WriteSet.CapabilityMutationGranted && artifact.Effects.NetChangedPaths == 0 && !artifact.Effects.CapabilityMutationGranted && artifact.Authority.ArtifactUseAuthority == "NONE"
	if invariantGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "no-byte-authority", "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(artifact.Claims, "no-byte-authority"), "consumer-canonical-recipe-v2"))
	} else if !invariantStatementOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "no-byte-authority", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"}, "consumer-observation"))
	} else if !evidenceOK["INVARIANT"] || !invariantClaimOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "no-byte-authority", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_INVARIANT", "invariant-evidence", "INVARIANT_EVIDENCE_NOT_PRESERVED"}, "consumer-observation"))
	} else {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "no-byte-authority", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_INVARIANT", "invariant-evidence", "INVARIANT_POLICY_NOT_PRESERVED"}, "consumer-observation"))
	}

	externalRecipe, recipeErr := decodeRecipe(recipe)
	derivedRecipe, derivedErr := recipeFromSource(source)
	recipeEvidenceGood := claimOK(artifact.Claims, artifact.Evidence, artifact, "recipe-match")
	recipeStatementOK := claimStatementOK(artifact.Claims, artifact, "recipe-match")
	recipeGood := recipeErr == nil && derivedErr == nil && reflect.DeepEqual(externalRecipe, CanonicalRecipe()) && reflect.DeepEqual(derivedRecipe, CanonicalRecipe()) &&
		reflect.DeepEqual(artifact.Recipe, externalRecipe) && artifact.RecipeDigest == digestValue(externalRecipe) && recipeEvidenceGood
	noExternalEvidence := len(source) == 0 && len(operation) == 0 && len(recipe) == 0
	invariantEvidenceOnly := mismatchedEvidenceKinds["INVARIANT"] && !mismatchedEvidenceKinds["SOURCE"] && !mismatchedEvidenceKinds["OPERATION"]
	if recipeGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(artifact.Claims, "recipe-match"), "consumer-canonical-recipe-v2"))
	} else if !recipeStatementOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"}, "consumer-observation"))
	} else if noExternalEvidence {
		coordinate := Coordinate{"CONSUME_INPUT", "external-evidence", "ARTIFACT_BYTES_NOT_AUTHORITY"}
		if phase == ProofPhasePreliminary {
			adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", coordinate, "consumer-observation"))
		} else {
			adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", coordinate, "consumer-observation"))
		}
	} else if operationMissing {
		// An absent external operation attachment leaves the dependent recipe
		// proposition unresolved in either phase. The only declared phase
		// reclassification is the byte-only case below.
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_OPERATION_ATTACHMENT_NOT_OBSERVED"}, "consumer-observation"))
	} else if invariantEvidenceOnly {
		// An unrelated invariant evidence failure does not contradict the
		// independently reconstructed recipe; its dependency remains OPEN.
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_INVARIANT_EVIDENCE_NOT_RESOLVED"}, "consumer-observation"))
	} else if operationEvidenceMissing {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_OPERATION_EVIDENCE_MISSING"}, "consumer-observation"))
	} else if !recipeEvidenceGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_RECIPE", "recipe-evidence", "RECIPE_EVIDENCE_NOT_PRESERVED"}, "consumer-observation"))
	} else {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "recipe-match", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_RECIPE", "recipe", "INDEPENDENT_RECIPE_MISMATCH"}, "consumer-observation"))
	}

	allPrerequisites := claimsStructureOK && priorLedgerOK && claimStatus(adjudications, "source-bytes-bound") == "DISCHARGED" && claimStatus(adjudications, "operation-receipt-bound") == "DISCHARGED" && claimStatus(adjudications, "no-byte-authority") == "DISCHARGED" && claimStatus(adjudications, "recipe-match") == "DISCHARGED"
	authorityClaimOK := claimOK(artifact.Claims, artifact.Evidence, artifact, "consumer-authority")
	authorityGood := allPrerequisites && authorityClaimOK && len(raw) > 0
	if authorityGood {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "consumer-authority", "DISCHARGED", "EXACT", "CLAIM_DISCHARGED", claimStatementCoordinate(artifact.Claims, "consumer-authority"), "consumer-canonical-recipe-v2"))
	} else if !allPrerequisites {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "consumer-authority", "OPEN", "LOWER_RESOLUTION", "CLAIM_PENDING", Coordinate{"CONSUME_AUTHORITY", "authority-evidence", "AUTHORITY_DEPENDENCY_OPEN"}, "consumer-observation"))
	} else if !authorityClaimOK {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "consumer-authority", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_CLAIMS", "claim-statement", "PROOF_CLAIM_STATEMENT_MISMATCH"}, "consumer-observation"))
	} else {
		adjudications = append(adjudications, claimAdjudication(artifact.Claims, "consumer-authority", "REFUTED", "INVARIANT_ONLY", "CLAIM_REFUTED", Coordinate{"CONSUME_AUTHORITY", "authority-evidence", "AUTHORITY_ATTESTATION_NOT_PRESERVED"}, "consumer-observation"))
	}

	resolution, reason, stage, step := "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", "CONSUME_IDENTITY", "artifact"
	switch {
	case len(source) == 0 && len(operation) == 0 && len(recipe) == 0:
		resolution, reason, stage, step = "LOWER_RESOLUTION", "ARTIFACT_BYTES_NOT_AUTHORITY", "CONSUME_INPUT", "external-evidence"
	case len(byKind["OPERATION"].Kind) == 0:
		resolution, reason, stage, step = "LOWER_RESOLUTION", "PROOF_EVIDENCE_MISSING", "CONSUME_EVIDENCE", "operation-evidence"
	case operationMissing:
		resolution, reason, stage, step = "LOWER_RESOLUTION", "ARTIFACT_ATTACHMENT_MISSING", "CONSUME_INPUT", "operation-attachment"
	case evidenceDigestMismatch || duplicateEvidence:
		resolution, reason, stage, step = "INVARIANT_ONLY", "PROOF_EVIDENCE_DIGEST_MISMATCH", "CONSUME_EVIDENCE", "evidence-digest"
		if mismatchedEvidenceKinds["INVARIANT"] && !mismatchedEvidenceKinds["SOURCE"] && !mismatchedEvidenceKinds["OPERATION"] {
			reason, step = "INVARIANT_EVIDENCE_NOT_PRESERVED", "invariant-evidence"
		}
	case !priorLedgerOK:
		resolution, reason, stage, step = "INVARIANT_ONLY", "PROOF_EVIDENCE_LEDGER_MISMATCH", "CONSUME_LEDGER", "prior-ledger"
	case !claimsStructureOK:
		resolution, reason, stage, step = "INVARIANT_ONLY", "PROOF_CLAIM_STATEMENT_MISMATCH", "CONSUME_CLAIMS", "claim-statement"
	case !sourceGood:
		resolution, reason, stage, step = "INVARIANT_ONLY", "SOURCE_RECONSTRUCTION_MISMATCH", "CONSUME_SOURCE", "reconstruct"
	case !operationGood && operationDecoded && !operationDigestOK:
		resolution, reason, stage, step = "INVARIANT_ONLY", "OPERATION_ATTACHMENT_DIGEST_MISMATCH", "CONSUME_OPERATION", "attachment-digest"
	case !operationGood:
		resolution, reason, stage, step = "INVARIANT_ONLY", "OPERATION_RECONSTRUCTION_MISMATCH", "CONSUME_OPERATION", "receipt"
	case !invariantGood:
		resolution, reason, stage, step = "INVARIANT_ONLY", "INVARIANT_EVIDENCE_NOT_PRESERVED", "CONSUME_INVARIANT", "invariant-evidence"
	case !recipeGood:
		resolution, reason, stage, step = "INVARIANT_ONLY", "INDEPENDENT_RECIPE_MISMATCH", "CONSUME_RECIPE", "recipe"
	}
	if authorityGood {
		claims := exactClaims(artifact.Claims)
		result := observation{Decision: "PASS", Resolution: "EXACT", Reason: "PROOF_CARRYING_ARTIFACT_AUTHORIZED", Coordinate: Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", "CONSUMER_ONLY_READ_ONLY_AUTHORITY"}, ArtifactDigest: artifactEvidenceDigest, SourceDigest: sourceDigest, SemanticDigest: projection.SemanticDigest, OperationDigest: receipt.Digest, Claims: claims}
		result.EvidenceLinkDigest = digestValue(statementEvidence(artifact.Claims))
		result.ClaimTransitionDigest = digestValue(claimTransitions(claims))
		result.OperationAttachmentDigest = digestBytes(operation)
		result.RecipeAttachmentDigest = digestBytes(recipe)
		return result
	}
	return observedFailure(artifact, resolution, reason, stage, step, adjudications, artifactEvidenceDigest, sourceDigest, projection.SemanticDigest, receiptDigestIfValid(receipt))
}

func observedFailure(artifact Artifact, resolution, reason, stage, step string, adjudications []ClaimAdjudication, artifactDigestValue, sourceDigest, semanticDigest, operationDigest string) observation {
	result := failureWithAdjudications(artifact.Claims, adjudications, resolution, reason, stage, step)
	result.ArtifactDigest, result.SourceDigest, result.SemanticDigest, result.OperationDigest = artifactDigestValue, sourceDigest, semanticDigest, operationDigest
	result.EvidenceLinkDigest = digestValue(statementEvidence(artifact.Claims))
	result.ClaimTransitionDigest = digestValue(claimTransitions(result.Claims))
	return result
}

func validateClaimStatements(claims []ClaimStatement, evidence []Evidence, artifact Artifact) bool {
	if len(claims) != ClaimTemplateTotal {
		return false
	}
	byID := map[string]ClaimStatement{}
	for _, claim := range claims {
		if _, exists := byID[claim.ID]; exists || claim.Digest != claimStatementDigest(claim) || claim.ID == "" || claim.Proposition == "" {
			return false
		}
		byID[claim.ID] = claim
	}
	wantTargets := map[string]string{
		"source-bytes-bound":      artifact.SourceDigest,
		"operation-receipt-bound": artifact.OperationDigest,
		"no-byte-authority":       "READ_ONLY_CONSUMPTION",
		"recipe-match":            artifact.RecipeDigest,
		"consumer-authority":      "READ_ONLY_CONSUMPTION",
	}
	for _, spec := range claimSpecs() {
		claim, ok := byID[spec.ID]
		if !ok || !claimStatementMatches(claim, artifact, spec) || claim.TargetDigest != wantTargets[spec.ID] {
			return false
		}
		for _, link := range claim.EvidenceDigest {
			if !validDigest(link) {
				return false
			}
		}
	}
	_ = evidence
	return true
}

func claimStatementMatches(claim ClaimStatement, artifact Artifact, spec claimSpec) bool {
	target := spec.TargetDigest
	switch spec.ID {
	case "source-bytes-bound":
		target = artifact.SourceDigest
	case "operation-receipt-bound":
		target = artifact.OperationDigest
	case "recipe-match":
		target = artifact.RecipeDigest
	}
	return claim.ID == spec.ID && claim.Digest == claimStatementDigest(claim) && claim.Proposition == spec.Proposition && claim.TargetDigest == target &&
		reflect.DeepEqual(claim.Dependencies, spec.Dependencies) && claim.ProofChoice == spec.ProofChoice && claim.MetaOperation == spec.MetaOperation &&
		claim.Coordinate.Stage != "" && claim.Coordinate.Step != "" && claim.Coordinate.Reason != "" && len(claim.EvidenceDigest) > 0
}

func claimOK(claims []ClaimStatement, evidence []Evidence, artifact Artifact, id string) bool {
	for _, claim := range claims {
		var spec claimSpec
		found := false
		for _, candidate := range claimSpecs() {
			if candidate.ID == id {
				spec, found = candidate, true
				break
			}
		}
		if !found || !claimStatementMatches(claim, artifact, spec) {
			continue
		}
		available := map[string]bool{}
		for _, item := range evidence {
			if item.EvidenceDigest == evidenceDigest(item) {
				available[item.EvidenceDigest] = true
			}
		}
		for _, link := range claim.EvidenceDigest {
			if !available[link] {
				return false
			}
		}
		return true
	}
	return false
}

func claimStatementOK(claims []ClaimStatement, artifact Artifact, id string) bool {
	for _, claim := range claims {
		for _, spec := range claimSpecs() {
			if spec.ID == id && claim.ID == id {
				return claimStatementMatches(claim, artifact, spec)
			}
		}
	}
	return false
}

func statementEvidence(claims []ClaimStatement) []string {
	result := make([]string, 0, len(claims))
	for _, claim := range claims {
		result = append(result, claim.EvidenceDigest...)
	}
	return result
}

func decodeRecipe(raw []byte) (Recipe, error) { return decodeStrict[Recipe](raw) }

func decodeInto(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) == nil {
		return &extraJSONError{}
	}
	return nil
}

func receiptDigestIfValid(receipt operationReceipt) string {
	if receipt.Digest != "" {
		return receipt.Digest
	}
	return ""
}

func activityFrom(artifact Artifact) string {
	for _, item := range artifact.Evidence {
		if item.Kind == "OPERATION" {
			return item.Activity
		}
	}
	return ""
}

type extraJSONError struct{}

func (e *extraJSONError) Error() string { return "trailing JSON" }
