package guardedpromotion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func seal(report *Report) {
	report.ReportDigest = ""
	data, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	report.ReportDigest = "sha256:" + hex.EncodeToString(sum[:])
}
