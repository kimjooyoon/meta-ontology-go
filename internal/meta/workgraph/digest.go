package workgraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func DigestValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return DigestBytes(encoded)
}
