package promotioncontinuity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func fileSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func reportDigest(report Report) (string, error) {
	report.ReportDigest = ""
	data, err := json.Marshal(report)
	if err != nil {
		return "", err
	}
	return fileSHA256(data), nil
}

func seal(report Report) Report {
	digest, err := reportDigest(report)
	if err == nil {
		report.ReportDigest = digest
	}
	return report
}
