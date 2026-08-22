package proposalpromotion

type selectionView struct {
	Repository        string `json:"repository"`
	CurrentSubjectSHA string `json:"current_subject_sha"`
	PredecessorSHA    string `json:"predecessor_sha"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	Selected          struct {
		RunID                int64  `json:"run_id"`
		RunAttempt           int    `json:"run_attempt"`
		HeadSHA              string `json:"head_sha"`
		Event                string `json:"event"`
		Status               string `json:"status"`
		Conclusion           string `json:"conclusion"`
		WorkflowName         string `json:"workflow_name"`
		SynthesisJobID       int64  `json:"synthesis_job_id"`
		SynthesisJobName     string `json:"synthesis_job_name"`
		SynthesisJobStatus   string `json:"synthesis_job_status"`
		SynthesisJobConclusion string `json:"synthesis_job_conclusion"`
		ArtifactID           int64  `json:"artifact_id"`
		ArtifactName         string `json:"artifact_name"`
		ProposalFileSHA256   string `json:"proposal_file_sha256"`
		ProposalReportDigest string `json:"proposal_report_digest"`
		RepositoryWrites     int    `json:"repository_writes"`
		PromotionAuthorized  bool   `json:"promotion_authorized"`
	} `json:"selected"`
	Summary struct {
		ObservedRuns         int `json:"observed_runs"`
		ExactRuns            int `json:"exact_runs"`
		ObservedArtifacts    int `json:"observed_artifacts"`
		ExactArtifacts       int `json:"exact_artifacts"`
		ObservedJobs         int `json:"observed_jobs"`
		ExactJobs            int `json:"exact_jobs"`
		ValidCandidates      int `json:"valid_candidates"`
		AmbiguousCandidates  int `json:"ambiguous_candidates"`
		UnresolvedCandidates int `json:"unresolved_candidates"`
		RepositoryWrites     int `json:"repository_writes"`
		SelectionBPS         int `json:"selection_bps"`
		ProofsPassed         int `json:"proofs_passed"`
		ProofsTotal          int `json:"proofs_total"`
	} `json:"summary"`
	ReportDigest string `json:"report_digest"`
}

type contractView struct {
	SubjectSHA      string `json:"subject_sha"`
	Decision        string `json:"decision"`
	Reason          string `json:"reason"`
	SelectedActions int    `json:"selected_actions"`
	Summary         struct {
		Satisfied    int `json:"satisfied"`
		Total        int `json:"total"`
		Unresolved   int `json:"unresolved"`
		ReadinessBPS int `json:"readiness_bps"`
	} `json:"summary"`
	RepositoryWrites    int    `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
	ReportDigest        string `json:"report_digest"`
}
