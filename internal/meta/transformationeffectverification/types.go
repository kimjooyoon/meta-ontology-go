package transformationeffectverification

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"

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
	Operation         string `json:"operation"`
	Activity          string `json:"activity"`
	Output            string `json:"output"`
	Executor          string `json:"executor"`
	Evaluator         string `json:"evaluator"`
	ProofChoice       string `json:"proof_choice"`
	Status            string `json:"status"`
}

type ledger struct {
	Schema                    string   `json:"schema"`
	BaseSHA                   string   `json:"base_sha"`
	HeadSHA                   string   `json:"head_sha"`
	Effects                   []effect `json:"effects"`
	SelectedPlanOperations    int      `json:"selected_plan_operations"`
	BoundExecutorOperations   int      `json:"bound_executor_operations"`
	UnboundExecutorOperations int      `json:"unbound_executor_operations"`
	GeneratedReceiptReportDigest string `json:"generated_receipt_report_digest"`
	ExecutedProvenanceDigest  string   `json:"executed_provenance_digest"`
	PatchDigest               string   `json:"patch_digest"`
}

type patch struct {
	Schema      string `json:"schema"`
	HeadSHA     string `json:"head_sha"`
	PatchDigest string `json:"patch_digest"`
}

type provenance struct {
	Schema             string `json:"schema_version"`
	HeadSHA            string `json:"head_sha"`
	ReceiptReportDigest string `json:"receipt_report_digest"`
	EnvelopeDigest     string `json:"envelope_digest"`
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
