package integrationprogress

type Cell struct {
	PullRequest int    `json:"pull_request"`
	Stage       string `json:"stage"`
	Step        string `json:"step"`
	State       string `json:"state"`
	Reason      string `json:"reason"`
	Resolution  string `json:"resolution"`
	EvidenceRef string `json:"evidence_ref,omitempty"`
}

type Summary struct {
	PullRequestsTotal              int   `json:"pull_requests_total"`
	PullRequestsObserved           int   `json:"pull_requests_observed"`
	CellsTotal                     int   `json:"cells_total"`
	ClosedCells                    int   `json:"closed_cells"`
	OpenCells                      int   `json:"open_cells"`
	UnknownCells                   int   `json:"unknown_cells"`
	RefutedCells                   int   `json:"refuted_cells"`
	DenominatorConflicts           int   `json:"denominator_conflicts"`
	TerminalRuns                   int   `json:"terminal_runs"`
	EvidenceReachable              int   `json:"evidence_reachable"`
	MergedPullRequests             int   `json:"merged_pull_requests"`
	EvidencedMerges                int   `json:"evidenced_merges"`
	TimingSamples                  int   `json:"timing_samples"`
	EvidenceLatencySamples         int   `json:"evidence_latency_samples"`
	MergeDelaySamples              int   `json:"merge_delay_samples"`
	ProgressBasisPoints            int   `json:"progress_basis_points"`
	MergeBasisPoints               int   `json:"merge_basis_points"`
	EvidenceBasisPoints            int   `json:"evidence_basis_points"`
	QueueObservationUnknown        int   `json:"queue_observation_unknown"`
	QueuedRunsSnapshot             int   `json:"queued_runs_snapshot"`
	InProgressRunsSnapshot         int   `json:"in_progress_runs_snapshot"`
	QueuePressureBasisPoints       int   `json:"queue_pressure_basis_points"`
	RunStartDelaySecondsTotal      int64 `json:"run_start_delay_seconds_total"`
	ExecutionSecondsTotal          int64 `json:"execution_seconds_total"`
	EvidenceLatencySecondsTotal    int64 `json:"evidence_latency_seconds_total"`
	MergeAfterEvidenceSecondsTotal int64 `json:"merge_after_evidence_seconds_total"`
}

type Report struct {
	Schema              string      `json:"schema"`
	Repository          string      `json:"repository"`
	ObserverHeadSHA     string      `json:"observer_head_sha"`
	CohortID            string      `json:"cohort_id"`
	Decision            string      `json:"decision"`
	Reason              string      `json:"reason"`
	Resolution          string      `json:"resolution"`
	ObservationDigest   string      `json:"observation_digest"`
	MetaProgramDigest   string      `json:"meta_program_digest"`
	Summary             Summary     `json:"summary"`
	Cells               []Cell      `json:"cells"`
	Indicators          []Indicator `json:"indicators"`
	Proofs              []Proof     `json:"proofs"`
	RepositoryWrites    int         `json:"repository_writes"`
	PromotionAuthorized bool        `json:"promotion_authorized"`
	ReportDigest        string      `json:"report_digest"`
}
