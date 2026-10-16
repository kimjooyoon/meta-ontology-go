package conceptoperation

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

const (
	ObservationDocumentSchema = "gooo/concept-operation-observations/v1"
	ObservedReceiptSchema     = "gooo/observed-concept-operation-binding/v1"
	ObservationPolicyPath     = "examples/source-owned-cohort-observations/main.gooo"
	ObservationPolicyActivity = "AssessDeclared"
)

// ObservationInputs carries data separately from the executable strategy.
// Unknown observed operation IDs are data, not implementation requests.
type ObservationInputs struct {
	Document     []byte
	PolicySource []byte
}

type OperationObservation struct {
	Operation string                  `json:"operation"`
	Binding   *metricstrategy.Binding `json:"binding,omitempty"`
}

type ObservationDocument struct {
	Schema               string                 `json:"schema"`
	Repository           string                 `json:"repository"`
	SubjectSHA           string                 `json:"subject_sha"`
	ProgramDigest        string                 `json:"program_digest"`
	DeclaredCohortDigest string                 `json:"declared_cohort_digest"`
	Observations         []OperationObservation `json:"observations"`
}

type ObservationAssessment struct {
	ExpectedCount   int      `json:"expected_count"`
	BoundCount      int      `json:"bound_count"`
	UnknownCount    int      `json:"unknown_count"`
	CoverageBPS     int      `json:"coverage_bps"`
	Status          string   `json:"status"`
	UnknownClass    string   `json:"unknown_class"`
	Reason          string   `json:"reason"`
	NextOperation   string   `json:"next_operation"`
	ObservedIDs     []string `json:"observed_ids"`
	ReadinessCredit bool     `json:"readiness_credit"`
}

// ObservedReceipt keeps the original authenticated producer facts intact.
// Decision.Value is the separate source-produced assessment of observations.
type ObservedReceipt struct {
	Schema                    string                 `json:"schema"`
	MetricID                  string                 `json:"metric_id"`
	Baseline                  Receipt                `json:"baseline"`
	DeclaredCohortDigest       string                 `json:"declared_cohort_digest"`
	ObservationInputDigest    string                 `json:"observation_input_digest"`
	Decision                  bodycodegen.FixedValue `json:"decision"`
	Producer                  string                 `json:"producer"`
	Consumer                  string                 `json:"consumer"`
	RepositoryWorkspaceWrites bool                   `json:"repository_workspace_writes"`
	PromotionAuthorized       bool                   `json:"promotion_authorized"`
	Digest                    string                 `json:"digest"`
}

type EvidenceState struct {
	Digest string
	Status string
}
