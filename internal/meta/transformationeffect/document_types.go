package transformationeffect

const ledgerSchema = "gooo/transformation-effect-ledger/v1"
const patchSchema = "gooo/transformation-content-patch/v1"

type Indicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type Effect struct {
	ActionIndicatorID  string                     `json:"action_indicator_id"`
	MetricID           string                     `json:"metric_id"`
	Subject            string                     `json:"subject"`
	SubjectKind        string                     `json:"subject_kind"`
	Operation          string                     `json:"operation"`
	Activity           string                     `json:"activity"`
	Output             string                     `json:"output"`
	Executor           string                     `json:"executor"`
	Evaluator          string                     `json:"evaluator"`
	ProofChoice        string                     `json:"proof_choice"`
	BeforeTreeDigest   string                     `json:"before_tree_digest"`
	AfterTreeDigest    string                     `json:"after_tree_digest"`
	ChangedPathCount   int                        `json:"changed_path_count"`
	ChangedPathDigest  string                     `json:"changed_path_digest"`
	ResidualActionable int                        `json:"residual_actionable_count"`
	EvaluatorEvidence  string                     `json:"evaluator_evidence_digest"`
	ReceiptDigest      string                     `json:"receipt_digest"`
	Status             string                     `json:"status"`
	SplitGoEvaluation  *SplitGoEvaluationArtifact `json:"split_go_evaluation,omitempty"`
}

type Ledger struct {
	Schema                       string          `json:"schema"`
	Metaprogram                  string          `json:"metaprogram"`
	BaseSHA                      string          `json:"base_sha"`
	HeadSHA                      string          `json:"head_sha"`
	SourceSchema                 string          `json:"source_schema"`
	RootTopologyExempt           bool            `json:"root_topology_exempt"`
	Artifacts                    ArtifactDigests `json:"artifacts"`
	InputDigest                  string          `json:"input_digest"`
	IndicatorLedgerDigest        string          `json:"indicator_ledger_digest"`
	IndicatorLedgerCount         int             `json:"indicator_ledger_count"`
	Decision                     string          `json:"decision"`
	Reason                       string          `json:"reason"`
	WorkspaceMode                string          `json:"workspace_mode"`
	WriteBoundary                string          `json:"write_boundary"`
	SourceTreeBefore             string          `json:"source_tree_before"`
	SourceTreeAfter              string          `json:"source_tree_after"`
	SourceWorkspaceUnchanged     bool            `json:"source_workspace_unchanged"`
	SandboxTreeBefore            string          `json:"sandbox_tree_before"`
	SandboxTreeAfter             string          `json:"sandbox_tree_after"`
	Effects                      []Effect        `json:"effects"`
	EffectDigest                 string          `json:"effect_digest"`
	PatchDigest                  string          `json:"patch_digest"`
	InputReceiptReportDigest     string          `json:"input_receipt_report_digest"`
	GeneratedReceiptReportDigest string          `json:"generated_receipt_report_digest"`
	InputProvenanceDigest        string          `json:"input_provenance_digest"`
	ExecutedProvenanceDigest     string          `json:"executed_provenance_digest"`
	SelectedPlanOperations       int             `json:"selected_plan_operations"`
	BoundExecutorOperations      int             `json:"bound_executor_operations"`
	UnboundExecutorOperations    int             `json:"unbound_executor_operations"`
	Status                       string          `json:"status"`
	Indicators                   []Indicator     `json:"indicators"`
	SemanticDigest               string          `json:"semantic_digest"`
	LedgerDigest                 string          `json:"ledger_digest"`
	ReplayDigest                 string          `json:"replay_digest"`
	PromotionAuthorized          bool            `json:"promotion_authorized"`
}
