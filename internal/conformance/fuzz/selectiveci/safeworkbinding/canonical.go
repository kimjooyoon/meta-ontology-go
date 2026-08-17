package safeworkbinding

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	tagString byte = 1 + iota
	tagStableID
	tagDigest
	tagLegacyWorkID
	tagEnum
	tagBool
	tagReasonList
	tagU64
)

type frameField struct {
	name  string
	tag   byte
	value []byte
}

func appendU64(out []byte, value uint64) []byte {
	return append(out, byte(value>>56), byte(value>>48), byte(value>>40), byte(value>>32), byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
}

func appendLength(out, value []byte) []byte {
	out = appendU64(out, uint64(len(value)))
	return append(out, value...)
}

func frame(domain string, fields ...frameField) []byte {
	out := appendLength(nil, []byte(domain))
	out = appendU64(out, uint64(len(fields)))
	for _, field := range fields {
		out = appendLength(out, []byte(field.name))
		out = append(out, field.tag)
		out = appendLength(out, field.value)
	}
	return out
}

func decisionSpelling(value Decision) string {
	switch value {
	case DecisionPass:
		return "PASS"
	case DecisionUnknown:
		return "UNKNOWN"
	case DecisionFailClosed:
		return "FAIL_CLOSED"
	default:
		return ""
	}
}

func reasonSpelling(value Reason) string {
	spellings := [...]string{"NONE", "REQUIRED_INPUT_MISSING", "INVALID_UTF8", "BOM_FORBIDDEN", "INVALID_JSON", "TRAILING_VALUE", "DUPLICATE_KEY", "UNKNOWN_FIELD", "NULL_VALUE", "EMPTY_VALUE", "INVALID_SCHEMA", "INVALID_STABLE_ID", "INVALID_DIGEST", "BINDING_DIGEST_MISMATCH"}
	if int(value) >= len(spellings) {
		return ""
	}
	return spellings[value]
}

func effectSpelling(value EnforcementEffect) string {
	if value == EnforcementNoEffect {
		return "NO_EFFECT"
	}
	return ""
}

func reasonListBytes(faults []Reason) []byte {
	out := appendU64(nil, uint64(len(faults)))
	for _, fault := range faults {
		out = appendLength(out, []byte(reasonSpelling(fault)))
	}
	return out
}

func bindingFrame(binding SafeWorkBinding) []byte {
	return frame("gooo/safe-work-binding/input/v1\x00",
		frameField{"schema", tagString, []byte(binding.Schema)},
		frameField{"task_id", tagStableID, []byte(binding.TaskID)},
		frameField{"path_id", tagStableID, []byte(binding.PathID)},
		frameField{"obligation_id", tagStableID, []byte(binding.ObligationID)},
		frameField{"source_snapshot_digest", tagDigest, []byte(binding.SourceSnapshotDigest)},
		frameField{"semantic_snapshot_digest", tagDigest, []byte(binding.SemanticSnapshotDigest)},
		frameField{"policy_digest", tagDigest, []byte(binding.PolicyDigest)},
		frameField{"registry_digest", tagDigest, []byte(binding.RegistryDigest)},
		frameField{"toolchain_options_digest", tagDigest, []byte(binding.ToolchainOptionsDigest)})
}

func resultFrame(result ParseResult) []byte {
	return frame("gooo/safe-work-binding/result/v1\x00",
		frameField{"decision", tagEnum, []byte(decisionSpelling(result.Decision))},
		frameField{"reason", tagEnum, []byte(reasonSpelling(result.Reason))},
		frameField{"faults", tagReasonList, reasonListBytes(result.Faults)},
		frameField{"full_suite_required", tagBool, []byte{boolByte(result.FullSuiteRequired)}},
		frameField{"execution_authorized", tagBool, []byte{boolByte(result.ExecutionAuthorized)}},
		frameField{"enforcement_effect", tagEnum, []byte(effectSpelling(result.EnforcementEffect))})
}

func replayFrame(bindingDigest Digest, result ParseResult) []byte {
	return frame("gooo/safe-work-binding/replay/v1\x00",
		frameField{"binding_digest", tagDigest, []byte(bindingDigest)},
		frameField{"result_digest", tagDigest, []byte(result.ResultDigest)})
}

func boolByte(value bool) byte {
	if value {
		return 1
	}
	return 0
}

func digestFrame(value []byte) Digest {
	digest := sha256.Sum256(value)
	return Digest("sha256:" + hex.EncodeToString(digest[:]))
}

func bindingDigest(binding SafeWorkBinding) Digest { return digestFrame(bindingFrame(binding)) }
func resultDigest(result ParseResult) Digest       { return digestFrame(resultFrame(result)) }
func replayDigest(binding Digest, result ParseResult) Digest {
	return digestFrame(replayFrame(binding, result))
}
