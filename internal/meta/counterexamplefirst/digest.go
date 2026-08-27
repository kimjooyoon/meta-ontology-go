package counterexamplefirst

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return DigestBytes(raw)
}

func CounterexampleDigest(value Counterexample) string {
	return DigestValue(value)
}

func ResolutionDigest(value ResolutionEvidence) string {
	return DigestValue(value)
}

func ReceiptDigest(value DecisionReceipt) string {
	value.Digest = ""
	return DigestValue(value)
}

func ReportDigest(value Report) string {
	value.Digest = ""
	return DigestValue(value)
}
