package nonmonotonicrefutation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(encoded)
}
