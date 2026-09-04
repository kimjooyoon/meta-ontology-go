package selfimprovementtransport

const (
	ContractID            = "gooo.contract.self-improvement.exact-head-transport.v1"
	ProducerSchema        = "gooo/self-improvement-transport-producer/v1"
	MetadataSchema        = "gooo/github-actions-artifact-transport/v1"
	ReportSchema          = "gooo/exact-head-cross-job-transport/v1"
	MetricID              = "gooo.metric.transport.eht8.v1"
	ObservationSchema     = "gooo/self-improvement-language-observation/v1"
	ArtifactName          = "self-improvement-language-observation"
	DecisionPass          = "PASS"
	DecisionObserved      = "OBSERVED"
	DecisionFailClosed    = "FAIL_CLOSED"
	ResolutionExact       = "EXACT"
	ResolutionLower       = "LOWER_RESOLUTION"
	StatusVerified        = "VERIFIED"
	StatusUnknown         = "UNKNOWN"
	StatusFalse           = "FALSE"
	ReasonComplete        = "EHT8_COMPLETE"
	ReasonAttestation     = "PRODUCER_ATTESTATION_UNKNOWN"
	ReasonKnownMismatch   = "KNOWN_TRANSPORT_MISMATCH"
	attestationObligation = "producer-attestation"
	fixedObligationTotal  = 8
)

type Coordinate struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
}

type ContractEvidence struct {
	ContractID       string           `json:"contract_id"`
	Path             string           `json:"path"`
	Package          string           `json:"package"`
	Namespace        string           `json:"namespace"`
	EntityCount      int              `json:"entity_count"`
	ActivityCount    int              `json:"activity_count"`
	ObligationTotal  int              `json:"obligation_total"`
	SourceLines      int              `json:"source_lines"`
	SourceDigest     string           `json:"source_digest"`
	CanonicalDigest  string           `json:"canonical_digest"`
	SemanticDigest   string           `json:"semantic_digest"`
	ResolutionPolicy ResolutionPolicy `json:"resolution_policy"`
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
