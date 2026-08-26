package main

type branchProtection struct {
	Repository                     string                 `json:"repository"`
	Branch                         string                 `json:"branch"`
	PolicySHA                      string                 `json:"policy_sha256"`
	EventRef                       string                 `json:"event_ref"`
	CheckoutRef                    string                 `json:"checkout_ref"`
	TokenSource                    string                 `json:"token_source"`
	AppInstallationID              int64                  `json:"app_installation_id"`
	AppSlug                        string                 `json:"app_slug"`
	ReadStatus                     string                 `json:"read_status"`
	Exists                         bool                   `json:"exists"`
	Strict                         bool                   `json:"strict"`
	RequiredChecks                 []string               `json:"required_checks"`
	RequiredCheckBindings          []requiredCheckBinding `json:"required_check_bindings"`
	EnforceAdmins                  bool                   `json:"enforce_admins"`
	RequiredReviews                int64                  `json:"required_reviews"`
	DismissStaleReviews            bool                   `json:"dismiss_stale_reviews"`
	RequireLastPushApproval        bool                   `json:"require_last_push_approval"`
	LinearHistory                  bool                   `json:"linear_history"`
	AllowForcePushes               bool                   `json:"allow_force_pushes"`
	AllowDeletions                 bool                   `json:"allow_deletions"`
	RequiredSignatures             bool                   `json:"required_signatures"`
	RequiredConversationResolution bool                   `json:"required_conversation_resolution"`
	BlockCreations                 bool                   `json:"block_creations"`
	LockBranch                     bool                   `json:"lock_branch"`
	AllowForkSyncing               bool                   `json:"allow_fork_syncing"`
	Restrictions                   any                    `json:"restrictions"`
	MissingReason                  string                 `json:"missing_reason"`
	BaseSHA                        string                 `json:"base_sha"`
	HeadSHA                        string                 `json:"head_sha"`
	RunID                          int64                  `json:"run_id"`
	RunAttempt                     int64                  `json:"run_attempt"`
	WorkflowSHA                    string                 `json:"workflow_sha"`
	ObservedAt                     *string                `json:"observed_at"`
	ValidUntil                     *string                `json:"valid_until"`
	Digest                         string                 `json:"digest_sha256"`
}
type artifactInput struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Size       int64  `json:"size_bytes"`
	Expired    bool   `json:"expired"`
	Digest     string `json:"digest"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}
type guardianLiveRefs struct {
	DevSHA  string `json:"dev_sha"`
	MainSHA string `json:"main_sha"`
}
type guardianTopology struct {
	Status       string `json:"status"`
	AheadBy      int    `json:"ahead_by"`
	BehindBy     int    `json:"behind_by"`
	MergeBaseSHA string `json:"merge_base_sha"`
}
