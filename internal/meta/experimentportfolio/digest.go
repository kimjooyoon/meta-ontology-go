package experimentportfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func sealReceipt(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = digestValue(receipt)
	return receipt
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func sealReport(report Report) Report {
	report.Digest = ""
	report.Digest = digestValue(report)
	return report
}
