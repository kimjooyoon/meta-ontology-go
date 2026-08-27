package denominatorevolutionverify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func denominatorDigest(value Denominator) string {
	value.Digest = ""
	return digestValue(value)
}

func receiptDigest(value Receipt) string {
	value.Digest = ""
	return digestValue(value)
}

func recordDigest(value DenominatorRecord) string {
	value.Digest = ""
	return digestValue(value)
}

func claimLedgerDigest(value ClaimLedgerEntry) string {
	value.Digest = ""
	return digestValue(value)
}

func reportDigest(value Report) string {
	value.Digest = ""
	return digestValue(value)
}

func verificationDigest(value Verification) string {
	value.Digest = ""
	return digestValue(value)
}
