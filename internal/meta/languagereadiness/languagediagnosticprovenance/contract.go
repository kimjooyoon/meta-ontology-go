package languagediagnosticprovenance

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"

const (
	ReportSchema          = "gooo/language-diagnostic-provenance/v1"
	RegistrySchema        = "gooo/language-diagnostic-provenance-cases/v1"
	RegistryVersion       = "2026-08-23"
	ConceptID             = "language-diagnostic-provenance"
	ExpectedMetaOperation = "trace-diagnostic-through-semantic-provenance"
	RequiredGoVersion     = "go1.27"
	FixedTotal            = 18
	FixedSyntaxCases      = 3
	FixedTypeCases        = 3
	FixedSourceMapCases   = 4
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
	CaseSyntax         CaseKind   = "SYNTAX"
	CaseType           CaseKind   = "TYPE"
	CaseSourceMap      CaseKind   = "SOURCE_MAP"
	CaseGuardrail      CaseKind   = "GUARDRAIL"
	StatusSatisfied    CaseStatus = "SATISFIED"
	StatusNotSatisfied CaseStatus = "NOT_SATISFIED"
)

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact languageconcept.Artifact
	Registry        CaseRegistry
}

type Definition struct {
	ID              string   `json:"id"`
	Kind            CaseKind `json:"kind"`
	Fixture         string   `json:"fixture"`
	ExpectedOutcome string   `json:"expected_outcome"`
	ExpectedStage   string   `json:"expected_stage"`
	ExpectedReason  string   `json:"expected_reason,omitempty"`
	ProofChoice     string   `json:"proof_choice"`
	MetaOperation   string   `json:"meta_operation"`
	GuardrailClass  string   `json:"guardrail_class,omitempty"`
}

type CaseRegistry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}
