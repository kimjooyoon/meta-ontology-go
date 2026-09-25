package metainvocation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digest(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func bytesDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sealPlan(plan CheckPlan) CheckPlan {
	plan.Digest = ""
	plan.Digest = digest(plan)
	return plan
}

func sealReceipt(receipt VerificationReceipt) VerificationReceipt {
	receipt.Digest = ""
	receipt.Digest = digest(receipt)
	return receipt
}

func sealReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digest(report)
	return report
}
