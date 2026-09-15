package guardedcapability

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digest(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func seal(receipt *Receipt) {
	receipt.ReportDigest = ""
	receipt.ReportDigest = digest(*receipt)
}
