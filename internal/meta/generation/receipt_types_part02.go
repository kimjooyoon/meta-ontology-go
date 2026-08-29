package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

type IndicatorReceipt struct {
	ID             string           `json:"id"`
	Verdict        IndicatorVerdict `json:"verdict"`
	EvidenceDigest string           `json:"evidence_digest"`
	ProofChoice    ProofChoice      `json:"proof_choice"`
	Observation    *IndicatorObservation `json:"observation,omitempty"`
}

const OperationInstanceEvidenceSchema = "gooo/meta-operation-instance-evidence/v1"

type ProcessObservation struct {
	Command       []string `json:"command"`
	ExitCode      int      `json:"exit_code"`
	StdoutBytes   int      `json:"stdout_bytes"`
	StdoutDigest  string   `json:"stdout_digest"`
	StderrBytes   int      `json:"stderr_bytes"`
	StderrDigest  string   `json:"stderr_digest"`
}

type OperationInstanceEvidence struct {
	Schema                  string             `json:"schema"`
	ActionIndicatorID       string             `json:"action_indicator_id"`
	Subject                 string             `json:"subject"`
	HeadSHA                 string             `json:"head_sha"`
	OperationID             string             `json:"operation_id"`
	ContractEvidenceDigest  string             `json:"contract_evidence_digest"`
	InstanceEvidenceDigest  string             `json:"instance_evidence_digest"`
	ExecutorObservation    ProcessObservation `json:"executor_observation"`
	EvaluatorObservation   ProcessObservation `json:"evaluator_observation"`
	ReplayComparisons      int                `json:"replay_comparisons"`
	ReplayMatch             bool               `json:"replay_match"`
	VerifierObservation    *ProcessObservation `json:"verifier_observation,omitempty"`
}

type ObservationFailure struct {
	ActionIndicatorID string             `json:"action_indicator_id"`
	Decision          string             `json:"decision"`
	Stage             string             `json:"stage"`
	Step              string             `json:"step"`
	Reason            string             `json:"reason"`
	UnknownClass      string             `json:"unknown_class,omitempty"`
	NextOperation     string             `json:"next_operation"`
	BlockedBy         []string           `json:"blocked_by"`
	Executor          ProcessObservation `json:"executor_observation"`
}

type OperationReceipt struct {
	SchemaVersion                 string                 `json:"schema_version"`
	BaseSHA                       string                 `json:"base_sha"`
	HeadSHA                       string                 `json:"head_sha"`
	PlanDigest                    string                 `json:"plan_digest"`
	IndicatorDecisionLedgerDigest string                 `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int                    `json:"indicator_decision_ledger_count"`
	ActionIndicatorID             string                 `json:"action_indicator_id"`
	Operation                     sourcepolicy.Operation `json:"operation"`
	Activity                      string                 `json:"activity"`
	Output                        string                 `json:"output"`
	Executor                      string                 `json:"executor"`
	Evaluator                     string                 `json:"evaluator"`
	ProofChoice                   ProofChoice            `json:"proof_choice"`
	Indicators                    []IndicatorReceipt     `json:"indicators"`
	InstanceEvidence              *OperationInstanceEvidence `json:"instance_evidence,omitempty"`
	ReceiptDigest                 string                 `json:"receipt_digest"`
}

// ReceiptUnknown preserves an unavailable or malformed execution observation
// together with the operation contract it was meant to discharge.
type ReceiptUnknown struct {
	ActionIndicatorID   string                 `json:"action_indicator_id"`
	RequiredIndicatorID string                 `json:"required_indicator_id"`
	Operation           sourcepolicy.Operation `json:"operation"`
	Activity            string                 `json:"activity"`
	Output              string                 `json:"output"`
	Executor            string                 `json:"executor"`
	Evaluator           string                 `json:"evaluator"`
	Stage               string                 `json:"stage"`
	Step                string                 `json:"step"`
	Reason              ReceiptReason          `json:"reason"`
	UnknownClass        string                 `json:"unknown_class"`
	NextOperation       string                 `json:"next_operation"`
	BlockedBy           []string               `json:"blocked_by"`
}

type ReceiptReport struct {
	SchemaVersion                 string             `json:"schema_version"`
	BaseSHA                       string             `json:"base_sha"`
	HeadSHA                       string             `json:"head_sha"`
	PlanDigest                    string             `json:"plan_digest"`
	IndicatorDecisionLedgerDigest string             `json:"indicator_decision_ledger_digest,omitempty"`
	IndicatorDecisionLedgerCount  int                `json:"indicator_decision_ledger_count"`
	InputDigest                   string             `json:"input_digest"`
	Decision                      ReceiptDecision    `json:"decision"`
	Reason                        ReceiptReason      `json:"reason"`
	Receipts                      []OperationReceipt `json:"receipts"`
	MissingIndicatorIDs           []string           `json:"missing_indicator_ids"`
	UnknownIndicatorIDs           []string           `json:"unknown_indicator_ids"`
	RejectedIndicatorIDs          []string           `json:"rejected_indicator_ids"`
	Unknowns                      []ReceiptUnknown   `json:"unknowns"`
	PromotionAuthorized           bool               `json:"promotion_authorized"`
	ReportDigest                  string             `json:"report_digest"`
	ReplayDigest                  string             `json:"replay_digest"`
}

// PromotionAuthorizedByReceipts is always false: CI remains the authority.
func (ReceiptReport) PromotionAuthorizedByReceipts() bool { return false }
