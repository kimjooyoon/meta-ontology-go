package artifactfeedback

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
)

var bareDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func validPrefixedDigest(value string) bool {
	return len(value) == 71 && value[:7] == "sha256:" && bareDigest.MatchString(value[7:])
}

func validBareDigest(value string) bool {
	return bareDigest.MatchString(value)
}

func validCoverageReportDigest(report artifactcoverage.Report) bool {
	expected := report.ReportDigest
	report.ReportDigest = ""
	return validPrefixedDigest(expected) && digestJSON(report) == expected
}
