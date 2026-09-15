package main

type guardianEnvironment struct {
	Repository             string                         `json:"repository"`
	Name                   string                         `json:"name"`
	DeploymentBranchPolicy guardianDeploymentBranchPolicy `json:"deployment_branch_policy"`
	ProtectionRules        []string                       `json:"protection_rules"`
	WaitTimer              int                            `json:"wait_timer"`
	Reviewers              []string                       `json:"reviewers"`
	TokenSource            string                         `json:"token_source"`
	ReadStatus             string                         `json:"read_status"`
	MissingReason          string                         `json:"missing_reason"`
	RunID                  int64                          `json:"run_id"`
	RunAttempt             int64                          `json:"run_attempt"`
	WorkflowSHA            string                         `json:"workflow_sha"`
	ObservedAt             *string                        `json:"observed_at"`
	ValidUntil             *string                        `json:"valid_until"`
	Digest                 string                         `json:"digest_sha256"`
}
type guardianInstallationScope struct {
	Repository      string   `json:"repository"`
	InstallationID  int64    `json:"installation_id"`
	TokenSource     string   `json:"token_source"`
	ReadStatus      string   `json:"read_status"`
	RepositoryCount int      `json:"repository_count"`
	Repositories    []string `json:"repositories"`
	ExactMatch      bool     `json:"exact_match"`
	MissingReason   string   `json:"missing_reason"`
	RunID           int64    `json:"run_id"`
	RunAttempt      int64    `json:"run_attempt"`
	WorkflowSHA     string   `json:"workflow_sha"`
	ObservedAt      *string  `json:"observed_at"`
	ValidUntil      *string  `json:"valid_until"`
	Digest          string   `json:"digest_sha256"`
}
type guardianDeploymentBranchPolicy struct {
	ProtectedBranches    bool `json:"protected_branches"`
	CustomBranchPolicies bool `json:"custom_branch_policies"`
}
type missingReasons struct {
	Protection string `json:"protection,omitempty"`
	Provenance string `json:"provenance,omitempty"`
}
