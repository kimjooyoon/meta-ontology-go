package metacircularboundaryconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func sealReceipt(receipt contract.Receipt) contract.Receipt {
	receipt.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(receipt)
	return receipt
}

func sealReport(report contract.Report) contract.Report {
	report.ReportDigest = ""
	report.ReportDigest = digestValue(report)
	return report
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
