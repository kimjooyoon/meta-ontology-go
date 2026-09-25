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

type Report struct {
	Schema             string                 `json:"schema"`
	SubjectSHA         string                 `json:"subject_sha"`
	EvidenceSubjectSHA string                 `json:"evidence_subject_sha"`
	Decision           string                 `json:"decision"`
	Resolution         string                 `json:"resolution"`
	EnforcementEffect  string                 `json:"enforcement_effect"`
	Reason             string                 `json:"reason"`
	DenominatorID      string                 `json:"denominator_id"`
	DenominatorDigest  string                 `json:"denominator_digest"`
	Artifacts          []ArtifactBinding      `json:"artifacts"`
	Transition         Transition             `json:"transition"`
	Summary            Summary                `json:"summary"`
	Indicators         []Indicator            `json:"indicators"`
	MetaOperations     []MetaOperationBinding `json:"meta_operations"`
	RepositoryWrites   int                    `json:"repository_writes"`
	PromotionApplied   int                    `json:"promotion_applied"`
	ReportDigest       string                 `json:"report_digest"`
}
