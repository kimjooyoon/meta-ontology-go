package selfimprovementattestation

type LogicalSubject struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Bytes  int64  `json:"bytes"`
}

type Producer struct {
	Schema         string         `json:"schema"`
	Contract       Contract       `json:"contract"`
	RepositoryURI  string         `json:"repository_uri"`
	SubjectSHA     string         `json:"subject_sha"`
	CheckoutSHA    string         `json:"checkout_sha"`
	WorkflowRef    string         `json:"workflow_ref"`
	WorkflowSHA    string         `json:"workflow_sha"`
	RunID          int64          `json:"run_id"`
	RunAttempt     int            `json:"run_attempt"`
	Job            string         `json:"job"`
	ArtifactName   string         `json:"artifact_name"`
	LogicalSubject LogicalSubject `json:"logical_subject"`
	Decision       string         `json:"decision"`
	Resolution     string         `json:"resolution"`
	Reason         string         `json:"reason"`
	Digest         string         `json:"digest"`
}

type Transport struct {
	Schema               string `json:"schema"`
	Repository           string `json:"repository"`
	ProducerRunID        int64  `json:"producer_run_id"`
	ProducerRunAttempt   int    `json:"producer_run_attempt"`
	OrchestrationHeadSHA string `json:"orchestration_head_sha"`
	WorkflowPath         string `json:"workflow_path"`
	ArtifactID           int64  `json:"artifact_id"`
	ArtifactName         string `json:"artifact_name"`
	ArtifactDigest       string `json:"artifact_digest"`
	ArtifactSizeBytes    int64  `json:"artifact_size_bytes"`
}

type TransportReceipt struct {
	Schema                  string       `json:"schema"`
	MetricID                string       `json:"metric_id"`
	Contract                Contract     `json:"contract"`
	SubjectSHA              string       `json:"subject_sha"`
	OrchestrationHeadSHA    string       `json:"orchestration_head_sha"`
	SourceObservationDigest string       `json:"source_observation_digest"`
	ActualArchiveDigest     string       `json:"actual_archive_digest"`
	Decision                string       `json:"decision"`
	Resolution              string       `json:"resolution"`
	Reason                  string       `json:"reason"`
	Coordinate              Coordinate   `json:"coordinate"`
	Producer                Producer     `json:"producer"`
	Transport               Transport    `json:"transport"`
	Obligations             []Obligation `json:"obligations"`
	OpenObligationIDs       []string     `json:"open_obligation_ids"`
	Metrics                 Metrics      `json:"metrics"`
	NotClaimed              []string     `json:"not_claimed"`
	Digest                  string       `json:"digest"`
}
