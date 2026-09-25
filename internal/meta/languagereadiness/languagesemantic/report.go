package languagesemantic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func finalizeReport(report *Report) {
	report.ReportDigest = ""
	raw, _ := json.Marshal(report)
	report.ReportDigest = digestBytes(raw)
}

func marshalReport(report Report) ([]byte, error) {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func caseDigest(result CaseResult) string {
	result.Digest = ""
	raw, _ := json.Marshal(result)
	return digestBytes(raw)
}
