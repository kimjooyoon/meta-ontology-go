package verticalsliceclosureshadow

import (
	_ "embed"
	"encoding/json"

	languagesyntax "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

//go:embed evidence/assurance.json
var embeddedAssurance []byte

//go:embed evidence/denominator.json
var embeddedDenominator []byte

//go:embed evidence/denominator-v22.json
var embeddedDenominatorV22 []byte

//go:embed evidence/denominator-v23.json
var embeddedDenominatorV23 []byte

//go:embed evidence/denominator-v24.json
var embeddedDenominatorV24 []byte

//go:embed evidence/denominator-v25.json
var embeddedDenominatorV25 []byte

//go:embed evidence/denominator-v26.json
var embeddedDenominatorV26 []byte

//go:embed evidence/denominator-v27.json
var embeddedDenominatorV27 []byte

//go:embed evidence/denominator-v28.json
var embeddedDenominatorV28 []byte

//go:embed evidence/denominator-v29.json
var embeddedDenominatorV29 []byte

//go:embed evidence/denominator-v30.json
var embeddedDenominatorV30 []byte

func EmbeddedAssurance() []byte {
	return append([]byte(nil), embeddedAssurance...)
}

func EmbeddedDenominator() []byte {
	return append([]byte(nil), embeddedDenominator...)
}

func activeDenominator() []byte {
	switch languagesyntax.FixedCapabilityTotal {
	case 45:
		return embeddedDenominator
	case 46:
		return embeddedDenominatorV22
	case 47:
		return embeddedDenominatorV23
	case 48:
		return embeddedDenominatorV24
	case 49:
		return embeddedDenominatorV25
	case 50:
		return embeddedDenominatorV26
	case 52:
		return embeddedDenominatorV27
	case 56:
		return embeddedDenominatorV28
	case 57:
		return embeddedDenominatorV29
	case 58:
		return embeddedDenominatorV30
	default:
		return nil
	}
}

func activeDenominatorDigest() string {
	switch languagesyntax.FixedCapabilityTotal {
	case 45:
		return DenominatorDigest
	case 46:
		return DenominatorMigrationDigest
	case 47:
		return DenominatorMigrationV23Digest
	case 48:
		return DenominatorMigrationV24Digest
	case 49:
		return DenominatorMigrationV25Digest
	case 50:
		return DenominatorMigrationV26Digest
	case 52:
		return DenominatorMigrationV27Digest
	case 56:
		return DenominatorMigrationV28Digest
	case 57:
		return DenominatorMigrationV29Digest
	case 58:
		return DenominatorMigrationV30Digest
	default:
		return ""
	}
}

type assuranceSummary struct {
	DenominatorTotal     int `json:"denominator_total"`
	Operating            int `json:"operating"`
	NotImplemented       int `json:"not_implemented"`
	CoverageBPS          int `json:"implementation_coverage_bps"`
	UnknownTopDecisions  int `json:"unknown_top_decisions"`
	UnresolvedIndicators int `json:"unresolved_indicators"`
	ViolatedGuardrails   int `json:"violated_guardrails"`
	RepositoryWrites     int `json:"repository_writes"`
}

type assuranceObligation struct {
	MetricID      string `json:"metric_id"`
	Status        string `json:"status"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type assuranceReport struct {
	Schema            string                `json:"schema"`
	SubjectSHA        string                `json:"subject_sha"`
	AssuranceDecision string                `json:"assurance_decision"`
	CandidateDecision string                `json:"candidate_decision"`
	Summary           assuranceSummary      `json:"summary"`
	Obligations       []assuranceObligation `json:"obligations"`
}

func inspectAssurance(raw []byte) (string, string, string) {
	if len(raw) == 0 {
		return "", ResolutionLower, ReasonAssuranceMissing
	}
	if digestBytes(raw) != AssuranceDigest {
		return "", ResolutionInvariant, ReasonAssuranceDigest
	}
	var report assuranceReport
	if json.Unmarshal(raw, &report) != nil || !validAssurance(report) {
		return "", ResolutionInvariant, ReasonAssuranceBase
	}
	return report.SubjectSHA, ResolutionExact, ""
}
