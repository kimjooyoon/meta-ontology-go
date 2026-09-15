package toolchaincli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func caseDigest(result CaseResult) string {
	result.First.PeakRSSKiB = 0
	result.Replay.PeakRSSKiB = 0
	result.EvidenceDigest = ""
	return digestJSON(result)
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}
