package safeworkbinding

import (
	"crypto/sha256"
	"encoding/hex"
)

func bindingFrame(binding SafeWorkBinding) []byte {
	return encodeFrame("gooo/safe-work-binding/input/v1\x00", []frameField{
		encodeStringField("schema", binding.Schema),
		encodeStableIDField("task_id", binding.TaskID),
		encodeStableIDField("path_id", binding.PathID),
		encodeStableIDField("obligation_id", binding.ObligationID),
		encodeDigestField("source_snapshot_digest", binding.SourceSnapshotDigest),
		encodeDigestField("semantic_snapshot_digest", binding.SemanticSnapshotDigest),
		encodeDigestField("policy_digest", binding.PolicyDigest),
		encodeDigestField("registry_digest", binding.RegistryDigest),
		encodeDigestField("toolchain_options_digest", binding.ToolchainOptionsDigest),
	})
}
func bindingDigest(binding SafeWorkBinding) Digest {
	digest := sha256.Sum256(bindingFrame(binding))
	return Digest("sha256:" + hex.EncodeToString(digest[:]))
}
func resultFrame(result ParseResult) ([]byte, bool) {
	decision, ok := decisionField("decision", result.Decision)
	if !ok {
		return nil, false
	}
	reason, ok := reasonField("reason", result.Reason)
	if !ok {
		return nil, false
	}
	faults, ok := reasonListField("faults", result.Faults)
	if !ok {
		return nil, false
	}
	effect, ok := enforcementEffectField("enforcement_effect", result.EnforcementEffect)
	if !ok {
		return nil, false
	}
	return encodeFrame("gooo/safe-work-binding/result/v1\x00", []frameField{
		decision,
		reason,
		faults,
		encodeBoolField("full_suite_required", result.FullSuiteRequired),
		encodeBoolField("execution_authorized", result.ExecutionAuthorized),
		effect,
	}), true
}
func resultDigest(result ParseResult) (Digest, bool) {
	frame, ok := resultFrame(result)
	if !ok {
		return "", false
	}
	digest := sha256.Sum256(frame)
	return Digest("sha256:" + hex.EncodeToString(digest[:])), true
}
func replayFrame(binding, result Digest) []byte {
	return encodeFrame("gooo/safe-work-binding/replay/v1\x00", []frameField{
		encodeDigestField("binding_digest", binding),
		encodeDigestField("result_digest", result),
	})
}
func replayDigest(binding, result Digest) Digest {
	digest := sha256.Sum256(replayFrame(binding, result))
	return Digest("sha256:" + hex.EncodeToString(digest[:]))
}
