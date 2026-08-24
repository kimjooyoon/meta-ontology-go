package assuranceeligibility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil { panic(err) }
	return digestBytes(encoded)
}

func sealReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func sealSuite(suite Suite) Suite {
	suite.SuiteDigest = ""
	suite.SuiteDigest = digestJSON(suite)
	return suite
}

func validSHA(value string) bool {
	if len(value) != 40 { return false }
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') { return false }
	}
	return true
}
