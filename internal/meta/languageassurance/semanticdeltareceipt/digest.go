package semanticdeltareceipt

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
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sealReceipt(receipt *Receipt) {
	copy := *receipt
	copy.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(copy)
}

func sealReport(report *Report) {
	copy := *report
	copy.ReportDigest = ""
	report.ReportDigest = digestValue(copy)
}

func sealSuite(suite *Suite) {
	copy := *suite
	copy.SuiteDigest = ""
	suite.SuiteDigest = digestValue(copy)
}
