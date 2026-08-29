package opentofuobservation

type CommandReceipt struct {
	Name         string   `json:"name"`
	Command      []string `json:"command"`
	CwdRole      string   `json:"cwd_role"`
	ExitCode     int      `json:"exit_code"`
	StdoutBytes  int      `json:"stdout_bytes"`
	StdoutDigest string   `json:"stdout_digest"`
	StderrBytes  int      `json:"stderr_bytes"`
	StderrDigest string   `json:"stderr_digest"`
	WallMS       int      `json:"wall_ms"`
	PeakRSSKiB   int      `json:"peak_rss_kib"`
	Executed     bool     `json:"executed"`
}

type ReleaseObservation struct {
	ReleaseID       string         `json:"release_id"`
	AssetURL        string         `json:"asset_url"`
	AssetSHA256     string         `json:"asset_sha256"`
	AssetBytes      int            `json:"asset_bytes"`
	ChecksumsSHA256 string         `json:"checksums_sha256"`
	VersionJSON     any            `json:"version_json"`
	VersionJSONSHA  string         `json:"version_json_digest"`
	Version         string         `json:"version"`
	Platform        string         `json:"platform"`
	Command         CommandReceipt `json:"command"`
}

type ExecutionRun struct {
	Index                   int              `json:"index"`
	FixtureDigest           string           `json:"fixture_digest"`
	PlanJSONDigest          string           `json:"plan_json_digest"`
	PlanRawDigest           string           `json:"plan_raw_digest"`
	PlanCanonicalizer       string           `json:"plan_canonicalizer"`
	PlanCanonicalizerDigest string           `json:"plan_canonicalizer_digest"`
	PlanVolatileFields      []string         `json:"plan_volatile_fields"`
	PlanJSONBytes            int              `json:"plan_json_bytes"`
	PlanSchemaValid          bool             `json:"plan_schema_valid"`
	TestEventDigest          string           `json:"test_event_digest"`
	TestRawDigest            string           `json:"test_raw_digest"`
	TestEventCount           int              `json:"test_event_count"`
	TestTypeCounts           map[string]int   `json:"test_type_counts"`
	TestAbstractDiscovered   int              `json:"test_abstract_discovered"`
	TestRunExecuted          int              `json:"test_run_executed"`
	TestSummaryPassed        int              `json:"test_summary_passed"`
	TestSummaryFailed        int              `json:"test_summary_failed"`
	TestSummaryErrored       int              `json:"test_summary_errored"`
	TestSummarySkipped       int              `json:"test_summary_skipped"`
	TestEventsValid          bool             `json:"test_events_valid"`
	Commands                 []CommandReceipt `json:"commands"`
}

type ReuseAccounting struct {
	Discovered            int    `json:"discovered"`
	Executed              int    `json:"executed"`
	Reused                int    `json:"reused"`
	Skipped               int    `json:"skipped"`
	PriorCandidates       int    `json:"prior_candidates"`
	Invalidated           int    `json:"invalidated"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
	SourceDigest          string `json:"source_digest"`
	FixtureDigest         string `json:"fixture_digest"`
	ArgumentDigest        string `json:"argument_digest"`
	EnvironmentDigest     string `json:"environment_allowlist_digest"`
	ReleaseDigest         string `json:"release_digest"`
	ToolchainDigest       string `json:"observer_toolchain_digest"`
	DependencyGraphDigest string `json:"dependency_graph_digest"`
	ExpectedResultDigest  string `json:"expected_result_digest"`
	PriorReceiptDigest    string `json:"prior_receipt_digest"`
}

type RuntimeSummary struct {
	ConsumerBuildMS       int `json:"consumer_build_ms"`
	ConsumerBuildPeakRSS  int `json:"consumer_build_peak_rss_kib"`
	TofuInitMS             int `json:"tofu_init_ms"`
	TofuInitPeakRSS        int `json:"tofu_init_peak_rss_kib"`
	TofuPlanMS             int `json:"tofu_plan_ms"`
	TofuPlanPeakRSS        int `json:"tofu_plan_peak_rss_kib"`
	TofuShowMS             int `json:"tofu_show_ms"`
	TofuShowPeakRSS        int `json:"tofu_show_peak_rss_kib"`
	TofuTestMS             int `json:"tofu_test_ms"`
	TofuTestPeakRSS        int `json:"tofu_test_peak_rss_kib"`
	TofuTestExecutions     int `json:"tofu_test_executions"`
	TotalWallMS            int `json:"total_wall_ms"`
	MaxPeakRSSKiB          int `json:"max_peak_rss_kib"`
}

type Inventory struct {
	InputRegularFiles   int `json:"input_regular_files"`
	InputPhysicalLines  int `json:"input_physical_lines"`
	OutputArtifactFiles int `json:"output_artifact_files"`
}

type GraphBinding struct {
	CellID         string `json:"cell_id"`
	ActivityID     string `json:"activity_id"`
	InputID        string `json:"input_id"`
	OutputID       string `json:"output_id"`
	UsedEdgeCount  int    `json:"used_edge_count"`
	GeneratedCount int    `json:"generated_edge_count"`
}

type GraphNode struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type GraphRelation struct {
	Status    string `json:"status"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type GraphObservation struct {
	Schema        string          `json:"schema"`
	ProgramDigest string          `json:"program_digest"`
	GraphHash     string          `json:"graph_hash"`
	ActivityCount int             `json:"activity_count"`
	EdgeCount     int             `json:"edge_count"`
	Nodes         []GraphNode     `json:"nodes"`
	Relations     []GraphRelation `json:"relations"`
	Bindings      []GraphBinding  `json:"bindings"`
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
	ID         string  `json:"id"`
	Expected   string  `json:"expected"`
	Decision   string  `json:"decision"`
	Resolution string  `json:"resolution"`
	Reason     string  `json:"reason"`
	Unknown    *Unknown `json:"unknown,omitempty"`
}

type Observation struct {
	Schema                  string             `json:"schema"`
	ContractID              string             `json:"contract_id"`
	SubjectSHA              string             `json:"subject_sha"`
	UserPaths               []string           `json:"user_paths"`
	Release                 ReleaseObservation `json:"release"`
	FixtureDigest           string             `json:"fixture_digest"`
	FixtureFiles            []string           `json:"fixture_files"`
	FixturePhysicalLines    int                `json:"fixture_physical_lines"`
	Executions              []ExecutionRun     `json:"executions"`
	Reuse                   ReuseAccounting    `json:"reuse"`
	Runtime                 RuntimeSummary     `json:"runtime"`
	Inventory               Inventory          `json:"inventory"`
	ObserverGoVersion       string             `json:"observer_go_version"`
	ObserverGOVERSION       string             `json:"observer_go_env_goversion"`
	ObserverToolchainDigest string             `json:"observer_toolchain_digest"`
	CellEvidenceProjections map[string]string  `json:"cell_evidence_projections"`
	CellEvidenceDigests     map[string]string  `json:"cell_evidence_digests"`
	Graph                   GraphObservation   `json:"graph"`
	RepositoryWrites        int                `json:"repository_writes"`
	LocalTestExecutions     int                `json:"local_test_executions"`
	ReleaseBinaryBuilds     int                `json:"release_binary_builds"`
	ReleaseBinaryBuildReason string            `json:"release_binary_build_reason"`
	HumanReportReady        bool               `json:"human_report_ready"`
}
