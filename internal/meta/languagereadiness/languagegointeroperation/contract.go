package languagegointeroperation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"

const (
	ReportSchema          = "gooo/language-go-interoperation/v1"
	RegistrySchema        = "gooo/language-go-interoperation-cases/v1"
	RegistryVersion       = "2026-08-23"
	ConceptID             = "language-go-interoperation"
	ExpectedMetaOperation = "reify-go-projection-and-prove-type-identity"
	RequiredGoVersion     = "go1.27"
	FixedTotal            = 24
	FixedGeneratorCases   = 8
	FixedGo127Cases       = 8
	FixedGuardrailCases   = 8
	FixedIndicators       = 18
)

type Decision string
type Resolution string
type CaseKind string
type CaseStatus string

const (
	DecisionPass       Decision   = "PASS"
	DecisionFailClosed Decision   = "FAIL_CLOSED"
	ResolutionExact    Resolution = "EXACT"
	ResolutionLower    Resolution = "LOWER_RESOLUTION"
	CaseGenerator      CaseKind   = "GENERATOR"
	CaseGo127          CaseKind   = "GO_1_27"
	CaseGuardrail      CaseKind   = "GUARDRAIL"
	StatusSatisfied    CaseStatus = "SATISFIED"
	StatusNotSatisfied CaseStatus = "NOT_SATISFIED"
)

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact languageconcept.Artifact
}

type Definition struct {
	ID              string   `json:"id"`
	Kind            CaseKind `json:"kind"`
	Fixture         string   `json:"fixture"`
	ExpectedOutcome string   `json:"expected_outcome"`
	ExpectedStage   string   `json:"expected_stage"`
	ProofChoice     string   `json:"proof_choice"`
	MetaOperation   string   `json:"meta_operation"`
}

type CaseRegistry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}
