package selfimprovementtransport

const (
	ContractID             = "gooo.contract.self-improvement.exact-head-transport.v1"
	ProducerSchema         = "gooo/self-improvement-transport-producer/v1"
	MetadataSchema         = "gooo/github-actions-artifact-transport/v1"
	ReportSchema           = "gooo/exact-head-cross-job-transport/v1"
	MetricID               = "gooo.metric.transport.eht8.v1"
	ObservationSchema      = "gooo/self-improvement-language-observation/v1"
	ArtifactName           = "self-improvement-language-observation"
	DecisionPass           = "PASS"
	DecisionObserved       = "OBSERVED"
	DecisionFailClosed     = "FAIL_CLOSED"
	ResolutionExact        = "EXACT"
	ResolutionLower        = "LOWER_RESOLUTION"
	StatusVerified         = "VERIFIED"
	StatusUnknown          = "UNKNOWN"
	StatusFalse            = "FALSE"
	ReasonComplete         = "EHT8_COMPLETE"
	ReasonAttestation      = "PRODUCER_ATTESTATION_UNKNOWN"
	ReasonKnownMismatch    = "KNOWN_TRANSPORT_MISMATCH"
	attestationObligation  = "producer-attestation"
	fixedObligationTotal   = 8
)

type Coordinate struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
}

type ContractEvidence struct {
	ContractID      string `json:"contract_id"`
	Path            string `json:"path"`
	Package         string `json:"package"`
	Namespace       string `json:"namespace"`
	EntityCount     int    `json:"entity_count"`
	ActivityCount   int    `json:"activity_count"`
	ObligationTotal int    `json:"obligation_total"`
	SourceLines     int    `json:"source_lines"`
	SourceDigest    string `json:"source_digest"`
	CanonicalDigest string `json:"canonical_digest"`
}

type ProducerInput struct {
	Repository   string
	SubjectSHA   string
	CheckoutSHA  string
	WorkflowRef  string
	WorkflowSHA  string
	RunID        int64
	RunAttempt   int
	Job          string
	ArtifactName string
}

type LogicalSubject struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}

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
	Status         string `json:"status,omitempty"`
	Digest         string `json:"digest,omitempty"`
	ProducerIdentity string `json:"producer_identity,omitempty"`
}

type TransportMetadata struct {
	Schema                 string      `json:"schema"`
	Repository             string      `json:"repository"`
	ProducerRunID          int64       `json:"producer_run_id"`
	ProducerRunAttempt     int         `json:"producer_run_attempt"`
	OrchestrationHeadSHA   string      `json:"orchestration_head_sha"`
	WorkflowPath           string      `json:"workflow_path"`
	ArtifactID             int64       `json:"artifact_id"`
	ArtifactName           string      `json:"artifact_name"`
	ArtifactDigest         string      `json:"artifact_digest"`
	ArtifactSizeBytes      int64       `json:"artifact_size_bytes"`
	Attestation            Attestation `json:"attestation,omitempty"`
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

type Report struct {
	Schema                string            `json:"schema"`
	MetricID              string            `json:"metric_id"`
	Contract              ContractEvidence  `json:"contract"`
	SubjectSHA            string            `json:"subject_sha"`
	OrchestrationHeadSHA  string            `json:"orchestration_head_sha"`
	SourceObservationDigest string          `json:"source_observation_digest"`
	ActualArchiveDigest   string            `json:"actual_archive_digest"`
	Decision              string            `json:"decision"`
	Resolution            string            `json:"resolution"`
	Reason                string            `json:"reason"`
	Coordinate            Coordinate        `json:"coordinate"`
	Producer              ProducerReceipt   `json:"producer"`
	Transport             TransportMetadata `json:"transport"`
	Obligations           []Obligation      `json:"obligations"`
	OpenObligationIDs     []string          `json:"open_obligation_ids"`
	Metrics               Metrics           `json:"metrics"`
	NotClaimed            []string          `json:"not_claimed"`
	Digest                string            `json:"digest"`
}

type observationHeader struct {
	Schema     string `json:"schema"`
	SubjectSHA string `json:"subject_sha"`
}
