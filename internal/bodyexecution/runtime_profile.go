package bodyexecution

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// Consistency checks preserve failed observations. They do not attest a build,
// executable, source authority or model operation by reading caller-owned JSON.
func runtimeProfileBinding(result Result, receipt *completeness.CompletenessReceipt) error {
	_, present := receipt.Scope["owned_artifact"]
	_, toolPresent := receipt.Scope["owned_toolchain"]
	a := result.Observation.Artifact
	if receipt.ProfileID == RuntimeProfileV1 && a == nil && !present && !toolPresent && result.Observation.ToolchainReference == nil {
		return nil
	}
	if (receipt.ProfileID != RuntimeProfileV2 && receipt.ProfileID != RuntimeProfileV3) || a == nil || !present {
		return fmt.Errorf("unsupported or inconsistent runtime ownership profile")
	}
	scopeArtifact, err := ownedScopeArtifact(receipt)
	if err != nil || !reflect.DeepEqual(*a, scopeArtifact) {
		return fmt.Errorf("runtime owned artifact scope differs or is unsupported")
	}
	observation, err := json.Marshal(result.Observation)
	if err != nil || receipt.Scope["runtime_observation_sha256"] != digest(observation) {
		return fmt.Errorf("runtime observation digest differs")
	}
	if receipt.ProfileID == RuntimeProfileV2 {
		if toolPresent || result.Observation.ToolchainReference != nil {
			return fmt.Errorf("v2 cannot carry toolchain reuse")
		}
	} else {
		reference, err := ownedScopeToolchain(receipt)
		if err != nil || result.Observation.ToolchainReference == nil || !reflect.DeepEqual(reference, *result.Observation.ToolchainReference) ||
			reference.GoToolSHA256 != result.Observation.GoToolSHA256 || reference.GoVersion != result.Observation.GoVersion {
			return fmt.Errorf("owned toolchain scope differs or is incomplete")
		}
		if reference.Reused {
			if !reflect.DeepEqual(result.Observation.Toolchain, ProcessObservation{}) {
				return fmt.Errorf("reused toolchain invents current process work")
			}
		} else if !reflect.DeepEqual(result.Observation.Toolchain, reference.SourceCheck) {
			return fmt.Errorf("fresh toolchain process differs from retained check")
		}
	}
	return sourceBuildBinding(result.Observation)
}

// ValidateOwnedRuntimeScope checks registered v2/v3 ownership records in a bare
// receipt. It validates recorded bindings, without attesting runtime execution.
func ValidateOwnedRuntimeScope(receipt *completeness.CompletenessReceipt) error {
	_, err := ownedScopeArtifact(receipt)
	if err == nil && receipt.ProfileID == RuntimeProfileV3 {
		_, err = ownedScopeToolchain(receipt)
	}
	if err == nil && receipt.ProfileID == RuntimeProfileV2 {
		if _, present := receipt.Scope["owned_toolchain"]; present {
			return fmt.Errorf("v2 cannot carry owned toolchain")
		}
	}
	return err
}

func ownedScopeArtifact(receipt *completeness.CompletenessReceipt) (ArtifactObservation, error) {
	var artifact ArtifactObservation
	if receipt == nil || (receipt.ProfileID != RuntimeProfileV2 && receipt.ProfileID != RuntimeProfileV3) {
		return artifact, fmt.Errorf("registered owned runtime receipt is required")
	}
	raw, err := json.Marshal(receipt.Scope["owned_artifact"])
	observationDigest, _ := receipt.Scope["runtime_observation_sha256"].(string)
	if err != nil || decode(raw, &artifact, 1<<20) != nil ||
		artifact.Schema != "gooo/owned-native-artifact/v1" || !runtimeDigest(artifact.KeySHA256) ||
		artifact.WaitNS < 0 || !runtimeDigest(observationDigest) {
		return artifact, fmt.Errorf("incomplete or unsupported owned runtime scope")
	}
	scope, ok := receipt.Scope["runtime_scope"].(map[string]any)
	if !ok {
		return artifact, fmt.Errorf("owned runtime execution scope is missing")
	}
	executable, ok := scope["executable_sha256"].(string)
	if !ok {
		return artifact, fmt.Errorf("owned runtime executable binding is missing")
	}
	return artifact, sourceBuildBinding(Observation{Artifact: &artifact, ExecutableSHA256: executable})
}

func sourceBuildBinding(o Observation) error {
	a := o.Artifact
	if a.SourceBuildSHA256 == "" {
		if a.Reused || a.ExecutableVerified || !reflect.DeepEqual(a.SourceBuild, ProcessObservation{}) {
			return fmt.Errorf("runtime owned build binding is missing")
		}
		return nil // The build stage may have failed before ownership was recorded.
	}
	bound, err := json.Marshal([]any{a.KeySHA256, o.ExecutableSHA256, a.SourceBuild})
	p := a.SourceBuild
	if err != nil || !a.ExecutableVerified || !runtimeDigest(o.ExecutableSHA256) ||
		!p.Started || !p.Completed || p.ExitCode == nil || *p.ExitCode != 0 || p.Canceled ||
		p.TimedOut || p.OutputTruncated || p.DiagnosticsBytes != 0 || a.SourceBuildSHA256 != digest(bound) {
		return fmt.Errorf("runtime source build binding differs")
	}
	return nil
}

func runtimeDigest(s string) bool {
	if len(s) != 71 || !strings.HasPrefix(s, "sha256:") {
		return false
	}
	for _, c := range s[7:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
