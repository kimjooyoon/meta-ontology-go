package languagetestexperiment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func reportDigest(report Report) string {
	report.Digest = ""
	payload, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func sealReport(report Report) Report {
	report.Digest = reportDigest(report)
	return report
}
