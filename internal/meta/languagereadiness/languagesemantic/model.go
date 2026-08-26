package languagesemantic

const (
	ReportSchema   = "gooo/language-semantic-model/v1"
	RegistrySchema = "gooo/language-semantic-model-corpus/v1"
	ConceptID      = "language-semantic-model"
	FixedTotal     = 27
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

	CaseSource            CaseKind = "SOURCE"
	CaseLaw               CaseKind = "LAW"
	CaseUpstreamRejection CaseKind = "UPSTREAM_REJECTION"

	StatusSatisfied    CaseStatus = "SATISFIED"
	StatusNotSatisfied CaseStatus = "NOT_SATISFIED"
	StatusUnresolved   CaseStatus = "UNRESOLVED"
)

type Definition struct {
	ID            string   `json:"id"`
	Kind          CaseKind `json:"kind"`
	Path          string   `json:"path,omitempty"`
	Law           string   `json:"law,omitempty"`
	UpstreamCase  string   `json:"upstream_case,omitempty"`
	ProofChoice   string   `json:"proof_choice"`
	MetaOperation string   `json:"meta_operation"`
}

type Registry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}
