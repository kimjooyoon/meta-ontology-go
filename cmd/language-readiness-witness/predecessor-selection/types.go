package main

type workflowRunList struct {
	TotalCount int `json:"total_count"`

	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

type workflowRun struct {
	ID int64 `json:"id"`

	Name string `json:"name"`

	HeadBranch string `json:"head_branch"`

	HeadSHA string `json:"head_sha"`

	Event string `json:"event"`

	Status string `json:"status"`

	Conclusion string `json:"conclusion"`

	RunAttempt int `json:"run_attempt"`
}

type artifactList struct {
	TotalCount int `json:"total_count"`

	Artifacts []artifactMetadata `json:"artifacts"`
}

type artifactMetadata struct {
	ID int64 `json:"id"`

	Name string `json:"name"`

	Expired bool `json:"expired"`
}

type commit struct {
	SHA string `json:"sha"`

	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}
