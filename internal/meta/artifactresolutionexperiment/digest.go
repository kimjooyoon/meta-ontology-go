package artifactresolutionexperiment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func seal(report Report) Report {
	report.Digest = ""
	report.Digest = digestValue(report)
	return report
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
