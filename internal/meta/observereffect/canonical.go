package observereffect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func DigestValue(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(payload)
}

func ReceiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return DigestValue(receipt)
}

func ReportDigest(report Report) string {
	report.Digest = ""
	return DigestValue(report)
}
