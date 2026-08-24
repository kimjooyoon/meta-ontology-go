package candidateleakageeligibility

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
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
