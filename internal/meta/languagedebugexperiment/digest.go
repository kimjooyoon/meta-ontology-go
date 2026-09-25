package languagedebugexperiment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func sealReport(report Report) Report {
	report.Digest = ""
	data, _ := json.Marshal(report)
	sum := sha256.Sum256(data)
	report.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return report
}

func validDigest(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
