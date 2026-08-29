package transformationeffectverification

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type Options struct {
	PlanPath       string
	ExecutionPath  string
	LedgerPath     string
	ReceiptsPath   string
	ProvenancePath string
	PatchPath      string
	RuntimePath    string
	ExpectedHead   string
	OutputPath     string
	Counterexample string
}

type Runtime struct {
	First                       RuntimePhase `json:"first"`
	Replay                      RuntimePhase `json:"replay"`
	ReplayEqual                 bool         `json:"replay_equal"`
	RepositoryWrites            int          `json:"repository_writes"`
	LocalTestExecutions         int          `json:"local_test_executions"`
	CrossProjectRequiredGates   int          `json:"cross_project_required_gates"`
}

type RuntimePhase struct {
	WallMS      int `json:"wall_ms"`
	PeakRSSKiB  int `json:"peak_rss_kib"`
}

type Counterexample struct {
	ID         string `json:"id"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
	FieldPath  string `json:"field_path"`
	Expected   string `json:"expected"`
	Observed   string `json:"observed"`
}

type Report struct {
	Schema                    string          `json:"schema"`
	Decision                  string          `json:"decision"`
	Resolution                string          `json:"resolution"`
	Stage                     string          `json:"stage"`
	Step                      string          `json:"step"`
	Reason                    string          `json:"reason"`
	FieldPath                 string          `json:"field_path,omitempty"`
	Expected                  string          `json:"expected,omitempty"`
	Observed                  string          `json:"observed,omitempty"`
	UnknownClass              string          `json:"unknown_class,omitempty"`
	NextOperation             string          `json:"next_operation"`
	BlockedBy                 []string        `json:"blocked_by"`
	SelectedPlanOperations    int             `json:"selected_plan_operations"`
	BoundExecutorOperations   int             `json:"bound_executor_operations"`
	UnboundExecutorOperations int             `json:"unbound_executor_operations"`
	ReceiptCount              int             `json:"receipt_count"`
	FailureCount              int             `json:"failure_count"`
	PhysicalCommands          int             `json:"physical_commands"`
	PhysicalTests             int             `json:"physical_tests"`
	ReusedCommands            int             `json:"reused_commands"`
	ReusedTests               int             `json:"reused_tests"`
	RepositoryWrites          int             `json:"repository_writes"`
	LocalTestExecutions       int             `json:"local_test_executions"`
	CrossProjectRequiredGates int             `json:"cross_project_required_gates"`
	Improvement               string          `json:"improvement"`
	OperationOutcome          string          `json:"operation_outcome"`
	PromotionAuthorized       bool            `json:"promotion_authorized"`
	Runtime                   Runtime         `json:"runtime"`
	Counterexamples           []Counterexample `json:"counterexamples,omitempty"`
}

type effect struct {
	ActionIndicatorID string `json:"action_indicator_id"`
	MetricID          string `json:"metric_id"`
	Subject           string `json:"subject"`
	SubjectKind       string `json:"subject_kind"`
	Operation         string `json:"operation"`
	Activity          string `json:"activity"`
	Output            string `json:"output"`
	Executor          string `json:"executor"`
	Evaluator         string `json:"evaluator"`
	ProofChoice       string `json:"proof_choice"`
	BeforeTreeDigest  string `json:"before_tree_digest"`
	AfterTreeDigest   string `json:"after_tree_digest"`
	ChangedPathCount  int    `json:"changed_path_count"`
	ChangedPathDigest string `json:"changed_path_digest"`
	ResidualActionable int   `json:"residual_actionable_count"`
	EvaluatorEvidence string `json:"evaluator_evidence_digest"`
	ReceiptDigest     string `json:"receipt_digest"`
	Status            string `json:"status"`
	SplitGoEvaluation json.RawMessage `json:"split_go_evaluation,omitempty"`
}

type ledger struct {
	Schema                         string            `json:"schema"`
	Metaprogram                    string            `json:"metaprogram"`
	BaseSHA                        string            `json:"base_sha"`
	HeadSHA                        string            `json:"head_sha"`
	SourceSchema                   string            `json:"source_schema"`
	RootTopologyExempt             bool              `json:"root_topology_exempt"`
	Artifacts                      json.RawMessage   `json:"artifacts"`
	InputDigest                    string            `json:"input_digest"`
	IndicatorLedgerDigest          string            `json:"indicator_ledger_digest"`
	IndicatorLedgerCount           int               `json:"indicator_ledger_count"`
	Decision                       string            `json:"decision"`
	Reason                         string            `json:"reason"`
	WorkspaceMode                  string            `json:"workspace_mode"`
	WriteBoundary                  string            `json:"write_boundary"`
	SourceTreeBefore               string            `json:"source_tree_before"`
	SourceTreeAfter                string            `json:"source_tree_after"`
	SourceWorkspaceUnchanged       bool              `json:"source_workspace_unchanged"`
	SandboxTreeBefore              string            `json:"sandbox_tree_before"`
	SandboxTreeAfter               string            `json:"sandbox_tree_after"`
	Effects                        []effect          `json:"effects"`
	EffectDigest                   string            `json:"effect_digest"`
	PatchDigest                    string            `json:"patch_digest"`
	InputReceiptReportDigest       string            `json:"input_receipt_report_digest"`
	GeneratedReceiptReportDigest string `json:"generated_receipt_report_digest"`
	InputProvenanceDigest          string            `json:"input_provenance_digest"`
	ExecutedProvenanceDigest       string            `json:"executed_provenance_digest"`
	SelectedPlanOperations         int               `json:"selected_plan_operations"`
	BoundExecutorOperations        int               `json:"bound_executor_operations"`
	UnboundExecutorOperations      int               `json:"unbound_executor_operations"`
	Status                         string            `json:"status"`
	Indicators                     []json.RawMessage `json:"indicators"`
	SemanticDigest                 string            `json:"semantic_digest"`
	LedgerDigest                   string            `json:"ledger_digest"`
	ReplayDigest                   string            `json:"replay_digest"`
	PromotionAuthorized            bool              `json:"promotion_authorized"`
}

type patch struct {
	Schema       string            `json:"schema"`
	HeadSHA      string            `json:"head_sha"`
	Changes      []json.RawMessage `json:"changes"`
	ChangeDigest string            `json:"change_digest"`
	PatchDigest  string            `json:"patch_digest"`
}

type provenance struct {
	Schema                        string            `json:"schema_version"`
	BaseSHA                       string            `json:"base_sha"`
	HeadSHA                       string            `json:"head_sha"`
	PlanDigest                    string            `json:"plan_digest"`
	ExecutionManifestDigest       string            `json:"execution_manifest_digest"`
	ReceiptReportDigest            string            `json:"receipt_report_digest"`
	IndicatorDecisionLedgerDigest string            `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int               `json:"indicator_decision_ledger_count"`
	InputDigest                   string            `json:"input_digest"`
	Decision                      string            `json:"decision"`
	Reason                        string            `json:"reason"`
	Indicators                    []json.RawMessage `json:"indicators"`
	Summary                       json.RawMessage   `json:"summary"`
	PromotionAuthorized           bool              `json:"promotion_authorized"`
	EnvelopeDigest                string            `json:"envelope_digest"`
	ReplayDigest                  string            `json:"replay_digest"`
}

type bundle struct {
	Plan       generation.Plan
	Execution  generation.ExecutionManifest
	Receipts   generation.ReceiptReport
	Ledger     ledger
	Provenance provenance
	Patch      patch
	Runtime    Runtime
}

type validationFailure struct {
	Decision   string
	Resolution string
	Stage      string
	Step       string
	Reason     string
	Unknown    string
	Next       string
	Blocked    []string
	FieldPath  string
	Expected   string
	Observed   string
}

func (f *validationFailure) Error() string {
	return f.Stage + "/" + f.Step + "/" + f.Reason
}
