package languagesemanticbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func loadJSON[T any](path string) (T, []byte, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, nil, err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, nil, err
	}
	return value, data, nil
}

func fileDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func validEvidenceDigest(value string) bool {
	if validDigest(value) {
		return true
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func require(condition bool, message string) error {
	if !condition {
		return fmt.Errorf("semantic readiness binding: %s", message)
	}
	return nil
}
