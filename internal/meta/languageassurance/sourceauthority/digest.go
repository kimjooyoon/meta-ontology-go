package sourceauthority

import (
	"crypto/sha256"
	"encoding/hex"
)

func Digest() string {
	return digestBytes(contractBytes)
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
