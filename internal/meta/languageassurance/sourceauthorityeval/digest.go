package sourceauthorityeval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func DigestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func seal(report Report) Report {
	report.ReceiptDigest = ""
	raw, err := json.Marshal(report)
	if err != nil {
		report.Observation = "ERROR"
		report.Resolution = "INVARIANT_ONLY"
		report.Enforcement = "BLOCK"
		report.Reason = "REPORT_ENCODING_ERROR"
		return report
	}
	report.ReceiptDigest = DigestBytes(raw)
	return report
}
