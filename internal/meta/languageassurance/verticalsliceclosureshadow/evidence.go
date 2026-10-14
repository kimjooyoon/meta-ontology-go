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

//go:embed evidence/denominator-v31.json
var embeddedDenominatorV31 []byte

//go:embed evidence/denominator-v32.json
var embeddedDenominatorV32 []byte

//go:embed evidence/denominator-v33.json
var embeddedDenominatorV33 []byte

//go:embed evidence/denominator-v34.json
var embeddedDenominatorV34 []byte

//go:embed evidence/denominator-v35.json
var embeddedDenominatorV35 []byte

//go:embed evidence/denominator-v36.json
var embeddedDenominatorV36 []byte

//go:embed evidence/denominator-v37.json
var embeddedDenominatorV37 []byte

//go:embed evidence/denominator-v38.json
var embeddedDenominatorV38 []byte

//go:embed evidence/denominator-v39.json
var embeddedDenominatorV39 []byte

//go:embed evidence/denominator-v40.json
var embeddedDenominatorV40 []byte

//go:embed evidence/denominator-v41.json
var embeddedDenominatorV41 []byte

//go:embed evidence/denominator-v42.json
var embeddedDenominatorV42 []byte

//go:embed evidence/denominator-v43.json
var embeddedDenominatorV43 []byte

//go:embed evidence/denominator-v44.json
var embeddedDenominatorV44 []byte

//go:embed evidence/denominator-v45.json
var embeddedDenominatorV45 []byte

//go:embed evidence/denominator-v46.json
var embeddedDenominatorV46 []byte

//go:embed evidence/denominator-v47.json
var embeddedDenominatorV47 []byte

func EmbeddedAssurance() []byte {
	return append([]byte(nil), embeddedAssurance...)
}

func EmbeddedDenominator() []byte {
	return append([]byte(nil), embeddedDenominator...)
}

func activeDenominator() []byte {
	if value := activeDenominatorChunk0(languagesyntax.FixedCapabilityTotal); value != nil {
		return value
	}
	if value := activeDenominatorChunk1(languagesyntax.FixedCapabilityTotal); value != nil {
		return value
	}
	return activeDenominatorChunk2(languagesyntax.FixedCapabilityTotal)
}

func activeDenominatorDigest() string {
	if value := activeDenominatorDigestChunk0(languagesyntax.FixedCapabilityTotal); value != "" {
		return value
	}
	if value := activeDenominatorDigestChunk1(languagesyntax.FixedCapabilityTotal); value != "" {
		return value
	}
	return activeDenominatorDigestChunk2(languagesyntax.FixedCapabilityTotal)
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

//go:embed evidence/denominator-v48.json
var embeddedDenominatorV48 []byte

//go:embed evidence/denominator-v49.json
var embeddedDenominatorV49 []byte

//go:embed evidence/denominator-v50.json
var embeddedDenominatorV50 []byte

//go:embed evidence/denominator-v51.json
var embeddedDenominatorV51 []byte

//go:embed evidence/denominator-v52.json
var embeddedDenominatorV52 []byte

//go:embed evidence/denominator-v53.json
var embeddedDenominatorV53 []byte

//go:embed evidence/denominator-v54.json
var embeddedDenominatorV54 []byte

//go:embed evidence/denominator-v55.json
var embeddedDenominatorV55 []byte

//go:embed evidence/denominator-v56.json
var embeddedDenominatorV56 []byte

//go:embed evidence/denominator-v57.json
var embeddedDenominatorV57 []byte

//go:embed evidence/denominator-v58.json
var embeddedDenominatorV58 []byte

//go:embed evidence/denominator-v59.json
var embeddedDenominatorV59 []byte

//go:embed evidence/denominator-v60.json
var embeddedDenominatorV60 []byte

//go:embed evidence/denominator-v61.json
var embeddedDenominatorV61 []byte

//go:embed evidence/denominator-v62.json
var embeddedDenominatorV62 []byte

//go:embed evidence/denominator-v63.json
var embeddedDenominatorV63 []byte

//go:embed evidence/denominator-v64.json
var embeddedDenominatorV64 []byte

//go:embed evidence/denominator-v65.json
var embeddedDenominatorV65 []byte

//go:embed evidence/denominator-v66.json
var embeddedDenominatorV66 []byte

//go:embed evidence/denominator-v67.json
var embeddedDenominatorV67 []byte

func activeDenominatorDigestChunk0(total int) string {
	switch total {
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
	case 60:
		return DenominatorMigrationV31Digest
	case 63:
		return DenominatorMigrationV32Digest
	case 64:
		return DenominatorMigrationV33Digest
	case 65:
		return DenominatorMigrationV34Digest
	case 66:
		return DenominatorMigrationV35Digest
	case 67:
		return DenominatorMigrationV36Digest
	}
	return ""
}

func activeDenominatorDigestChunk1(total int) string {
	switch total {
	case 68:
		return DenominatorMigrationV37Digest
	case 69:
		return DenominatorMigrationV38Digest
	case 73:
		return DenominatorMigrationV42Digest
	case 75:
		return DenominatorMigrationV43Digest
	case 77:
		return DenominatorMigrationV45Digest
	case 79:
		return DenominatorMigrationV47Digest
	case 80:
		return DenominatorMigrationV48Digest
	case 82:
		return DenominatorMigrationV49Digest
	case 83:
		return DenominatorMigrationV50Digest
	case 84:
		return DenominatorMigrationV51Digest
	case 85:
		return DenominatorMigrationV52Digest
	case 86:
		return DenominatorMigrationV53Digest
	case 87:
		return DenominatorMigrationV54Digest
	case 88:
		return DenominatorMigrationV55Digest
	case 89:
		return DenominatorMigrationV56Digest
	case 90:
		return DenominatorMigrationV57Digest
	}
	return ""
}

func activeDenominatorDigestChunk2(total int) string {
	switch total {
	case 91:
		return DenominatorMigrationV58Digest
	case 92:
		return DenominatorMigrationV59Digest
	case 93:
		return DenominatorMigrationV60Digest
	case 94:
		return DenominatorMigrationV61Digest
	case 95:
		return DenominatorMigrationV62Digest
	case 96:
		return DenominatorMigrationV63Digest
	case 97:
		return DenominatorMigrationV64Digest
	case 98:
		return DenominatorMigrationV65Digest
	case 99:
		return DenominatorMigrationV66Digest
	case 100:
		return DenominatorMigrationV67Digest
	}
	return ""
}

func activeDenominatorChunk0(total int) []byte {
	switch total {
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
	case 60:
		return embeddedDenominatorV31
	case 63:
		return embeddedDenominatorV32
	case 64:
		return embeddedDenominatorV33
	case 65:
		return embeddedDenominatorV34
	case 66:
		return embeddedDenominatorV35
	case 67:
		return embeddedDenominatorV36
	}
	return nil
}

func activeDenominatorChunk1(total int) []byte {
	switch total {
	case 68:
		return embeddedDenominatorV37
	case 69:
		return embeddedDenominatorV38
	case 73:
		return embeddedDenominatorV42
	case 75:
		return embeddedDenominatorV43
	case 77:
		return embeddedDenominatorV45
	case 79:
		return embeddedDenominatorV47
	case 80:
		return embeddedDenominatorV48
	case 82:
		return embeddedDenominatorV49
	case 83:
		return embeddedDenominatorV50
	case 84:
		return embeddedDenominatorV51
	case 85:
		return embeddedDenominatorV52
	case 86:
		return embeddedDenominatorV53
	case 87:
		return embeddedDenominatorV54
	case 88:
		return embeddedDenominatorV55
	case 89:
		return embeddedDenominatorV56
	case 90:
		return embeddedDenominatorV57
	}
	return nil
}

func activeDenominatorChunk2(total int) []byte {
	switch total {
	case 91:
		return embeddedDenominatorV58
	case 92:
		return embeddedDenominatorV59
	case 93:
		return embeddedDenominatorV60
	case 94:
		return embeddedDenominatorV61
	case 95:
		return embeddedDenominatorV62
	case 96:
		return embeddedDenominatorV63
	case 97:
		return embeddedDenominatorV64
	case 98:
		return embeddedDenominatorV65
	case 99:
		return embeddedDenominatorV66
	case 100:
		return embeddedDenominatorV67
	}
	return nil
}
