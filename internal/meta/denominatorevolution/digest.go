package denominatorevolution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return DigestBytes(raw)
}

func denominatorDigest(value Denominator) string {
	value.Digest = ""
	return digestValue(value)
}

func receiptDigest(value MigrationReceipt) string {
	value.Digest = ""
	return digestValue(value)
}

func reportDigest(value Report) string {
	value.Digest = ""
	return digestValue(value)
}
