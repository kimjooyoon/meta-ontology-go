package main

const (
	evidenceSchema       = "gooo/ci-evidence/v3"
	proofSchema          = "gooo/ci-proof/v3"
	receiptSchema        = "gooo/provenance-receipt/v3"
	domainEvidenceSchema = "gooo/domain-evidence/v2"
	apiTerminalSuccess   = "api_terminal_success"
	observerLag          = "observer_lag"
)

var proofJobs = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type proofInputs struct {
	Governance governanceInput
	Evidence   evidenceInput
	Scheduler  []schedulerInput
	Jobs       []jobInput
	Context    contextInput
}

type governanceInput struct {
	Schema           string             `json:"schema"`
	RequiredContexts governanceContexts `json:"required_contexts"`
	GuardianContexts guardianContexts   `json:"guardian_contexts"`
	ProofJobs        []string           `json:"proof_jobs"`
	Promotion        promotionInput     `json:"promotion"`
}

type governanceContexts struct {
	Dev  []string `json:"dev"`
	Main []string `json:"main"`
}

type guardianContexts struct {
	DevShadow    string `json:"dev_shadow"`
	MainRequired string `json:"main_required"`
}

type promotionInput struct {
	Source                   string   `json:"source"`
	Target                   string   `json:"target"`
	RequiredChecks           []string `json:"required_checks"`
	BranchProtectionRequired bool     `json:"branch_protection_required"`
}

type evidenceInput struct {
	Schema      string           `json:"schema"`
	Repository  string           `json:"repository"`
	Event       string           `json:"event"`
	EventRef    string           `json:"event_ref"`
	CheckoutRef string           `json:"checkout_ref"`
	BaseRef     string           `json:"base_ref"`
	BaseSHA     string           `json:"base_sha"`
	HeadSHA     string           `json:"head_sha"`
	RunID       int64            `json:"run_id"`
	Attempt     int64            `json:"run_attempt"`
	WorkflowSHA string           `json:"workflow_sha"`
	Scheduler   []schedulerInput `json:"scheduler"`
	Jobs        []jobInput       `json:"jobs"`
	Digests     evidenceDigests  `json:"digests"`
}

type evidenceDigests struct {
	Source    string `json:"source_sha256"`
	IR        string `json:"ir_sha256"`
	Generated string `json:"generated_output_sha256"`
	Policy    string `json:"policy_sha256"`
	Toolchain string `json:"toolchain_sha256"`
	Bundle    string `json:"bundle_sha256"`
}

type jobInput struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Status           *string `json:"status"`
	Conclusion       *string `json:"conclusion"`
	HeadSHA          string  `json:"head_sha"`
	RunID            int64   `json:"run_id"`
	RunAttempt       int64   `json:"run_attempt"`
	CompletedAt      *string `json:"completed_at"`
	ObservationState string  `json:"observation_state"`
}

type schedulerInput struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	Result     string `json:"result"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}

type contextInput struct {
	Repository           string                `json:"repository"`
	Event                string                `json:"event"`
	Ref                  string                `json:"ref"`
	EventRef             string                `json:"event_ref"`
	CheckoutRef          string                `json:"checkout_ref"`
	BaseRef              string                `json:"base_ref"`
	BaseSHA              string                `json:"base_sha"`
	HeadSHA              string                `json:"head_sha"`
	WorkflowSHA          string                `json:"workflow_sha"`
	PRNumber             int64                 `json:"pr_number"`
	RunID                int64                 `json:"run_id"`
	RunAttempt           int64                 `json:"run_attempt"`
	Actor                string                `json:"actor"`
	Builder              string                `json:"builder"`
	Gate                 string                `json:"gate"`
	BranchProtected      bool                  `json:"branch_protected"`
	BranchProtection     branchProtection      `json:"branch_protection"`
	DevBranchProtection  branchProtection      `json:"dev_branch_protection"`
	DomainEvidence       domainEvidence        `json:"domain_evidence"`
	ScopeDecision        string                `json:"scope_decision"`
	FixtureStatus        string                `json:"fixture_status"`
	SourceStatus         string                `json:"source_status"`
	SemanticStatus       string                `json:"semantic_status"`
	ProvenanceStatus     string                `json:"provenance_status"`
	ArtifactsStatus      string                `json:"artifacts_status"`
	WriteEffect          string                `json:"write_effect"`
	NoWrite              bool                  `json:"no_write_outside_generated"`
	Artifacts            []artifactInput       `json:"artifacts"`
	FixturePaths         []string              `json:"fixture_paths"`
	Cache                cacheInput            `json:"cache"`
	DiagnosticIDs        []string              `json:"diagnostic_ids"`
	RepairIDs            []string              `json:"repair_ids"`
	Predecessors         []string              `json:"predecessors"`
	MissingReasons       missingReasons        `json:"missing_reasons"`
	GuardianEvidence     *guardianEvidence     `json:"guardian_evidence,omitempty"`
	PromotionObservation *promotionObservation `json:"promotion_observation,omitempty"`
}

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

type provenanceReceipt struct {
	Schema                 string                  `json:"schema"`
	Operation              string                  `json:"operation"`
	Relation               string                  `json:"relation"`
	Delta                  string                  `json:"delta"`
	AllowedIntent          string                  `json:"allowed_intent"`
	Locality               string                  `json:"locality"`
	Repository             string                  `json:"repository"`
	Event                  string                  `json:"event"`
	BaseRef                string                  `json:"base_ref"`
	BaseSHA                string                  `json:"base_sha"`
	HeadSHA                string                  `json:"head_sha"`
	Ref                    string                  `json:"ref"`
	EventRef               string                  `json:"event_ref"`
	CheckoutRef            string                  `json:"checkout_ref"`
	PRNumber               int64                   `json:"pr_number"`
	RunID                  int64                   `json:"run_id"`
	RunAttempt             int64                   `json:"run_attempt"`
	WorkflowSHA            string                  `json:"workflow_sha"`
	Scheduler              []schedulerInput        `json:"scheduler"`
	BranchProtection       branchProtection        `json:"branch_protection"`
	DevBranchProtection    branchProtection        `json:"dev_branch_protection"`
	DomainEvidence         domainEvidence          `json:"domain_evidence"`
	Jobs                   []jobInput              `json:"jobs"`
	Artifacts              []artifactInput         `json:"artifacts"`
	Digests                receiptDigests          `json:"digests"`
	Cache                  cacheReceipt            `json:"cache"`
	DiagnosticIDs          []string                `json:"diagnostic_ids"`
	RepairIDs              []string                `json:"repair_ids"`
	WriteEffect            string                  `json:"write_effect"`
	Producer               string                  `json:"producer"`
	Role                   string                  `json:"role"`
	Predecessors           []string                `json:"predecessors"`
	Decision               string                  `json:"decision"`
	MissingReasons         missingReasons          `json:"missing_reasons"`
	GuardianEvidence       *guardianEvidence       `json:"guardian_evidence,omitempty"`
	PromotionObservation   *promotionObservation   `json:"promotion_observation,omitempty"`
	PromotionAuthorization *promotionAuthorization `json:"promotion_authorization,omitempty"`
}

type receiptDigests struct {
	Source     string `json:"source_sha256"`
	IR         string `json:"ir_sha256"`
	Projection string `json:"projection_sha256"`
	Build      string `json:"build_sha256"`
	Policy     string `json:"policy_sha256"`
	Schema     string `json:"schema_sha256"`
	Toolchain  string `json:"toolchain_sha256"`
	Target     string `json:"target_sha256"`
	Bundle     string `json:"bundle_sha256"`
}

type cacheReceipt struct {
	Key     string `json:"cache_key"`
	Outcome string `json:"outcome"`
}
