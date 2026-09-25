package foundationseed

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func seal(report Report) Report {
	report.Digest = ""
	raw, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(raw)
	report.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return report
}
