package symbolicinvocationusecase

func resolutionFailure(input Input) (facts, string, string, bool) {
	receipt, artifact, observation := input.ProducerReceipt, input.ProducerArtifact, input.Observation
	if !validSHA(input.SubjectSHA) {
		return facts{Unknowns: 1}, reasonDecisionUnknown, "LOWER_RESOLUTION", true
	}
	if unknownTop(receipt.Decision, receipt.Resolution) || unknownTop(artifact.Decision, artifact.Resolution) ||
		unknownTop(observation.Decision, observation.Resolution) {
		return facts{Unknowns: 1}, reasonDecisionUnknown, "LOWER_RESOLUTION", true
	}
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || artifact.Decision != "PASS" ||
		artifact.Resolution != "SYMBOLIC_ONLY" || observation.Decision != "PASS" || observation.Resolution != "EXACT" {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY", true
	}
	return facts{}, "", "", false
}

func identityFailure(input Input) string {
	receipt, artifact, observation := input.ProducerReceipt, input.ProducerArtifact, input.Observation
	if receipt.SubjectSHA != input.SubjectSHA || observation.SubjectSHA != input.SubjectSHA {
		return reasonSubjectMismatch
	}
	if receipt.Schema != "gooo/symbolic-invocation-schema-receipt/v1" ||
		receipt.Reason != "EXTERNAL_SCHEMA_VALIDATION_OBSERVED" ||
		artifact.Schema != "gooo/symbolic-invocation-schema-artifact/v1" ||
		artifact.Reason != "SYMBOLIC_INVOCATION_SCHEMA_EMITTED" || artifact.Kind != "symbolic-invocation-schema" ||
		observation.Schema != "gooo/symbolic-invocation-usecase-observation/v1" ||
		observation.Reason != "EXTERNAL_USER_VALIDATION_REPLAYED" {
		return reasonEvidenceInvalid
	}
	if !validDigest(receipt.Compiler.BinaryDigest) || !validDigest(receipt.Artifact.Digest) ||
		!validDigest(receipt.Artifact.JSONSchemaDigest) || !validDigest(receipt.Validation.ToolDigest) ||
		!validDigest(artifact.Digest) || !validDigest(observation.ArtifactDigest) ||
		!validDigest(observation.JSONSchemaDigest) || !validDigest(observation.ToolDigest) {
		return reasonEvidenceInvalid
	}
	if receipt.Artifact.ArtifactSchema != artifact.Schema || receipt.Artifact.Digest != artifact.Digest ||
		receipt.Artifact.JSONSchemaDigest != observation.JSONSchemaDigest ||
		receipt.Validation.ToolDigest != observation.ToolDigest || artifact.Digest != observation.ArtifactDigest {
		return reasonLinkMismatch
	}
	return ""
}

func unknownTop(decision, resolution string) bool {
	knownDecision := decision == "PASS" || decision == "FAIL_CLOSED"
	knownResolution := resolution == "EXACT" || resolution == "SYMBOLIC_ONLY" ||
		resolution == "LOWER_RESOLUTION" || resolution == "INVARIANT_ONLY"
	return !knownDecision || !knownResolution
}
