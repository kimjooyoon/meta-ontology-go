package selfimprovementtransport

type LifecycleIndicator struct {
	Ordinal          int        `json:"ordinal"`
	MetricID         string     `json:"metric_id"`
	Class            string     `json:"class"`
	ProofChoice      string     `json:"proof_choice"`
	Coordinate       Coordinate `json:"coordinate"`
	MetaOperation    string     `json:"meta_operation"`
	Value            int        `json:"value"`
	Target           int        `json:"target"`
	Status           string     `json:"status"`
	Reason           string     `json:"reason"`
	ObservationClass string     `json:"observation_class,omitempty"`
	ExpectedDigest   string     `json:"expected_digest,omitempty"`
	ObservedDigest   string     `json:"observed_digest,omitempty"`
	EvidenceDigest   string     `json:"evidence_digest,omitempty"`
}

type LifecycleClaim struct {
	ClaimID        string `json:"claim_id"`
	Stage          string `json:"stage"`
	Statement      string `json:"statement"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type LifecycleMetrics struct {
	FixedStepTotal      int `json:"fixed_step_total"`
	VerifiedTotal       int `json:"verified_total"`
	UnknownTotal        int `json:"unknown_total"`
	OpenTotal           int `json:"open_total"`
	DischargedTotal     int `json:"discharged_total"`
	CoverageBasisPoints int `json:"coverage_basis_points"`
	UnknownPathCount    int `json:"unknown_path_count"`
}

type LifecycleAuthority struct {
	Execution        bool `json:"execution"`
	Mutation         bool `json:"mutation"`
	Promotion        bool `json:"promotion"`
	RepositoryWrites int  `json:"repository_writes"`
	ExternalWrites   int  `json:"external_repository_writes"`
}

type LifecycleReceipt struct {
	Schema                string               `json:"schema"`
	MetricID              string               `json:"metric_id"`
	DenominatorID         string               `json:"denominator_id"`
	Contract              ContractEvidence     `json:"contract"`
	Repository            string               `json:"repository"`
	ExpectedRunID         int64                `json:"expected_run_id"`
	ExpectedRunAttempt    int                  `json:"expected_run_attempt"`
	OrchestrationHeadSHA  string               `json:"orchestration_head_sha,omitempty"`
	WorkflowPath          string               `json:"workflow_path,omitempty"`
	ArtifactID            int64                `json:"artifact_id,omitempty"`
	ArtifactName          string               `json:"artifact_name"`
	ArtifactDigest        string               `json:"artifact_digest,omitempty"`
	ArtifactSizeBytes     int64                `json:"artifact_size_bytes,omitempty"`
	ArtifactInstanceCount int                  `json:"artifact_instance_count"`
	ArtifactTypeCount     int                  `json:"artifact_type_count"`
	ActualArchiveDigest   string               `json:"actual_archive_digest,omitempty"`
	Decision              string               `json:"decision"`
	Resolution            string               `json:"resolution"`
	EnforcementEffect     string               `json:"enforcement_effect"`
	Reason                string               `json:"reason"`
	Coordinate            Coordinate           `json:"coordinate"`
	Indicators            []LifecycleIndicator `json:"indicators"`
	Claims                []LifecycleClaim     `json:"claims"`
	Metrics               LifecycleMetrics     `json:"metrics"`
	Authority             LifecycleAuthority   `json:"authority"`
	Digest                string               `json:"digest"`
}
