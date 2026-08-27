package capabilityscopedexpansion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	var normalized any
	if json.Unmarshal(raw, &normalized) == nil {
		raw, _ = json.Marshal(normalized)
	}
	return digestBytes(raw)
}

func sealReceipt(receipt Receipt) Receipt {
	receipt.ReportDigest = ""
	receipt.ReportDigest = digestValue(receipt)
	return receipt
}

func SealSuite(suite Suite) Suite {
	suite.SuiteDigest = ""
	suite.SuiteDigest = digestValue(suite)
	return suite
}
