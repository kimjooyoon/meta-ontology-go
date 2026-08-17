package safeworkbinding

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

type frameTag byte

const (
	frameTagString       frameTag = 0x01
	frameTagStableID     frameTag = 0x02
	frameTagDigest       frameTag = 0x03
	frameTagLegacyWorkID frameTag = 0x04
	frameTagEnum         frameTag = 0x05
	frameTagBool         frameTag = 0x06
	frameTagReasonList   frameTag = 0x07
	frameTagU64          frameTag = 0x08
)

type frameField struct {
	name  string
	tag   frameTag
	value []byte
}

func appendU64BE(out []byte, value uint64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	return append(out, encoded[:]...)
}
func encodeFrame(domain string, fields []frameField) []byte {
	out := appendU64BE(nil, uint64(len(domain)))
	out = append(out, domain...)
	out = appendU64BE(out, uint64(len(fields)))
	for _, field := range fields {
		out = appendU64BE(out, uint64(len(field.name)))
		out = append(out, field.name...)
		out = append(out, byte(field.tag))
		out = appendU64BE(out, uint64(len(field.value)))
		out = append(out, field.value...)
	}
	return out
}
func encodeStringField(name, value string) frameField {
	return frameField{name: name, tag: frameTagString, value: []byte(value)}
}
func encodeStableIDField(name string, value StableID) frameField {
	return frameField{name: name, tag: frameTagStableID, value: []byte(value)}
}
func encodeDigestField(name string, value Digest) frameField {
	return frameField{name: name, tag: frameTagDigest, value: []byte(value)}
}
func encodeLegacyWorkIDField(name string, value LegacyWorkID) frameField {
	return frameField{name: name, tag: frameTagLegacyWorkID, value: []byte(value)}
}
func encodeEnumField(name string, spelling []byte) frameField {
	value := append([]byte(nil), spelling...)
	return frameField{name: name, tag: frameTagEnum, value: value}
}
func encodeBoolField(name string, value bool) frameField {
	encoded := byte(0)
	if value {
		encoded = 1
	}
	return frameField{name: name, tag: frameTagBool, value: []byte{encoded}}
}
func encodeListField(name string, values [][]byte) frameField {
	encoded := appendU64BE(nil, uint64(len(values)))
	for _, value := range values {
		encoded = appendU64BE(encoded, uint64(len(value)))
		encoded = append(encoded, value...)
	}
	return frameField{name: name, tag: frameTagReasonList, value: encoded}
}

func decisionSpelling(value Decision) ([]byte, bool) {
	switch value {
	case DecisionPass:
		return []byte("PASS"), true
	case DecisionUnknown:
		return []byte("UNKNOWN"), true
	case DecisionFailClosed:
		return []byte("FAIL_CLOSED"), true
	default:
		return nil, false
	}
}

func reasonSpelling(value Reason) ([]byte, bool) {
	spellings := [...]string{
		"NONE", "REQUIRED_INPUT_MISSING", "INVALID_UTF8", "BOM_FORBIDDEN",
		"INVALID_JSON", "TRAILING_VALUE", "DUPLICATE_KEY", "UNKNOWN_FIELD",
		"NULL_VALUE", "EMPTY_VALUE", "INVALID_SCHEMA", "INVALID_STABLE_ID",
		"INVALID_DIGEST", "BINDING_DIGEST_MISMATCH",
	}
	if int(value) >= len(spellings) {
		return nil, false
	}
	return []byte(spellings[value]), true
}

func enforcementEffectSpelling(value EnforcementEffect) ([]byte, bool) {
	if value != EnforcementEffectNoEffect {
		return nil, false
	}
	return []byte("NO_EFFECT"), true
}

func decisionField(name string, value Decision) (frameField, bool) {
	spelling, ok := decisionSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func reasonField(name string, value Reason) (frameField, bool) {
	spelling, ok := reasonSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func enforcementEffectField(name string, value EnforcementEffect) (frameField, bool) {
	spelling, ok := enforcementEffectSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func reasonListField(name string, values []Reason) (frameField, bool) {
	spellings := make([][]byte, len(values))
	for i, value := range values {
		spelling, ok := reasonSpelling(value)
		if !ok {
			return frameField{}, false
		}
		spellings[i] = spelling
	}
	return encodeListField(name, spellings), true
}

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
