package languagedeterministicquery

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"

const (
	ReportSchema          = "gooo/language-deterministic-query/v1"
	RegistrySchema        = "gooo/language-deterministic-query-plans/v1"
	RegistryVersion       = "2026-08-23"
	ConceptID             = "language-deterministic-query"
	ExpectedMetaOperation = "execute-reified-deterministic-query-plan"
	FixedTotal            = 32
	FixedBindingPlans     = 28
	FixedLawPlans         = 4
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
	CaseBinding        CaseKind   = "BINDING"
	CaseLaw            CaseKind   = "LAW"
	StatusSatisfied    CaseStatus = "SATISFIED"
	StatusNotSatisfied CaseStatus = "NOT_SATISFIED"
)

const (
	BindingConcept = "CONCEPT"
	BindingCode    = "CODE"
	BindingMetric  = "METRIC"
	BindingUseCase = "USE_CASE"
)

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact languageconcept.Artifact
}

type Definition struct {
	ID            string   `json:"id"`
	Kind          CaseKind `json:"kind"`
	BindingClass  string   `json:"binding_class,omitempty"`
	Binding       string   `json:"binding,omitempty"`
	ProofChoice   string   `json:"proof_choice"`
	MetaOperation string   `json:"meta_operation"`
}

type PlanRegistry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}
