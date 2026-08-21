package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

type IndicatorReceipt struct {
	ID             string           `json:"id"`
	Verdict        IndicatorVerdict `json:"verdict"`
	EvidenceDigest string           `json:"evidence_digest"`
	ProofChoice    ProofChoice      `json:"proof_choice"`
}

type OperationReceipt struct {
	SchemaVersion     string                 `json:"schema_version"`
	BaseSHA           string                 `json:"base_sha"`
	HeadSHA           string                 `json:"head_sha"`
	PlanDigest        string                 `json:"plan_digest"`
	ActionIndicatorID string                 `json:"action_indicator_id"`
	Operation         sourcepolicy.Operation `json:"operation"`
	Evaluator         string                 `json:"evaluator"`
	ProofChoice       ProofChoice            `json:"proof_choice"`
	Indicators        []IndicatorReceipt     `json:"indicators"`
	ReceiptDigest     string                 `json:"receipt_digest"`
}

type ReceiptReport struct {
	SchemaVersion          string             `json:"schema_version"`
	BaseSHA                string             `json:"base_sha"`
	HeadSHA                string             `json:"head_sha"`
	PlanDigest             string             `json:"plan_digest"`
	InputDigest            string             `json:"input_digest"`
	Decision               ReceiptDecision    `json:"decision"`
	Reason                 ReceiptReason      `json:"reason"`
	Receipts               []OperationReceipt `json:"receipts"`
	MissingIndicatorIDs    []string           `json:"missing_indicator_ids"`
	UnknownIndicatorIDs    []string           `json:"unknown_indicator_ids"`
	RejectedIndicatorIDs   []string           `json:"rejected_indicator_ids"`
	PromotionAuthorized    bool               `json:"promotion_authorized"`
	ReportDigest           string             `json:"report_digest"`
	ReplayDigest           string             `json:"replay_digest"`
}

// PromotionAuthorizedByReceipts is always false: CI remains the authority.
func (ReceiptReport) PromotionAuthorizedByReceipts() bool { return false }
