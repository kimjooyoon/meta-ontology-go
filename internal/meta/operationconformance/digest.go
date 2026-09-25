package operationconformance

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
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestValue(report)
	return report
}
