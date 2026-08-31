package selfimprovementcandidate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validEvidenceDigest(value string) bool {
	if validDigest(value) {
		return true
	}
	_, err := hex.DecodeString(value)
	return err == nil && len(value) == 64 && value == strings.ToLower(value)
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
