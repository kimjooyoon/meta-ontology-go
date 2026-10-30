package nativeobservation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// SourceCheckBindingDigest calculates the existing owned-toolchain five-tuple
// binding. SourceOutput uses Go []byte encoding: nil is null, a non-nil slice is
// a base64 string. Field order and omitted process fields follow these types.
// Schema and reuse flags are outside this tuple. The digest is a content
// identity, not an execution nonce.
func SourceCheckBindingDigest(reference ToolchainObservation) string {
	return bindingDigest([]any{reference.KeySHA256, reference.GoToolSHA256,
		reference.GoVersion, reference.SourceCheck, reference.SourceOutput})
}

// SourceBuildBindingDigest binds the artifact key, executable bytes' digest and
// original build observation. Current reuse/wait observations are not inputs.
// It calculates a recorded binding; it does not attest execution or validate
// that the process, keys or executable are genuine or successful.
func SourceBuildBindingDigest(artifactKey, executableSHA256 string, sourceBuild ProcessObservation) string {
	return bindingDigest([]any{artifactKey, executableSHA256, sourceBuild})
}

func bindingDigest(tuple []any) string {
	// These tuples contain only the concrete scalar/byte/process types above:
	// no floats, maps, cycles or custom marshalers can fail JSON encoding.
	raw, _ := json.Marshal(tuple)
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}
