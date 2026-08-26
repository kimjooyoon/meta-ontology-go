package toolchainlsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func finish(report Report) Report {
	for index := range report.Cases {
		report.Cases[index].EvidenceDigest = digestValue(struct {
			ID, Expected, Observed, Status string
		}{report.Cases[index].ID, report.Cases[index].Expected,
			report.Cases[index].Observed, report.Cases[index].Status})
	}
	report.ReportDigest = ""
	report.ReportDigest = digestValue(report)
	return report
}
