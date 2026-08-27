package selfimprovementattestation

type Request struct {
	TransportReceipt         TransportReceipt
	ArchiveDigest            string
	ArchiveProducer          Producer
	ArchiveObservationDigest string
	Verification             []VerificationItem
	VerifierExitCode         int
	VerifierVersion          string
}

type Checker struct {
	Name                string `json:"name"`
	Version             string `json:"version"`
	ExitCode            int    `json:"exit_code"`
	VerifiedResultTotal int    `json:"verified_result_total"`
}

type ProducerIdentity struct {
	WorkflowRef        string `json:"workflow_ref"`
	WorkflowSHA        string `json:"workflow_sha"`
	RunID              int64  `json:"run_id"`
	RunAttempt         int    `json:"run_attempt"`
	ArtifactID         int64  `json:"artifact_id"`
	SubjectName        string `json:"subject_name"`
	SubjectDigest      string `json:"subject_digest"`
	SignerURI          string `json:"signer_uri"`
	Issuer             string `json:"issuer"`
	RunnerEnvironment  string `json:"runner_environment"`
	VerifiedTimestamps int    `json:"verified_timestamps"`
}

type ClaimTransition struct {
	ClaimID        string `json:"claim_id"`
	Before         string `json:"before"`
	After          string `json:"after"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type ReaderView struct {
	Audience            string `json:"audience"`
	Resolution          string `json:"resolution"`
	VerifiedTotal       int    `json:"verified_total"`
	FixedTotal          int    `json:"fixed_total"`
	CoverageBasisPoints int    `json:"coverage_basis_points"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Authority struct {
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthorized          bool `json:"mutation_authorized"`
	ExecutionAuthorized         bool `json:"execution_authorized"`
	PromotionAuthorized         bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized bool `json:"automatic_adoption_authorized"`
}
