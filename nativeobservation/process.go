package nativeobservation

// ProcessObservation records one observed process. A zero value describes no
// recorded process; it does not assert a successful execution.
type ProcessObservation struct {
	Started          bool           `json:"started"`
	Completed        bool           `json:"completed"`
	Canceled         bool           `json:"canceled"`
	TimedOut         bool           `json:"timed_out"`
	DiagnosticsBytes int            `json:"diagnostics_bytes"`
	OutputTruncated  bool           `json:"output_truncated"`
	ExitCode         *int           `json:"exit_code"`
	WallNS           int64          `json:"wall_ns"`
	UserNS           int64          `json:"user_ns"`
	SystemNS         int64          `json:"system_ns"`
	PeakRSSBytes     *int64         `json:"peak_rss_bytes"`
	StdoutSHA256     string         `json:"stdout_sha256"`
	StderrSHA256     string         `json:"stderr_sha256"`
	Failure          string         `json:"failure,omitempty"`
	Diagnostics      []byte         `json:"diagnostics,omitempty"`
	Timing           *ProcessTiming `json:"timing,omitempty"`
	FailurePhase     string         `json:"failure_phase,omitempty"`
}
