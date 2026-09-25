package languagesemanticbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestParts(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func finalizeReport(report *Report) {
	report.ReportDigest = ""
	payload, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(payload)
	report.ReportDigest = "sha256:" + hex.EncodeToString(sum[:])
}
