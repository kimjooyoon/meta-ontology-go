package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func claimResolutionDigest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
