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
		return observedFailure(artifact, "INVARIANT_ONLY", "HEAD_BINDING_MISMATCH", "CONSUME_IDENTITY", "head",
			map[string]string{
				"source-bytes-bound":      "DISCHARGED",
				"operation-receipt-bound": "DISCHARGED",
				"no-byte-authority":       "DISCHARGED",
				"recipe-match":            "DISCHARGED",
				"consumer-authority":      "REFUTED",
			}, artifactEvidenceDigest, "", "", "")
	}

	identityOK := artifact.Schema == ArtifactSchema && artifact.HeadSHA == head && artifact.Producer == ProducerID && artifact.Consumer == ConsumerID && artifact.MetaOperation == "emit-proof-carrying-artifact" &&
		artifact.Decision == "CARRIED" && artifact.Resolution == "EVIDENCE_ATTACHED" && artifact.Reason == "PROOF_CARRYING_ARTIFACT_EMITTED" &&
		artifact.Authority.ArtifactUseAuthority == "NONE" && !artifact.Authority.CapabilityMutationGranted && !artifact.Authority.PromotionAuthority && !artifact.Authority.SemanticAuthority &&
		artifact.Authority.Basis == "INDEPENDENT_CONSUMER_VERIFICATION_REQUIRED" && artifact.SourcePath != "" && validDigest(artifact.SourceDigest) && validDigest(artifact.SemanticDigest) && validDigest(artifact.OperationDigest) &&
		artifact.Recipe.Version == 2 && artifact.RecipeDigest == digestValue(artifact.Recipe) && validateWriteSet(artifact.WriteSet) == nil
	if !identityOK {
		return observedFailure(artifact, "INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", "CONSUME_IDENTITY", "artifact", map[string]string{}, artifactEvidenceDigest, "", "", "")
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

	statuses := map[string]string{}
	sourceDigest := digestBytes(source)
	projection, projectionErr := projectSource(source, activityFrom(artifact))
	sourceGood := len(source) > 0 && projectionErr == nil && artifact.SourceDigest == sourceDigest && artifact.SemanticDigest == projection.SemanticDigest &&
		evidenceOK["SOURCE"] && claimOK(artifact.Claims, artifact.Evidence, artifact, "source-bytes-bound")
	if sourceGood {
		statuses["source-bytes-bound"] = "DISCHARGED"
	} else if len(source) == 0 || projectionErr != nil {
		statuses["source-bytes-bound"] = "OPEN"
	} else {
		statuses["source-bytes-bound"] = "REFUTED"
	}

	var receipt operationReceipt
	operationMissing := len(operation) == 0
	operationDecoded := !operationMissing && decodeInto(operation, &receipt) == nil
	operationDigestOK := operationDecoded && receipt.Digest == receiptDigest(receipt)
	operationGood := operationDecoded && operationDigestOK && sourceGood && verifyOperation(receipt, sourceDigest, artifact.SourcePath, projection) && artifact.OperationDigest == receipt.Digest &&
		evidenceOK["OPERATION"] && claimOK(artifact.Claims, artifact.Evidence, artifact, "operation-receipt-bound")
	if operationGood {
		statuses["operation-receipt-bound"] = "DISCHARGED"
	} else if operationMissing || !sourceGood {
		statuses["operation-receipt-bound"] = "OPEN"
	} else {
		statuses["operation-receipt-bound"] = "REFUTED"
	}

	invariantGood := evidenceOK["INVARIANT"] && claimOK(artifact.Claims, artifact.Evidence, artifact, "no-byte-authority") && artifact.WriteSet.NetChangedPaths == 0 &&
		!artifact.WriteSet.CapabilityMutationGranted && artifact.Effects.NetChangedPaths == 0 && !artifact.Effects.CapabilityMutationGranted && artifact.Authority.ArtifactUseAuthority == "NONE"
	if invariantGood {
		statuses["no-byte-authority"] = "DISCHARGED"
	} else {
		statuses["no-byte-authority"] = "REFUTED"
	}

	externalRecipe, recipeErr := decodeRecipe(recipe)
	derivedRecipe, derivedErr := recipeFromSource(source)
	recipeEvidenceGood := claimOK(artifact.Claims, artifact.Evidence, artifact, "recipe-match")
	recipeGood := recipeErr == nil && derivedErr == nil && reflect.DeepEqual(externalRecipe, CanonicalRecipe()) && reflect.DeepEqual(derivedRecipe, CanonicalRecipe()) &&
		reflect.DeepEqual(artifact.Recipe, externalRecipe) && artifact.RecipeDigest == digestValue(externalRecipe) && recipeEvidenceGood
	if recipeGood {
		statuses["recipe-match"] = "DISCHARGED"
	} else if phase == ProofPhasePreliminary && (operationMissing || (mismatchedEvidenceKinds["INVARIANT"] && !mismatchedEvidenceKinds["SOURCE"] && !mismatchedEvidenceKinds["OPERATION"])) {
		// A missing operation attachment or an invariant-only evidence change
		// leaves the recipe claim unresolved. The recipe itself has not been
		// contradicted, so this declared dependency remains OPEN rather than
		// being blanket-promoted to REFUTED.
		statuses["recipe-match"] = "OPEN"
	} else {
		statuses["recipe-match"] = "REFUTED"
	}

	allPrerequisites := claimsStructureOK && priorLedgerOK && statuses["source-bytes-bound"] == "DISCHARGED" && statuses["operation-receipt-bound"] == "DISCHARGED" && statuses["no-byte-authority"] == "DISCHARGED" && statuses["recipe-match"] == "DISCHARGED"
	authorityGood := allPrerequisites && claimOK(artifact.Claims, artifact.Evidence, artifact, "consumer-authority") && len(raw) > 0
	if authorityGood {
		statuses["consumer-authority"] = "DISCHARGED"
	} else if !allPrerequisites {
		statuses["consumer-authority"] = "OPEN"
	} else {
		statuses["consumer-authority"] = "REFUTED"
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
	return observedFailure(artifact, resolution, reason, stage, step, statuses, artifactEvidenceDigest, sourceDigest, projection.SemanticDigest, receiptDigestIfValid(receipt))
}

func observedFailure(artifact Artifact, resolution, reason, stage, step string, statuses map[string]string, artifactDigestValue, sourceDigest, semanticDigest, operationDigest string) observation {
	result := failureWithStates(resolution, reason, stage, step, statuses)
	result.ArtifactDigest, result.SourceDigest, result.SemanticDigest, result.OperationDigest = artifactDigestValue, sourceDigest, semanticDigest, operationDigest
	result.Claims = claimsFromStatements(artifact.Claims, statuses, resolution, reason, Coordinate{stage, step, reason})
	result.EvidenceLinkDigest = digestValue(statementEvidence(artifact.Claims))
	result.ClaimTransitionDigest = digestValue(claimTransitions(result.Claims))
	return result
}

func claimsFromStatements(statements []ClaimStatement, statuses map[string]string, resolution, reason string, coordinate Coordinate) []ClaimResult {
	if len(statements) != ClaimTemplateTotal {
		return failureClaims(statuses, resolution, reason, coordinate.Stage, coordinate.Step)
	}
	result := make([]ClaimResult, 0, len(statements))
	for _, statement := range statements {
		status := statuses[statement.ID]
		if status == "" {
			status = "OPEN"
		}
		claimResolution, claimReason, provenance := resolution, reason, "consumer-observation"
		if status == "DISCHARGED" {
			claimResolution, claimReason, provenance = "EXACT", "CLAIM_DISCHARGED", "consumer-canonical-recipe-v2"
		} else if status == "OPEN" {
			claimResolution = "LOWER_RESOLUTION"
		} else if status == "REFUTED" {
			// A case may be lower-resolution because another attachment is
			// missing, while this claim is independently contradicted by its
			// absent or invalid evidence. Keep the claim-level resolution
			// invariant-only instead of inheriting the case-level resolution.
			claimResolution, claimReason = "INVARIANT_ONLY", "CLAIM_REFUTED"
		}
		result = append(result, makeClaimResult(claimSpec{ID: statement.ID, Proposition: statement.Proposition, TargetDigest: statement.TargetDigest, Dependencies: statement.Dependencies, ProofChoice: statement.ProofChoice, MetaOperation: statement.MetaOperation, Coordinate: statement.Coordinate}, status, claimResolution, claimReason, coordinate, statement.EvidenceDigest, provenance))
	}
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
