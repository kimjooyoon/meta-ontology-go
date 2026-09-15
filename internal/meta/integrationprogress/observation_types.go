package integrationprogress

type Observation struct {
	Schema          string            `json:"schema"`
	Repository      string            `json:"repository"`
	ObserverHeadSHA string            `json:"observer_head_sha"`
	ObservedAt      string            `json:"observed_at"`
	CohortID        string            `json:"cohort_id"`
	QueueSnapshot   QueueSnapshot     `json:"queue_snapshot"`
	PullRequests    []PullObservation `json:"pull_requests"`
}

type QueueSnapshot struct {
	ObservationStatus string `json:"observation_status"`
	FailureReason     string `json:"failure_reason,omitempty"`
	QueuedRuns        int    `json:"queued_runs"`
	InProgressRuns    int    `json:"in_progress_runs"`
}

type PullObservation struct {
	Number            int             `json:"number"`
	ObservationStatus string          `json:"observation_status"`
	FailureReason     string          `json:"failure_reason,omitempty"`
	State             string          `json:"state,omitempty"`
	Draft             bool            `json:"draft"`
	HeadSHA           string          `json:"head_sha,omitempty"`
	CreatedAt         string          `json:"created_at,omitempty"`
	ClosedAt          string          `json:"closed_at,omitempty"`
	MergedAt          string          `json:"merged_at,omitempty"`
	RunsTotal         int             `json:"runs_total"`
	RunsConsumed      int             `json:"runs_consumed"`
	RunQueryFailure   string          `json:"run_query_failure,omitempty"`
	RunSelection      string          `json:"run_selection,omitempty"`
	EligibleRuns      int             `json:"eligible_runs"`
	AuthoritativeRun  *RunObservation `json:"authoritative_run,omitempty"`
}

type RunObservation struct {
	ID                   int64                `json:"id"`
	Name                 string               `json:"name"`
	HeadSHA              string               `json:"head_sha"`
	Status               string               `json:"status"`
	Conclusion           string               `json:"conclusion"`
	CreatedAt            string               `json:"created_at"`
	StartedAt            string               `json:"run_started_at"`
	UpdatedAt            string               `json:"updated_at"`
	ArtifactsTotal       int                  `json:"artifacts_total"`
	ArtifactsConsumed    int                  `json:"artifacts_consumed"`
	ArtifactMatches      int                  `json:"artifact_matches"`
	ArtifactQueryFailure string               `json:"artifact_query_failure,omitempty"`
	Artifact             *ArtifactObservation `json:"artifact,omitempty"`
}

type ArtifactObservation struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	HeadSHA   string `json:"head_sha"`
	CreatedAt string `json:"created_at"`
	Expired   bool   `json:"expired"`
}
