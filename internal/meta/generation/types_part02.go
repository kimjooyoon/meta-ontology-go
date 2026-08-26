package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

type Binding struct {
	Operation            sourcepolicy.Operation `json:"operation"`
	IndependenceGroupID  string                 `json:"independence_group_id"`
	ProofChoice          ProofChoice            `json:"proof_choice"`
	Executor             string                 `json:"executor"`
	Evaluator            string                 `json:"evaluator"`
	RequiredIndicatorIDs []string               `json:"required_indicator_ids"`
	ReceiptRequired      bool                   `json:"receipt_required"`
	Priority             uint32                 `json:"priority"`
}

type Action struct {
	IndicatorID          string                           `json:"indicator_id"`
	MetricID             sourcepolicy.Dimension           `json:"metric_id"`
	Subject              string                           `json:"subject"`
	SubjectKind          sourcepolicy.SubjectKind         `json:"subject_kind"`
	Applicability        sourcepolicy.Applicability       `json:"applicability"`
	ApplicabilityRule    string                           `json:"applicability_rule_id"`
	ApplicabilityReason  sourcepolicy.ApplicabilityReason `json:"applicability_reason"`
	Blocking             bool                             `json:"blocking"`
	SourceIndicator      sourcepolicy.Indicator           `json:"source_indicator"`
	IndicatorOutcome     sourcepolicy.IndicatorOutcome    `json:"indicator_outcome"`
	MetricProofChoice    sourcepolicy.ProofChoice         `json:"metric_proof_choice"`
	MetricProducer       string                           `json:"metric_producer"`
	MetricConsumer       string                           `json:"metric_consumer"`
	Operation            sourcepolicy.Operation           `json:"meta_operation"`
	IndependenceGroupID  string                           `json:"independence_group_id"`
	ProofChoice          ProofChoice                      `json:"proof_choice"`
	Executor             string                           `json:"executor"`
	Evaluator            string                           `json:"evaluator"`
	RequiredIndicatorIDs []string                         `json:"required_indicator_ids"`
	ReceiptRequired      bool                             `json:"receipt_required"`
	Priority             uint32                           `json:"priority"`
}

type Plan struct {
	SchemaVersion             string                   `json:"schema_version"`
	BaseSHA                   string                   `json:"base_sha"`
	HeadSHA                   string                   `json:"head_sha"`
	PolicyDigest              string                   `json:"policy_digest"`
	RegistryDigest            string                   `json:"registry_digest"`
	IndicatorsDigest          string                   `json:"indicators_digest"`
	IndicatorDecisionLedger   *IndicatorDecisionLedger `json:"indicator_decision_ledger,omitempty"`
	FloorDigest               string                   `json:"floor_digest"`
	InputDigest               string                   `json:"input_digest"`
	RequestedK                uint32                   `json:"requested_k"`
	MinimumIndependent        uint32                   `json:"minimum_independent"`
	ReplayProof               ProofChoice              `json:"replay_proof"`
	Decision                  Decision                 `json:"decision"`
	Reason                    Reason                   `json:"reason"`
	Registry                  []Binding                `json:"registry"`
	Selected                  []Action                 `json:"selected"`
	NotApplicableIndicatorIDs []string                 `json:"not_applicable_indicator_ids"`
	UnselectedIndicatorIDs    []string                 `json:"unselected_indicator_ids"`
	UnknownIndicatorIDs       []string                 `json:"unknown_indicator_ids"`
	Shortfall                 []string                 `json:"shortfall"`
	PromotionAuthorized       bool                     `json:"promotion_authorized"`
	PlanDigest                string                   `json:"plan_digest"`
	ReplayDigest              string                   `json:"replay_digest"`
}

// PromotionAuthorizedByPlan is always false: CI remains the authority.
func (Plan) PromotionAuthorizedByPlan() bool { return false }
