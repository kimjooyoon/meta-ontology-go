package guardedpromotion

type WorkflowEvidence struct {
	RunID      int64  `json:"run_id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	HeadBranch string `json:"head_branch"`
}

type ArtifactEvidence struct {
	RunID                  int64  `json:"run_id"`
	RunAttempt             int    `json:"run_attempt"`
	RunEvent               string `json:"run_event"`
	ArtifactID             int64  `json:"artifact_id"`
	ArtifactName           string `json:"artifact_name"`
	ArtifactDigest         string `json:"artifact_digest"`
	FileSHA256             string `json:"file_sha256"`
	ReportSchema           string `json:"report_schema"`
	ReportDigest           string `json:"report_digest"`
	ReportCurrentHeadSHA   string `json:"report_current_head_sha"`
	ReportDecision         string `json:"report_decision"`
	ReportSatisfied        int    `json:"report_satisfied"`
	ReportTotal            int    `json:"report_total"`
	ReportUnresolved       int    `json:"report_unresolved"`
	ReportRepositoryWrites int    `json:"report_repository_writes"`
}

type Source struct {
	RequestedRepository          string           `json:"requested_repository"`
	ObservedRepository           string           `json:"observed_repository"`
	DefaultBranch                string           `json:"default_branch"`
	CurrentHeadSHA               string           `json:"current_head_sha"`
	PredecessorSHA               string           `json:"predecessor_sha"`
	Workflow                     WorkflowEvidence `json:"workflow"`
	Artifact                     ArtifactEvidence `json:"artifact"`
	ObservedRuns                 int              `json:"observed_runs"`
	ObservedArtifacts            int              `json:"observed_artifacts"`
	ValidCandidates              int              `json:"valid_candidates"`
	AmbiguousCandidates          int              `json:"ambiguous_candidates"`
	UnresolvedCandidates         int              `json:"unresolved_candidates"`
	RepositoryWrites             int              `json:"repository_writes"`
	RepositoryMutationAuthorized bool             `json:"repository_mutation_authorized"`
	CollectionError              string           `json:"collection_error,omitempty"`
}
