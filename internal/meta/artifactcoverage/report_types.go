package artifactcoverage

type OperationWitness struct {
	Operation         string `json:"operation"`
	IndicatorCount    int    `json:"indicator_count"`
	ArtifactPattern   string `json:"artifact_pattern,omitempty"`
	ExpectedArtifact  string `json:"expected_artifact,omitempty"`
	EvidenceKey       string `json:"evidence_key,omitempty"`
	ObservedArtifacts int    `json:"observed_artifacts"`
	ExactHead         bool   `json:"exact_head"`
	DigestBound       bool   `json:"digest_bound"`
	ReplayBound       bool   `json:"replay_bound"`
	Canonical         bool   `json:"canonical"`
	Status            string `json:"status"`
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
	Schema              string             `json:"schema"`
	CommitSHA           string             `json:"commit_sha"`
	Repository          string             `json:"repository"`
	RunID               int64              `json:"run_id"`
	RunAttempt          int                `json:"run_attempt"`
	ActionabilityDigest string             `json:"actionability_digest"`
	ObservationDigest   string             `json:"observation_digest"`
	ProgramDigest       string             `json:"program_digest"`
	AuthorityDigest     string             `json:"authority_digest"`
	Decision            string             `json:"decision"`
	Reason              string             `json:"reason"`
	SelectedOperation   string             `json:"selected_operation,omitempty"`
	Summary             Summary            `json:"summary"`
	Indicators          []KPI              `json:"indicators"`
	Operations          []OperationWitness `json:"operations"`
	Proofs              []Proof            `json:"proofs"`
	ReportDigest        string             `json:"report_digest"`
}
