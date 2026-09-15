package selfimprovementtransport

type ArtifactSelectionInput struct {
	Repository         string
	ExpectedRunID      int64
	ExpectedRunAttempt int
	ArtifactName       string
}

type workflowRunAPI struct {
	ID         int64  `json:"id"`
	RunAttempt int    `json:"run_attempt"`
	HeadSHA    string `json:"head_sha"`
	Path       string `json:"path"`
}

type artifactListAPI struct {
	Artifacts []artifactAPI `json:"artifacts"`
}

type artifactAPI struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Expired     bool                   `json:"expired"`
	ExpiresAt   string                 `json:"expires_at"`
	Digest      string                 `json:"digest"`
	SizeInBytes int64                  `json:"size_in_bytes"`
	WorkflowRun artifactWorkflowRunAPI `json:"workflow_run"`
}

type artifactWorkflowRunAPI struct {
	ID      int64  `json:"id"`
	HeadSHA string `json:"head_sha"`
}
