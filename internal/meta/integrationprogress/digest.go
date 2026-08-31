package integrationprogress

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}
