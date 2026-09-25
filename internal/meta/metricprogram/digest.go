package metricprogram

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func bytesDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func valueDigest(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal digest subject: %w", err)
	}
	return bytesDigest(payload), nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func validSubjectSHA(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 20 && value == strings.ToLower(value)
}

func normalizeSemanticDigest(value string) (string, error) {
	if validDigest(value) {
		return value, nil
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return "", fmt.Errorf("semantic digest is not sha256: %q", value)
	}
	return "sha256:" + value, nil
}
