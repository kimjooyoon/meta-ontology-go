package proposalpromotion

type Source struct {
	Selection SelectionSource `json:"selection"`
	Contract  ContractSource  `json:"contract"`
}

type SelectionSource struct {
	Repository                  string `json:"repository"`
	CurrentSubjectSHA           string `json:"current_subject_sha"`
	PredecessorSHA              string `json:"predecessor_sha"`
	Decision                    string `json:"decision"`
	Reason                      string `json:"reason"`
	ReportDigest                string `json:"report_digest"`
	RunID                       int64  `json:"run_id"`
	RunAttempt                  int    `json:"run_attempt"`
	HeadSHA                     string `json:"head_sha"`
	Event                       string `json:"event"`
	Status                      string `json:"status"`
	Conclusion                  string `json:"conclusion"`
	WorkflowName                string `json:"workflow_name"`
	ArtifactID                  int64  `json:"artifact_id"`
	ArtifactName                string `json:"artifact_name"`
	ProposalFileSHA256          string `json:"proposal_file_sha256"`
	ProposalReportDigest        string `json:"proposal_report_digest"`
	ObservedRuns                int    `json:"observed_runs"`
	ExactRuns                   int    `json:"exact_runs"`
	ObservedArtifacts           int    `json:"observed_artifacts"`
	ExactArtifacts              int    `json:"exact_artifacts"`
	ValidCandidates             int    `json:"valid_candidates"`
	AmbiguousCandidates         int    `json:"ambiguous_candidates"`
	UnresolvedCandidates        int    `json:"unresolved_candidates"`
	SelectionBPS                int    `json:"selection_bps"`
	ProofsPassed                int    `json:"proofs_passed"`
	ProofsTotal                 int    `json:"proofs_total"`
	RepositoryWrites            int    `json:"repository_writes"`
	SelectedRepositoryWrites    int    `json:"selected_repository_writes"`
	SelectedPromotionAuthorized bool   `json:"selected_promotion_authorized"`
}

type ContractSource struct {
	SubjectSHA          string `json:"subject_sha"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	FileSHA256          string `json:"file_sha256"`
	ReportDigest        string `json:"report_digest"`
	SelectedActions     int    `json:"selected_actions"`
	Satisfied           int    `json:"satisfied"`
	Total               int    `json:"total"`
	Unresolved          int    `json:"unresolved"`
	ReadinessBPS        int    `json:"readiness_bps"`
	RepositoryWrites    int    `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}
