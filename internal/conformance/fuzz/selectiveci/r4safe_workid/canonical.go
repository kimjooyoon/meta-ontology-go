package r4safe_workid

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

func canonicalResultBytes(result Result) []byte {
	buffer := bytes.NewBufferString("gooo/legacy-work-id-audit/result/v1")
	buffer.WriteByte(0)
	encoded := buffer.Bytes()
	encoded = appendLengthPrefixed(encoded, decisionSpelling(result.Decision))
	encoded = appendLengthPrefixed(encoded, reasonSpelling(result.Reason))
	encoded = appendLengthPrefixed(encoded, []byte(result.LegacyWorkID))
	encoded = appendBool(encoded, result.FullSuiteRequired)
	encoded = appendBool(encoded, result.ExecutionAuthorized)
	return appendLengthPrefixed(encoded, effectSpelling(result.EnforcementEffect))
}
func canonicalDigest(result Result) AuditDigest {
	digest := sha256.Sum256(canonicalResultBytes(result))
	return AuditDigest(hex.EncodeToString(digest[:]))
}

func decisionSpelling(decision Decision) []byte {
	switch decision {
	case DecisionUnknown:
		return []byte("UNKNOWN")
	case DecisionPass:
		return []byte("PASS")
	case DecisionFailClosed:
		return []byte("FAIL_CLOSED")
	default:
		return nil
	}
}

func reasonSpelling(reason Reason) []byte {
	switch reason {
	case ReasonRequiredInputMissing:
		return []byte("REQUIRED_INPUT_MISSING")
	case ReasonDerived:
		return []byte("DERIVED")
	case ReasonMatchingCallerOverride:
		return []byte("MATCHING_CALLER_OVERRIDE")
	case ReasonMalformedOverride:
		return []byte("MALFORMED_OVERRIDE")
	case ReasonWorkIDMismatch:
		return []byte("WORK_ID_MISMATCH")
	default:
		return nil
	}
}

func effectSpelling(effect EnforcementEffect) []byte {
	if effect == EnforcementNoEffect {
		return []byte("NO_EFFECT")
	}
	return nil
}

func appendLengthPrefixed(buffer, value []byte) []byte {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	buffer = append(buffer, length[:]...)
	return append(buffer, value...)
}

func appendBool(buffer []byte, value bool) []byte {
	if value {
		return append(buffer, 1)
	}
	return append(buffer, 0)
}
