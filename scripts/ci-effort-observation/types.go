package main

const (
	reportSchema   = "gooo/ci-effort-observation/v1"
	contractSchema = "gooo/ci-effort-observation-contract/v1"
	manifestSchema = "gooo/ci-effort-operation-manifest/v1"
	goToolchain    = "go1.27.0"
)

type Config struct {
	RunPath              string
	JobsPath             string
	ManifestPath         string
	ContractPath         string
	ProgramPath          string
	TimeCausalityRoot    string
	SummaryPath          string
	EvidencePath         string
	RepositoryStatusPath string
	OpenTofuPath         string
	OpenTofuMetaPath     string
	PriorPath            string
	DependencyFiles      []string
	Environment          string
	OutputPath              string
	MarkdownPath            string
	ReadOnly                bool
	LineageObservationPath  string
	ReadOnlyObservationPath string
	Check                   bool
}

type SourceRun struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	WorkflowName   string `json:"workflow_name"`
	WorkflowPath   string `json:"path"`
	Event          string `json:"event"`
	Ref            string `json:"ref"`
	HeadBranch     string `json:"head_branch"`
	HeadSHA        string `json:"head_sha"`
	HeadRepository string `json:"head_repository"`
	Status         string `json:"status"`
	Conclusion     string `json:"conclusion"`
	RunAttempt     int64  `json:"run_attempt"`
	CreatedAt      string `json:"created_at"`
	RunStartedAt   string `json:"run_started_at"`
	UpdatedAt      string `json:"updated_at"`
	HTMLURL        string `json:"html_url"`
}

type JobsInput struct {
	Jobs []APIJob `json:"jobs"`
}

type APIJob struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	HeadSHA     string    `json:"head_sha"`
	RunID       int64     `json:"run_id"`
	StartedAt   string    `json:"started_at"`
	CompletedAt string    `json:"completed_at"`
	HTMLURL     string    `json:"html_url"`
	Steps       []APIStep `json:"steps"`
}

type APIStep struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

type Manifest struct {
	Schema         string          `json:"schema"`
	Workflow       string          `json:"workflow"`
	WorkflowSource string          `json:"workflow_source"`
	Operations     []OperationSpec `json:"operations"`
	ExcludedJobs   []ExcludedJob   `json:"excluded_jobs"`
}

type OperationSpec struct {
	ID                string            `json:"id"`
	JobName           string            `json:"job_name"`
	StepName          string            `json:"step_name"`
	EvidenceStepName  string            `json:"evidence_step,omitempty"`
	GuardStepName     string            `json:"guard_step,omitempty"`
	EventStepNames    map[string]string `json:"event_step_names,omitempty"`
	Kind              string            `json:"kind"`
	Command           []string          `json:"command"`
	ProofObligationID string            `json:"proof_obligation_id"`
}

type ExcludedJob struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type Contract struct {
	Schema       string     `json:"schema"`
	ID           string     `json:"id"`
	Cells        []CellSpec `json:"cells"`
	GraphProgram string     `json:"graph_program"`
	NotClaimed   []string   `json:"not_claimed"`
}

type CellSpec struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Indicator     string `json:"indicator"`
	Activity      string `json:"activity"`
	InputID       string `json:"input_id"`
	OutputID      string `json:"output_id"`
}

type JobObservation struct {
	ID                    int64             `json:"id"`
	OperationID           string            `json:"operation_id"`
	RunID                 int64             `json:"run_id"`
	Provider              string            `json:"provider"`
	ClockDomain           string            `json:"clock_domain"`
	Name                  string            `json:"name"`
	Status                string            `json:"status"`
	Conclusion            string            `json:"conclusion"`
	HeadSHA               string            `json:"head_sha"`
	StartedAt             string            `json:"started_at"`
	CompletedAt           string            `json:"completed_at"`
	WallMS                int64             `json:"wall_ms"`
	BelowSourceResolution bool              `json:"below_source_resolution,omitempty"`
	RejectionReason       string            `json:"rejection_reason,omitempty"`
	Skipped               bool              `json:"skipped,omitempty"`
	Steps                 []StepObservation `json:"steps"`
	Unknown               *Unknown          `json:"unknown,omitempty"`
}

type StepObservation struct {
	OperationID           string   `json:"operation_id"`
	RunID                 int64    `json:"run_id"`
	Provider              string   `json:"provider"`
	ClockDomain           string   `json:"clock_domain"`
	Name                  string   `json:"name"`
	Status                string   `json:"status"`
	Conclusion            string   `json:"conclusion"`
	StartedAt             string   `json:"started_at"`
	CompletedAt           string   `json:"completed_at"`
	WallMS                int64    `json:"wall_ms"`
	BelowSourceResolution bool     `json:"below_source_resolution,omitempty"`
	RejectionReason       string   `json:"rejection_reason,omitempty"`
	Skipped               bool     `json:"skipped,omitempty"`
	Unknown               *Unknown `json:"unknown,omitempty"`
}

type WorkflowWindow struct {
	OperationID                string   `json:"operation_id"`
	RunID                      int64    `json:"run_id"`
	Provider                   string   `json:"provider"`
	ClockDomain                string   `json:"clock_domain"`
	StartAt                    string   `json:"start_at"`
	EndAt                      string   `json:"end_at"`
	WallMS                     int64    `json:"wall_ms"`
	JobWallMSSum               int64    `json:"job_wall_ms_sum"`
	StepWallMSSum              int64    `json:"step_wall_ms_sum"`
	TimestampResolutionMS      int64    `json:"timestamp_resolution_ms"`
	IntervalModel              string   `json:"interval_model"`
	IntervalModelDigest        string   `json:"interval_model_digest"`
	BelowSourceResolutionJobs  int      `json:"below_source_resolution_jobs"`
	BelowSourceResolutionSteps int      `json:"below_source_resolution_steps"`
	JobIntervalCount           int      `json:"job_interval_count"`
	StepIntervalCount          int      `json:"step_interval_count"`
	JobWallMSNominal           int64    `json:"job_wall_ms_nominal"`
	StepWallMSNominal          int64    `json:"step_wall_ms_nominal"`
	RuntimeRejectionCount      int      `json:"runtime_rejection_count"`
	RuntimeRejectionReasons    []string `json:"runtime_rejection_reasons"`
}

type RuntimeCase struct {
	ID         string `json:"id"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type OperationObservation struct {
	ID                     string   `json:"id"`
	OperationID            string   `json:"operation_id"`
	RunID                  int64    `json:"run_id"`
	Provider               string   `json:"provider"`
	ClockDomain            string   `json:"clock_domain"`
	Kind                   string   `json:"kind"`
	ProofObligationID      string   `json:"proof_obligation_id"`
	SourceEvent            string   `json:"source_event"`
	JobName                string   `json:"job_name"`
	StepName               string   `json:"step_name"`
	BoundStepName          string   `json:"bound_step_name"`
	DeclaredStepCandidates []string `json:"declared_step_candidates"`
	EvidenceStepName       string   `json:"evidence_step"`
	GuardStepName          string   `json:"guard_step,omitempty"`
	Command                []string `json:"command"`
	WorkflowSourcePath     string   `json:"workflow_source_path"`
	WorkflowSourceDigest   string   `json:"workflow_source_digest"`
	CommandContextDigest   string   `json:"command_context_digest"`
	CommandBound           bool     `json:"command_bound"`
	JobID                  int64    `json:"job_id"`
	JobConclusion          string   `json:"job_conclusion"`
	StepStatus             string   `json:"step_status"`
	StepConclusion         string   `json:"step_conclusion"`
	GuardStepStatus        string   `json:"guard_step_status,omitempty"`
	GuardStepConclusion    string   `json:"guard_step_conclusion,omitempty"`
	GuardBound             bool     `json:"guard_bound,omitempty"`
	State                  string   `json:"state"`
	RejectionReason        string   `json:"rejection_reason,omitempty"`
	Skipped                bool     `json:"skipped,omitempty"`
	StartedAt              string   `json:"started_at"`
	CompletedAt            string   `json:"completed_at"`
	WallMS                 int64    `json:"wall_ms"`
	BelowSourceResolution  bool     `json:"below_source_resolution,omitempty"`
	Unknown                *Unknown `json:"unknown,omitempty"`
	EvidenceDigest         string   `json:"evidence_digest"`
}

type Accounting struct {
	ManifestOperations int `json:"manifest_operations"`
	Executed           int `json:"executed"`
	Skipped            int `json:"skipped"`
	Unknown            int `json:"unknown"`
	Rejected           int `json:"rejected"`
	ExecutedCommands   int `json:"executed_commands"`
	ExecutedTests      int `json:"executed_tests"`
	SkippedCommands    int `json:"skipped_commands"`
	SkippedTests       int `json:"skipped_tests"`
}

type RepositoryStatus struct {
	Before string `json:"before"`
	After  string `json:"after"`
	Writes int    `json:"writes"`
}

type ReuseKey struct {
	HeadSHA                    string            `json:"head_sha"`
	SourceEvent                string            `json:"source_event"`
	InputDigest                string            `json:"input_digest"`
	ToolchainDigest            string            `json:"toolchain_digest"`
	CommandContextDigest       string            `json:"command_context_digest"`
	EnvironmentAllowlistDigest string            `json:"environment_allowlist_digest"`
	DependencyGraphDigest      string            `json:"dependency_graph_digest"`
	DependencyInputs           []DependencyInput `json:"dependency_inputs"`
	ExpectedResultDigest       string            `json:"expected_result_digest"`
	OpenTofuReleaseDigest      string            `json:"opentofu_release_digest"`
	TimeCausalityDigest        string            `json:"time_causality_digest"`
}

type DependencyInput struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	Digest string `json:"digest"`
}

type ReuseObservation struct {
	Decision            string      `json:"decision"`
	Resolution          string      `json:"resolution"`
	Reason              string      `json:"reason"`
	Requests            int         `json:"requests"`
	PriorCandidates     int         `json:"prior_candidates"`
	PriorReceiptsValid  int         `json:"prior_receipts_valid"`
	Reused              int         `json:"reused"`
	Rejected            int         `json:"rejected"`
	Unknown             int         `json:"unknown"`
	Skipped             int         `json:"skipped"`
	ReusedCommands      int         `json:"reused_commands"`
	ReusedTests         int         `json:"reused_tests"`
	RequiresExecution   bool        `json:"requires_execution"`
	NextOperation       string      `json:"next_operation,omitempty"`
	UnknownEvidence     *Unknown    `json:"unknown_evidence,omitempty"`
	Key                 ReuseKey    `json:"key"`
	PriorReceiptDigest  string      `json:"prior_receipt_digest"`
	PriorEvidenceDigest string      `json:"prior_evidence_digest"`
	Cases               []ReuseCase `json:"cases"`
}

type PriorRecord struct {
	Schema           string   `json:"schema"`
	Decision         string   `json:"decision"`
	Resolution       string   `json:"resolution"`
	HeadSHA          string   `json:"head_sha"`
	Key              ReuseKey `json:"key"`
	EvidenceDigest   string   `json:"evidence_digest"`
	ResultDigest     string   `json:"result_digest"`
	RepositoryWrites int      `json:"repository_writes"`
}

type ReuseCase struct {
	ID         string   `json:"id"`
	Decision   string   `json:"decision"`
	Resolution string   `json:"resolution"`
	Reason     string   `json:"reason"`
	Unknown    *Unknown `json:"unknown,omitempty"`
}

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Counterexample struct {
	ID         string   `json:"id"`
	Decision   string   `json:"decision"`
	Resolution string   `json:"resolution"`
	Reason     string   `json:"reason"`
	Unknown    *Unknown `json:"unknown,omitempty"`
}

type CellObservation struct {
	ID             string `json:"id"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Indicator      string `json:"indicator"`
	Activity       string `json:"activity"`
	InputID        string `json:"input_id"`
	OutputID       string `json:"output_id"`
	Observed       string `json:"observed"`
	Expected       string `json:"expected"`
	Decision       string `json:"decision"`
	EvidenceDigest string `json:"evidence_digest"`
}

type ExternalOpenTofu struct {
	Workflow           string   `json:"workflow"`
	RunID              int64    `json:"run_id"`
	ArtifactID         int64    `json:"artifact_id"`
	ArtifactName       string   `json:"artifact_name"`
	ArtifactDigest     string   `json:"artifact_digest"`
	ArtifactSize       int64    `json:"artifact_size"`
	ReportDigest       string   `json:"report_digest"`
	ReleaseAssetDigest string   `json:"release_asset_digest"`
	ReportSchema       string   `json:"report_schema"`
	SubjectSHA         string   `json:"subject_sha"`
	Decision           string   `json:"decision"`
	Resolution         string   `json:"resolution"`
	CellsClosed        int      `json:"cells_closed"`
	CellsTotal         int      `json:"cells_total"`
	ReuseDecision      string   `json:"reuse_decision"`
	ReuseCount         int      `json:"reuse_count"`
	Unknown            *Unknown `json:"unknown,omitempty"`
}

type ArtifactMeta struct {
	RunID   int64  `json:"run_id"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Size    int64  `json:"size_in_bytes"`
	Expired bool   `json:"expired"`
}

type OpenTofuReportInput struct {
	Schema       string `json:"schema"`
	SubjectSHA   string `json:"subject_sha"`
	Decision     string `json:"decision"`
	Resolution   string `json:"resolution"`
	ReportDigest string `json:"report_digest"`
	Release      struct {
		AssetSHA256 string `json:"asset_sha256"`
	} `json:"release"`
	Summary struct {
		CellsTotal       int `json:"cells_total"`
		ClosedCells      int `json:"closed_cells"`
		ReplayMatches    int `json:"replay_matches"`
		RepositoryWrites int `json:"repository_writes"`
		LocalTests       int `json:"local_test_executions"`
	} `json:"summary"`
	Reuse struct {
		Decision string `json:"decision"`
		Reused   int    `json:"reused"`
	} `json:"reuse"`
}

type GraphObservation struct {
	ProgramPath   string   `json:"program_path"`
	ProgramDigest string   `json:"program_digest"`
	ActivityCount int      `json:"activity_count"`
	BindingCount  int      `json:"binding_count"`
	Activities    []string `json:"activities"`
}

type Report struct {
	Schema                    string                 `json:"schema"`
	ContractID                string                 `json:"contract_id"`
	Repository                string                 `json:"repository"`
	SourceWorkflow            string                 `json:"source_workflow"`
	SourceEvent               string                 `json:"source_event"`
	SourceRef                 string                 `json:"source_ref"`
	HeadSHA                   string                 `json:"head_sha"`
	SourceRunConclusion       string                 `json:"source_run_conclusion"`
	SourceRunID               int64                  `json:"source_run_id"`
	SourceRunAttempt          int64                  `json:"source_run_attempt"`
	SourceRunURL              string                 `json:"source_run_url"`
	WorkflowSourcePath        string                 `json:"workflow_source_path"`
	WorkflowSourceDigest      string                 `json:"workflow_source_digest"`
	Window                    WorkflowWindow         `json:"workflow_window"`
	RuntimeResolution         string                 `json:"runtime_resolution"`
	RuntimeCases              []RuntimeCase          `json:"runtime_cases"`
	Jobs                      []JobObservation       `json:"jobs"`
	Operations                []OperationObservation `json:"operations"`
	Accounting                Accounting             `json:"accounting"`
	OperationManifestDigest   string                 `json:"operation_manifest_digest"`
	Reuse                     ReuseObservation       `json:"reuse"`
	OpenTofu                  ExternalOpenTofu       `json:"external_opentofu"`
	Cells                     []CellObservation      `json:"cells"`
	Graph                     GraphObservation       `json:"graph"`
	TimeCausality             TimeCausalityBinding   `json:"time_causality"`
	RepositoryStatus          RepositoryStatus       `json:"repository_status"`
	RepositoryWrites          int                    `json:"repository_writes"`
	LocalTestExecutions       int                    `json:"local_test_executions"`
	CrossProjectRequiredGates int                    `json:"cross_project_required_gates"`
	Improvement               string                 `json:"improvement"`
	Decision                  string                 `json:"decision"`
	Resolution                string                 `json:"resolution"`
	Reason                    string                 `json:"reason"`
	UnknownEvidence           *Unknown               `json:"unknown_evidence,omitempty"`
	Counterexamples           []Counterexample       `json:"counterexamples"`
	HumanReport               string                 `json:"human_report"`
	ReportDigest              string                 `json:"report_digest"`
}
