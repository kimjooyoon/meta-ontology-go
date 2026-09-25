package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
