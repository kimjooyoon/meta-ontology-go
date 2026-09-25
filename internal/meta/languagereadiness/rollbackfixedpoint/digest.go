package rollbackfixedpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestJSON(value any) string {
	payload, _ := json.Marshal(value)
	return digestBytes(payload)
}

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	raw := strings.TrimPrefix(value, "sha256:")
	if len(raw) != 64 || raw != strings.ToLower(raw) {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func seal(value Report) Report {
	value.ReportDigest = ""
	value.ReportDigest = digestJSON(value)
	return value
}

func Encode(value Report) []byte {
	payload, _ := json.MarshalIndent(value, "", "  ")
	return append(payload, '\n')
}
