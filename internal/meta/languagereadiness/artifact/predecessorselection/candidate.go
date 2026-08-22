package predecessorselection

type Input struct {
	Repository     string      `json:"repository"`
	CurrentHeadSHA string      `json:"current_head_sha"`
	PredecessorSHA string      `json:"predecessor_sha"`
	Branch         string      `json:"branch"`
	Workflow       string      `json:"workflow"`
	Candidates     []Candidate `json:"candidates"`
}

type Candidate struct {
	RunID                   int64  `json:"run_id"`
	RunAttempt              int    `json:"run_attempt"`
	Workflow                string `json:"workflow"`
	HeadBranch              string `json:"head_branch"`
	HeadSHA                 string `json:"head_sha"`
	Event                   string `json:"event"`
	Conclusion              string `json:"conclusion"`
	ReadinessArtifactID     int64  `json:"readiness_artifact_id"`
	ReadinessArtifactName   string `json:"readiness_artifact_name"`
	ReadinessExpired        bool   `json:"readiness_expired"`
	ReadinessPayloadBase64 string `json:"readiness_payload_base64,omitempty"`
	BindingArtifactID       int64  `json:"binding_artifact_id"`
	BindingArtifactName     string `json:"binding_artifact_name"`
	BindingExpired          bool   `json:"binding_expired"`
	BindingPayloadBase64   string `json:"binding_payload_base64,omitempty"`
	RepositoryWrites        int    `json:"repository_writes"`
}
