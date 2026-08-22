package guardedpromotion

type repositoryResponse struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
}

type pullReference struct {
	Number int `json:"number"`
}

type workflowRunResponse struct {
	ID           int64           `json:"id"`
	RunAttempt   int             `json:"run_attempt"`
	Name         string          `json:"name"`
	Path         string          `json:"path"`
	Event        string          `json:"event"`
	Status       string          `json:"status"`
	Conclusion   string          `json:"conclusion"`
	HeadSHA      string          `json:"head_sha"`
	HeadBranch   string          `json:"head_branch"`
	PullRequests []pullReference `json:"pull_requests"`
}

type workflowRunsResponse struct {
	WorkflowRuns []workflowRunResponse `json:"workflow_runs"`
}

type artifactResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Digest  string `json:"digest"`
	Expired bool   `json:"expired"`
}

type artifactsResponse struct {
	Artifacts []artifactResponse `json:"artifacts"`
}

type commitResponse struct {
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

type pullResponse struct {
	Base struct {
		SHA string `json:"sha"`
	} `json:"base"`
}

type promotionEnvelope struct {
	Schema         string `json:"schema"`
	CurrentHeadSHA string `json:"current_head_sha"`
	Decision       string `json:"decision"`
	ReportDigest   string `json:"report_digest"`
	Summary        struct {
		Satisfied        int `json:"satisfied"`
		Total            int `json:"total"`
		Unresolved       int `json:"unresolved"`
		RepositoryWrites int `json:"repository_writes"`
	} `json:"summary"`
}
