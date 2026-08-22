package feedbackstate

import "encoding/json"

const (
	ReportSchema   = "gooo/meta-feedback-semantic-snapshot/v1"
	ReceiptSchema  = "gooo/meta-artifact-feedback-resolution-receipt/v1"
	ResolutionSchema = "gooo/meta-artifact-feedback-resolution/v1"
)

type Input struct {
	Repository       string          `json:"repository"`
	PredecessorSHA   string          `json:"predecessor_sha"`
	Selection        Selection       `json:"selection"`
	PayloadDigest    string          `json:"payload_digest"`
	Receipt          json.RawMessage `json:"receipt"`
	RepositoryWrites int             `json:"repository_writes"`
}

type Selection struct {
	ArtifactID    int64  `json:"artifact_id"`
	RunID         int64  `json:"run_id"`
	RunAttempt    int    `json:"run_attempt"`
	ReceiptDigest string `json:"receipt_digest"`
}

type Snapshot struct {
	SourceDecision   string `json:"source_decision"`
	Decision         string `json:"decision"`
	Reason           string `json:"reason"`
	FromResolution   string `json:"from_resolution"`
	ToResolution     string `json:"to_resolution"`
	NextOperation    string `json:"next_operation"`
	PreviousDescents int    `json:"previous_descents"`
	Descents         int    `json:"descents"`
	ReceiptDigest    string `json:"receipt_digest"`
	ReportDigest     string `json:"report_digest"`
}

type Report struct {
	Schema         string      `json:"schema"`
	Repository     string      `json:"repository"`
	PredecessorSHA string      `json:"predecessor_sha"`
	Decision       string      `json:"decision"`
	Reason         string      `json:"reason"`
	Snapshot       *Snapshot   `json:"snapshot,omitempty"`
	Summary        Summary     `json:"summary"`
	Indicators     []Indicator `json:"indicators"`
	Proofs         []Proof     `json:"proofs"`
	ReportDigest   string      `json:"report_digest"`
}

type Summary struct {
	ReadinessBasisPoints int `json:"readiness_basis_points"`
	ResolutionDescents   int `json:"resolution_descents"`
	FalseFixedPoints     int `json:"false_fixed_points"`
	RepositoryWrites     int `json:"repository_writes"`
}
