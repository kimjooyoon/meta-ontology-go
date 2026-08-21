package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

type ExecutionStep struct {
	ActionIndicatorID    string                           `json:"action_indicator_id"`
	MetricID             sourcepolicy.Dimension           `json:"metric_id"`
	Subject              string                           `json:"subject"`
	SubjectKind          sourcepolicy.SubjectKind         `json:"subject_kind"`
	Applicability        sourcepolicy.Applicability       `json:"applicability"`
	ApplicabilityRule    string                           `json:"applicability_rule_id"`
	ApplicabilityReason  sourcepolicy.ApplicabilityReason `json:"applicability_reason"`
	MetricProofChoice    sourcepolicy.ProofChoice         `json:"metric_proof_choice"`
	MetricProducer       string                           `json:"metric_producer"`
	MetricConsumer       string                           `json:"metric_consumer"`
	Operation            sourcepolicy.Operation           `json:"operation"`
	IndependenceGroupID  string                           `json:"independence_group_id"`
	ProofChoice          ProofChoice                      `json:"proof_choice"`
	Executor             string                           `json:"executor"`
	Evaluator            string                           `json:"evaluator"`
	RequiredIndicatorIDs []string                         `json:"required_indicator_ids"`
	ReceiptRequired      bool                             `json:"receipt_required"`
	Priority             uint32                           `json:"priority"`
	WorkspaceMode        WorkspaceMode                    `json:"workspace_mode"`
	WriteBoundary        WriteBoundary                    `json:"write_boundary"`
}

type ExecutionManifest struct {
	SchemaVersion             string            `json:"schema_version"`
	BaseSHA                   string            `json:"base_sha"`
	HeadSHA                   string            `json:"head_sha"`
	PlanDigest                string            `json:"plan_digest"`
	InputDigest               string            `json:"input_digest"`
	Decision                  ExecutionDecision `json:"decision"`
	Reason                    ExecutionReason   `json:"reason"`
	Steps                     []ExecutionStep   `json:"steps"`
	NotApplicableIndicatorIDs []string          `json:"not_applicable_indicator_ids"`
	PromotionAuthorized       bool              `json:"promotion_authorized"`
	ManifestDigest            string            `json:"manifest_digest"`
	ReplayDigest              string            `json:"replay_digest"`
}

// PromotionAuthorizedByExecution is always false: CI remains the authority.
func (ExecutionManifest) PromotionAuthorizedByExecution() bool { return false }
