package artifactfeedback

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"

const (
	CycleSchema	= "gooo/self-improvement-cycle-envelope/v3"
	ReportSchema	= "gooo/meta-operation-artifact-feedback-report/v1"
)

type CycleObservation struct {
	Schema              string `json:"schema"`
	HeadSHA             string `json:"head_sha"`
	Status              string `json:"status"`
	CIConclusion        string `json:"ci_conclusion"`
	EnvelopeDigest      string `json:"envelope_digest"`
	ReplayDigest        string `json:"replay_digest"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

type Input struct {
	Coverage             artifactcoverage.Report `json:"coverage"`
	CoverageReplayDigest string                  `json:"coverage_replay_digest"`
	Cycle                CycleObservation        `json:"cycle"`
	RepositoryWrites     int                     `json:"repository_writes"`
}

type Summary struct {
	RequiredInputs          int `json:"required_inputs"`
	ExactHeadInputs         int `json:"exact_head_inputs"`
	BoundInputs             int `json:"bound_inputs"`
	ReplayBoundInputs       int `json:"replay_bound_inputs"`
	StaleInputs             int `json:"stale_inputs"`
	AmbiguousNextOperations int `json:"ambiguous_next_operations"`
	RepositoryWrites        int `json:"repository_writes"`
	ReadinessBasisPoints    int `json:"readiness_basis_points"`
}

type KPI struct {
	Indicator
	Value     int  `json:"value"`
	Satisfied bool `json:"satisfied"`
}

type Proof struct {
	Choice         ProofChoice `json:"choice"`
	MetaOperation  string      `json:"meta_operation"`
	Activity       string      `json:"activity"`
	Satisfied      bool        `json:"satisfied"`
	EvidenceDigest string      `json:"evidence_digest"`
}

type Report struct {
	Schema               string    `json:"schema"`
	CommitSHA            string    `json:"commit_sha"`
	Repository           string    `json:"repository"`
	Decision             string    `json:"decision"`
	Reason               string    `json:"reason"`
	NextOperation        string    `json:"next_operation,omitempty"`
	CoverageReportDigest string    `json:"coverage_report_digest"`
	CycleEnvelopeDigest  string    `json:"cycle_envelope_digest"`
	ProgramDigest        string    `json:"program_digest"`
	InputDigest          string    `json:"input_digest"`
	Summary              Summary   `json:"summary"`
	Indicators           []KPI     `json:"indicators"`
	Proofs               []Proof   `json:"proofs"`
	ReportDigest         string    `json:"report_digest"`
}
