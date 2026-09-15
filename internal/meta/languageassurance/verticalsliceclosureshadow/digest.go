package verticalsliceclosureshadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	payload, _ := json.Marshal(value)
	return digestBytes(payload)
}

func validDigest(value string) bool {
	raw := strings.TrimPrefix(value, "sha256:")
	_, err := hex.DecodeString(raw)
	return len(raw) == 64 && raw == strings.ToLower(raw) && err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func Encode(report Report) []byte {
	payload, _ := json.MarshalIndent(report, "", "  ")
	return append(payload, '\n')
}
