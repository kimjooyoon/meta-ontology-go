package selfimprovementtransport

type ProducerReceipt struct {
	Schema        string           `json:"schema"`
	Contract      ContractEvidence `json:"contract"`
	RepositoryURI string           `json:"repository_uri"`
	SubjectSHA    string           `json:"subject_sha"`
	CheckoutSHA   string           `json:"checkout_sha"`
	WorkflowRef   string           `json:"workflow_ref"`
	WorkflowSHA   string           `json:"workflow_sha"`
	RunID         int64            `json:"run_id"`
	RunAttempt    int              `json:"run_attempt"`
	Job           string           `json:"job"`
	ArtifactName  string           `json:"artifact_name"`
	Subject       LogicalSubject   `json:"logical_subject"`
	Decision      string           `json:"decision"`
	Resolution    string           `json:"resolution"`
	Reason        string           `json:"reason"`
	Digest        string           `json:"digest"`
}

type Attestation struct {
	Status           string `json:"status,omitempty"`
	Digest           string `json:"digest,omitempty"`
	ProducerIdentity string `json:"producer_identity,omitempty"`
}

type TransportMetadata struct {
	Schema               string      `json:"schema"`
	Repository           string      `json:"repository"`
	ProducerRunID        int64       `json:"producer_run_id"`
	ProducerRunAttempt   int         `json:"producer_run_attempt"`
	OrchestrationHeadSHA string      `json:"orchestration_head_sha"`
	WorkflowPath         string      `json:"workflow_path"`
	ArtifactID           int64       `json:"artifact_id"`
	ArtifactName         string      `json:"artifact_name"`
	ArtifactDigest       string      `json:"artifact_digest"`
	ArtifactSizeBytes    int64       `json:"artifact_size_bytes"`
	Attestation          Attestation `json:"attestation"`
}

type Obligation struct {
	ID             string     `json:"id"`
	ProofRoute     string     `json:"proof_route"`
	Coordinate     Coordinate `json:"coordinate"`
	Status         string     `json:"status"`
	Reason         string     `json:"reason"`
	EvidenceDigest string     `json:"evidence_digest,omitempty"`
}

type Metrics struct {
	FixedObligationTotal int `json:"fixed_obligation_total"`
	VerifiedTotal        int `json:"verified_total"`
	UnknownTotal         int `json:"unknown_total"`
	FalseTotal           int `json:"false_total"`
	OpenTotal            int `json:"open_total"`
	CoverageBasisPoints  int `json:"coverage_basis_points"`
	FalsePromotionCount  int `json:"false_promotion_count"`
}
