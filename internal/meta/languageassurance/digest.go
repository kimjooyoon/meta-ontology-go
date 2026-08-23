package languageassurance

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

func seal(report *Report) {
	copy := *report
	copy.ReportDigest = ""
	report.ReportDigest = digest(copy)
}
