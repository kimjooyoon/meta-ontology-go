package languagesemantic

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"

const (
	ReportSchema   = "gooo/language-semantic-model/v1"
	RegistrySchema = "gooo/language-semantic-model-corpus/v1"
	ConceptID      = "language-semantic-model"
	FixedTotal     = 18
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
	ID           string   `json:"id"`
	Kind         CaseKind `json:"kind"`
	Path         string   `json:"path,omitempty"`
	Law          string   `json:"law,omitempty"`
	UpstreamCase string   `json:"upstream_case,omitempty"`
	ProofChoice  string   `json:"proof_choice"`
	MetaOperation string  `json:"meta_operation"`
}

type Registry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}

type UpstreamEvidence struct {
	CaseID           string   `json:"case_id"`
	ObservedDecision string   `json:"observed_decision"`
	Diagnostics      []string `json:"diagnostics,omitempty"`
}

type LawEvidence struct {
	Law         string                `json:"law"`
	Satisfied   bool                  `json:"satisfied"`
	Observation replay.LawObservation `json:"observation"`
}

type CaseEvidence struct {
	Source   *replay.Observation `json:"source,omitempty"`
	Law      *LawEvidence        `json:"law,omitempty"`
	Upstream *UpstreamEvidence   `json:"upstream,omitempty"`
	Error    string              `json:"error,omitempty"`
}

type CaseResult struct {
	Definition Definition   `json:"definition"`
	Evidence   CaseEvidence `json:"evidence"`
	Status     CaseStatus   `json:"status"`
	Digest     string       `json:"evidence_digest"`
}

type SyntaxSummary struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	ValidCases  int `json:"valid_cases"`
	InvalidCases int `json:"invalid_cases"`
	GoooLines   int `json:"gooo_lines"`
}

type GoooFile struct {
	Path         string `json:"path"`
	GoooLines    int    `json:"gooo_lines"`
	SourceDigest string `json:"source_digest"`
}

type Source struct {
	ExpectedHeadSHA      string        `json:"expected_head_sha"`
	ConceptID            string        `json:"concept_id"`
	Producer             string        `json:"producer"`
	Consumer             string        `json:"consumer"`
	MetaOperation        string        `json:"meta_operation"`
	RegistryDigest       string        `json:"registry_digest"`
	SyntaxArtifactDigest string        `json:"syntax_artifact_digest"`
	SyntaxReportDigest   string        `json:"syntax_report_digest"`
	SyntaxSummary        SyntaxSummary `json:"syntax_summary"`
	GoooFiles            []GoooFile    `json:"gooo_files"`
	ObservationKnown     bool          `json:"observation_known"`
	ConceptBound         bool          `json:"concept_bound"`
}

type Summary struct {
	Satisfied                  int `json:"satisfied"`
	Total                      int `json:"total"`
	Executed                   int `json:"executed"`
	NotSatisfied               int `json:"not_satisfied"`
	Unresolved                 int `json:"unresolved"`
	ReadinessBPS               int `json:"readiness_bps"`
	SourceModels               int `json:"source_models"`
	NormalizedIRs              int `json:"normalized_irs"`
	SemanticReplays            int `json:"semantic_replays"`
	ProvenanceReplays          int `json:"provenance_replays"`
	EvidenceReplays            int `json:"evidence_replays"`
	PresentationLaws           int `json:"presentation_laws"`
	CandidateAuthorityLaws     int `json:"candidate_authority_laws"`
	DeterministicAuthorityLaws int `json:"deterministic_authority_laws"`
	UpstreamRejections         int `json:"upstream_rejections"`
	UnregisteredGooo           int `json:"unregistered_gooo"`
	MissingRegistered          int `json:"missing_registered"`
	StageOrderViolations       int `json:"stage_order_violations"`
	EffectfulStages            int `json:"effectful_stages"`
	RegistryDrift              int `json:"registry_drift"`
}

type Indicator struct {
	MetricID     string     `json:"metric_id"`
	Class        string     `json:"class"`
	ProofChoice  string     `json:"proof_choice"`
	Producer     string     `json:"producer"`
	Consumer     string     `json:"consumer"`
	MetaOperation string    `json:"meta_operation"`
	Resolution   Resolution `json:"resolution"`
	Value        int        `json:"value"`
	Target       int        `json:"target"`
	Satisfied    bool       `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema             string      `json:"schema"`
	Decision           Decision    `json:"decision"`
	Resolution         Resolution  `json:"resolution"`
	ReasonCode         string      `json:"reason_code"`
	Source             Source      `json:"source"`
	Summary            Summary     `json:"summary"`
	Cases              []CaseResult `json:"cases"`
	Indicators         []Indicator `json:"indicators"`
	Proofs             []Proof     `json:"proofs"`
	RepositoryWrites   int         `json:"repository_writes"`
	MutationAuthorized bool        `json:"mutation_authorized"`
	ReportDigest       string      `json:"report_digest"`
}
