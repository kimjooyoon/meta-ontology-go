package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

func rejectInput(input Input) (string, string) {
	contract, profile := input.Contract, input.Profile
	if !ContractValid(contract) {
		return "EXPERIMENT_CONTRACT_INVALID", "EXACT"
	}
	if profile.Schema != ProfileSchema || profile.SubjectSHA != input.ExpectedHead ||
		!artifactemit.ValidSHA256(profile.ExecutableDigest) {
		return "EXPERIMENT_PROFILE_IDENTITY_MISMATCH", "EXACT"
	}
	if reason, resolution := rejectDecisions(input); reason != "" {
		return reason, resolution
	}
	if input.Artifact.Schema != contract.ArtifactSchema || input.Replay.Schema != contract.ArtifactSchema ||
		input.UnknownEmitter.Schema != contract.ArtifactSchema || input.Artifact.Kind != contract.EmitterKind ||
		input.Replay.Kind != contract.EmitterKind {
		return "ARTIFACT_IDENTITY_MISMATCH", "EXACT"
	}
	if !artifactemit.ValidDigest(input.Artifact) || !artifactemit.ValidDigest(input.Replay) ||
		!artifactemit.ValidDigest(input.UnknownEmitter) {
		return "ARTIFACT_DIGEST_INVALID", "EXACT"
	}
	if !profileEvidenceValid(profile, contract.Fixed) {
		return "PROFILE_SAMPLE_INVALID", "EXACT"
	}
	if !artifactEffectsValid(input) {
		return "EVIDENCE_EFFECTS_INVALID", "EXACT"
	}
	return "", ""
}
