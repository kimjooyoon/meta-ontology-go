package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func sealedReportDigest(report capability.Report) string {
	report.ReportDigest = ""
	return digestValue(report)
}
