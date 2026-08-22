package predecessorselection

import readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"

type Selection struct {
	RunID                 int64  `json:"run_id"`
	RunAttempt            int    `json:"run_attempt"`
	WorkflowConclusion    string `json:"workflow_conclusion"`
	ProducerJobID         int64  `json:"producer_job_id"`
	ProducerJobRunAttempt int    `json:"producer_job_run_attempt"`
	ProducerJobName       string `json:"producer_job_name"`
	ProducerJobConclusion string `json:"producer_job_conclusion"`
	ReadinessArtifactID   int64  `json:"readiness_artifact_id"`
	BindingArtifactID     int64  `json:"binding_artifact_id"`

	Baseline readinessartifact.BaselineReference `json:"baseline"`
}

type Summary struct {
	ObservedCandidates   int `json:"observed_candidates"`
	ExactHeadCandidates  int `json:"exact_head_candidates"`
	CanonicalCandidates  int `json:"canonical_candidates"`
	SuccessfulCandidates int `json:"successful_candidates"`
	ProducerConformantCandidates int `json:"producer_conformant_candidates"`
	AvailableCandidates  int `json:"available_candidates"`
	ValidCandidates      int `json:"valid_candidates"`
	AmbiguousCandidates  int `json:"ambiguous_candidates"`
	RepositoryWrites     int `json:"repository_writes"`
}

type Proof struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Passed bool   `json:"passed"`
}

type Report struct {
	Schema         string     `json:"schema"`
	Repository     string     `json:"repository"`
	CurrentHeadSHA string     `json:"current_head_sha"`
	PredecessorSHA string     `json:"predecessor_sha"`
	Decision       string     `json:"decision"`
	Reason         string     `json:"reason"`
	Selected       *Selection `json:"selected,omitempty"`
	Summary        Summary    `json:"summary"`
	Proofs         []Proof    `json:"proofs"`
	ReportDigest   string     `json:"report_digest"`
}

type Result struct {
	Report      Report
	BaselineRaw []byte
	BindingRaw  []byte
}
