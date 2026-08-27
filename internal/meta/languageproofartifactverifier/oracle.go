package languageproofartifactverifier

import "reflect"

func verifyArtifact(raw, source, operation, recipe []byte, head string) observation {
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
	if artifact.Schema != ArtifactSchema || artifact.HeadSHA != head || artifact.Producer != ProducerID ||
		artifact.Consumer != ConsumerID || artifact.MetaOperation != "emit-proof-carrying-artifact" ||
		artifact.Decision != "CARRIED" || artifact.Resolution != "EVIDENCE_ATTACHED" ||
		artifact.Reason != "PROOF_CARRYING_ARTIFACT_EMITTED" || artifact.Authority.ArtifactUseAuthority != "NONE" ||
		artifact.Authority.MutationAuthority || artifact.Authority.PromotionAuthority || artifact.Authority.SemanticAuthority ||
		artifact.Authority.Basis != "INDEPENDENT_CONSUMER_VERIFICATION_REQUIRED" || artifact.SourcePath == "" ||
		!validDigest(artifact.SourceDigest) || !validDigest(artifact.SemanticDigest) || artifact.Recipe.Version != 1 ||
		artifact.RecipeDigest != digestValue(artifact.Recipe) || validateWriteSet(artifact.WriteSet) != nil ||
		artifact.Effects.RepositoryWrites != artifact.WriteSet.RepositoryWrites || artifact.Effects.MutationAuthority != artifact.WriteSet.MutationAuthority {
		return failure("INVARIANT_ONLY", "PROOF_CARRYING_ARTIFACT_INVALID", "CONSUME_IDENTITY", "artifact")
	}
	if len(source) == 0 || len(operation) == 0 || len(recipe) == 0 {
		result := failure("LOWER_RESOLUTION", "ARTIFACT_BYTES_NOT_AUTHORITY", "CONSUME_INPUT", "external-evidence")
		result.ArtifactDigest, result.SourceDigest = artifactEvidenceDigest, digestBytes(source)
		return result
	}

	externalRecipe, err := decodeStrict[Recipe](recipe)
	if err != nil || !reflect.DeepEqual(externalRecipe, CanonicalRecipe()) || !reflect.DeepEqual(artifact.Recipe, externalRecipe) {
		return failure("INVARIANT_ONLY", "INDEPENDENT_RECIPE_MISMATCH", "CONSUME_RECIPE", "recipe")
	}
	if len(artifact.Evidence) != EvidenceTotal {
		return failure("LOWER_RESOLUTION", "PROOF_EVIDENCE_MISSING", "CONSUME_EVIDENCE", "evidence-count")
	}
	byKind := map[string]Evidence{}
	for _, item := range artifact.Evidence {
		if item.EvidenceDigest != evidenceDigest(item) || byKind[item.Kind].Kind != "" {
			return failure("INVARIANT_ONLY", "PROOF_EVIDENCE_DIGEST_MISMATCH", "CONSUME_EVIDENCE", "evidence-digest")
		}
		byKind[item.Kind] = item
	}
	sourceEvidence, sourceOK := byKind["SOURCE"]
	operationEvidence, operationOK := byKind["OPERATION"]
	invariantEvidence, invariantOK := byKind["INVARIANT"]
	if !sourceOK || !operationOK || !invariantOK {
		return failure("LOWER_RESOLUTION", "PROOF_EVIDENCE_MISSING", "CONSUME_EVIDENCE", "evidence-kind")
	}
	if validatePriorLedger(artifact.PriorLedger, artifact.Evidence) != nil {
		return failure("INVARIANT_ONLY", "PROOF_EVIDENCE_LEDGER_MISMATCH", "CONSUME_LEDGER", "prior-ledger")
	}
	sourceDigest := digestBytes(source)
	projection, err := projectSource(source, activityFrom(artifact))
	if err != nil {
		return failure("LOWER_RESOLUTION", "SOURCE_PROJECTION_UNKNOWN", "CONSUME_SOURCE", "projection")
	}
	if artifact.SourceDigest != sourceDigest || artifact.SemanticDigest != projection.SemanticDigest {
		return failure("INVARIANT_ONLY", "SOURCE_RECONSTRUCTION_MISMATCH", "CONSUME_SOURCE", "reconstruct")
	}
	receipt, err := decodeStrict[operationReceipt](operation)
	if err != nil || !verifyOperation(receipt, sourceDigest, artifact.SourcePath, projection) {
		return failure("INVARIANT_ONLY", "OPERATION_RECONSTRUCTION_MISMATCH", "CONSUME_OPERATION", "receipt")
	}
	if receipt.Digest == "" || !validDigest(receipt.Digest) {
		return failure("INVARIANT_ONLY", "OPERATION_EVIDENCE_LINK_MISMATCH", "CONSUME_OPERATION", "receipt-digest")
	}
	if sourceEvidence.ClaimID != "source-bytes-bound" || sourceEvidence.ProofChoice != "FOUNDATION" ||
		sourceEvidence.MetaOperation != "bind-source-bytes" || sourceEvidence.SourceDigest != sourceDigest || sourceEvidence.SemanticDigest != projection.SemanticDigest {
		return failure("INVARIANT_ONLY", "SOURCE_EVIDENCE_LINK_MISMATCH", "CONSUME_SOURCE", "source-evidence")
	}
	if operationEvidence.ClaimID != "operation-receipt-bound" || operationEvidence.ProofChoice != "COHERENCE" ||
		operationEvidence.MetaOperation != "bind-operation-receipt" || operationEvidence.SourceDigest != sourceDigest ||
		operationEvidence.SemanticDigest != projection.SemanticDigest || operationEvidence.ReceiptDigest != receipt.Digest || operationEvidence.Activity != receipt.Entry.Activity {
		return failure("INVARIANT_ONLY", "OPERATION_EVIDENCE_LINK_MISMATCH", "CONSUME_OPERATION", "operation-evidence")
	}
	if invariantEvidence.ClaimID != "no-byte-authority" || invariantEvidence.ProofChoice != "REGRESSION" ||
		invariantEvidence.MetaOperation != "preserve-no-byte-authority" || invariantEvidence.SourceDigest != sourceDigest ||
		invariantEvidence.SemanticDigest != projection.SemanticDigest ||
		invariantEvidence.Predicate != "generated-bytes-do-not-grant-authority" || invariantEvidence.RepositoryWrites != 0 ||
		invariantEvidence.MutationAuthority || invariantEvidence.ArtifactUseAuthority != "NONE" || !invariantEvidence.IndependentVerificationRequired {
		return failure("INVARIANT_ONLY", "INVARIANT_EVIDENCE_NOT_PRESERVED", "CONSUME_INVARIANT", "invariant-evidence")
	}
	evidence := []Evidence{sourceEvidence, operationEvidence, invariantEvidence}
	claims := exactClaims(evidence)
	return observation{Decision: "PASS", Resolution: "EXACT", Reason: "PROOF_CARRYING_ARTIFACT_AUTHORIZED",
		Coordinate: Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", "CONSUMER_ONLY_READ_ONLY_AUTHORITY"}, ArtifactDigest: artifactEvidenceDigest,
		SourceDigest: sourceDigest, SemanticDigest: projection.SemanticDigest, OperationDigest: receipt.Digest,
		EvidenceLinkDigest:    digestValue([]string{sourceEvidence.EvidenceDigest, operationEvidence.EvidenceDigest, invariantEvidence.EvidenceDigest}),
		ClaimTransitionDigest: digestValue(claimTransitions(claims)), Claims: claims}
}

func activityFrom(artifact Artifact) string {
	for _, item := range artifact.Evidence {
		if item.Kind == "OPERATION" {
			return item.Activity
		}
	}
	return ""
}
