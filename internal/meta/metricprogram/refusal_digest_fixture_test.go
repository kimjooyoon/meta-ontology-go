package metricprogram_test

import (
	"crypto/sha256"
	"encoding/hex"
)

func fixtureBytesDigest1737(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
