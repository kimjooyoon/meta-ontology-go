package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}
