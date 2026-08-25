package languageexampleexperiment

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func rejectInput(input Input) (string, string) {
	contract, profile := input.Contract, input.Profile
	if contract.Schema != ContractSchema || contract.ID == "" || contract.ArtifactSchema != artifactemit.OperationManifestSchema ||
		contract.EmitterKind != artifactemit.OperationManifestKind || contract.Fixed.Indicators != 13 ||
		contract.Fixed.NotClaimed != len(contract.NotClaimed) {
		return "EXPERIMENT_CONTRACT_INVALID", "EXACT"
	}
	if profile.Schema != ProfileSchema || profile.SubjectSHA != input.ExpectedHead ||
		!strings.HasPrefix(profile.ExecutableDigest, "sha256:") {
		return "EXPERIMENT_PROFILE_IDENTITY_MISMATCH", "EXACT"
	}
	if input.Artifact.Decision != "PASS" || input.Replay.Decision != "PASS" {
		return "ARTIFACT_DECISION_UNKNOWN", "LOWER_RESOLUTION"
	}
	if input.Artifact.Schema != contract.ArtifactSchema || input.Replay.Schema != contract.ArtifactSchema ||
		input.UnknownEmitter.Schema != contract.ArtifactSchema || input.Artifact.Kind != contract.EmitterKind ||
		input.Replay.Kind != contract.EmitterKind || !strings.HasPrefix(input.Artifact.Digest, "sha256:") ||
		!strings.HasPrefix(input.Replay.Digest, "sha256:") {
		return "ARTIFACT_IDENTITY_MISMATCH", "EXACT"
	}
	return "", ""
}
