package sourceauthorityeval

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthority"
)

const (
	InputSchema  = "gooo/source-backed-authority-evidence/v1"
	ReportSchema = "gooo/source-backed-authority-report/v1"
)

func Evaluate(bundle Bundle) Report {
	report := newReport(bundle)
	if _, err := sourceauthority.Load(); err != nil {
		setOutcome(&report, "ERROR", "INVARIANT_ONLY", "BLOCK",
			"SOURCE_AUTHORITY_CONTRACT_UNAVAILABLE")
		return seal(report)
	}
	if bundle.Schema != InputSchema {
		setOutcome(&report, "ERROR", "EXACT", "BLOCK", "EVIDENCE_SCHEMA_UNKNOWN")
		return seal(report)
	}
	if !isExactSHA(bundle.SubjectSHA) {
		setOutcome(&report, "UNKNOWN", "INVARIANT_ONLY", "BLOCK",
			"SUBJECT_SHA_UNKNOWN")
		return seal(report)
	}
	if bundle.ContractDigest != report.ContractDigest {
		setOutcome(&report, "UNKNOWN", "INVARIANT_ONLY", "BLOCK",
			"CONTRACT_DIGEST_MISMATCH")
		return seal(report)
	}
	index, reason := buildIndex(bundle)
	if reason != "" {
		setOutcome(&report, "ERROR", "EXACT", "BLOCK", reason)
		report.Summary.ErrorFacts = 1
		return seal(report)
	}
	facts, reason := acceptedFacts(bundle.Facts)
	if reason != "" {
		setOutcome(&report, "ERROR", "EXACT", "BLOCK", reason)
		report.Summary.ErrorFacts = 1
		return seal(report)
	}
	if len(facts) == 0 {
		setOutcome(&report, "UNKNOWN", "INVARIANT_ONLY", "BLOCK",
			"SOURCE_AUTHORITY_DENOMINATOR_EMPTY")
		return seal(report)
	}
	for _, fact := range facts {
		report.Receipts = append(report.Receipts, evaluateFact(fact, index))
	}
	return finalize(report)
}

func isExactSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
