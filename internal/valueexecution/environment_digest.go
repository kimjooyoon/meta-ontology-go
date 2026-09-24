package valueexecution

import "strings"

// EnvironmentInputs identifies the inputs that make an execution replayable.
// These are observations, not authorization grants.
type EnvironmentInputs struct {
	SourceDigest        string `json:"source_digest"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	ToolchainDigest     string `json:"toolchain_digest"`
	ContractDigest      string `json:"contract_digest"`
	ModelDigest         string `json:"model_digest"`
	SkillDigest         string `json:"skill_digest"`
	GatewayPolicyDigest string `json:"gateway_policy_digest"`
}

// EnvironmentStatus describes whether the environment boundary was measured.
type EnvironmentStatus string

const (
	EnvironmentStatusUnknown EnvironmentStatus = "UNKNOWN"
	EnvironmentStatusReady   EnvironmentStatus = "READY"
)

// EnvironmentObservation is a stable, non-authorizing description of the
// environment boundary used for one replay comparison.
type EnvironmentObservation struct {
	Schema         string            `json:"schema"`
	Inputs         EnvironmentInputs `json:"inputs"`
	Status         EnvironmentStatus `json:"status"`
	Missing        []string          `json:"missing,omitempty"`
	Reason         string            `json:"reason"`
	Digest         string            `json:"digest,omitempty"`
	NonAuthorizing bool              `json:"non_authorizing"`
}

const EnvironmentDigestSchema = "gooo/value-execution-environment/v1"

// EnvironmentTransition compares only complete environment observations.
type EnvironmentTransition string

const (
	EnvironmentTransitionUnknown   EnvironmentTransition = "UNKNOWN"
	EnvironmentTransitionUnchanged EnvironmentTransition = "UNCHANGED"
	EnvironmentTransitionChanged   EnvironmentTransition = "CHANGED"
)

// ObserveEnvironment normalizes and hashes the declared environment inputs.
// Missing or malformed identities remain UNKNOWN and never become replay proof.
func ObserveEnvironment(inputs EnvironmentInputs) EnvironmentObservation {
	inputs = normalizeEnvironmentInputs(inputs)
	observation := EnvironmentObservation{
		Schema:         EnvironmentDigestSchema,
		Inputs:         inputs,
		Status:         EnvironmentStatusUnknown,
		Reason:         "ENVIRONMENT_INPUT_INCOMPLETE",
		NonAuthorizing: true,
		Missing:        missingEnvironmentInputs(inputs)}
	if len(observation.Missing) != 0 {
		return observation
	}
	observation.Status = EnvironmentStatusReady
	observation.Reason = "ENVIRONMENT_BOUND"
	observation.Digest = digestValue(struct {
		Schema string            `json:"schema"`
		Inputs EnvironmentInputs `json:"inputs"`
	}{Schema: EnvironmentDigestSchema, Inputs: inputs})
	return observation
}

// CompareEnvironments refuses to treat an incomplete boundary as unchanged.
func CompareEnvironments(before, after EnvironmentObservation) EnvironmentTransition {
	if before.Status != EnvironmentStatusReady || after.Status != EnvironmentStatusReady ||
		!validDigest(before.Digest) || !validDigest(after.Digest) {
		return EnvironmentTransitionUnknown
	}
	if before.Digest == after.Digest {
		return EnvironmentTransitionUnchanged
	}
	return EnvironmentTransitionChanged
}

func normalizeEnvironmentInputs(inputs EnvironmentInputs) EnvironmentInputs {
	inputs.SourceDigest = strings.TrimSpace(inputs.SourceDigest)
	inputs.SemanticFingerprint = strings.TrimSpace(inputs.SemanticFingerprint)
	inputs.ToolchainDigest = strings.TrimSpace(inputs.ToolchainDigest)
	inputs.ContractDigest = strings.TrimSpace(inputs.ContractDigest)
	inputs.ModelDigest = strings.TrimSpace(inputs.ModelDigest)
	inputs.SkillDigest = strings.TrimSpace(inputs.SkillDigest)
	inputs.GatewayPolicyDigest = strings.TrimSpace(inputs.GatewayPolicyDigest)
	return inputs
}

func missingEnvironmentInputs(inputs EnvironmentInputs) []string {
	missing := make([]string, 0, 7)
	values := []struct {
		name  string
		value string
	}{
		{name: "source_digest", value: inputs.SourceDigest},
		{name: "semantic_fingerprint", value: inputs.SemanticFingerprint},
		{name: "toolchain_digest", value: inputs.ToolchainDigest},
		{name: "contract_digest", value: inputs.ContractDigest},
		{name: "model_digest", value: inputs.ModelDigest},
		{name: "skill_digest", value: inputs.SkillDigest},
		{name: "gateway_policy_digest", value: inputs.GatewayPolicyDigest},
	}
	for _, item := range values {
		if !validDigest(item.value) {
			missing = append(missing, item.name)
		}
	}
	return missing
}
