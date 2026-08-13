package main

const (
	evidenceSchema       = "gooo/ci-evidence/v2"
	proofSchema          = "gooo/ci-proof/v3"
	receiptSchema        = "gooo/provenance-receipt/v3"
	domainEvidenceSchema = "gooo/domain-evidence/v2"
)

var proofJobs = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type proofInputs struct {
	Governance governanceInput
	Evidence   evidenceInput
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
	Schema      string          `json:"schema"`
	Repository  string          `json:"repository"`
	Event       string          `json:"event"`
	EventRef    string          `json:"event_ref"`
	CheckoutRef string          `json:"checkout_ref"`
	BaseRef     string          `json:"base_ref"`
	BaseSHA     string          `json:"base_sha"`
	HeadSHA     string          `json:"head_sha"`
	RunID       int64           `json:"run_id"`
	Attempt     int64           `json:"run_attempt"`
	WorkflowSHA string          `json:"workflow_sha"`
	Jobs        []jobInput      `json:"jobs"`
	Digests     evidenceDigests `json:"digests"`
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
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
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
	Repository              string                 `json:"repository"`
	Branch                  string                 `json:"branch"`
	PolicySHA               string                 `json:"policy_sha256"`
	EventRef                string                 `json:"event_ref"`
	CheckoutRef             string                 `json:"checkout_ref"`
	TokenSource             string                 `json:"token_source"`
	ReadStatus              string                 `json:"read_status"`
	Exists                  bool                   `json:"exists"`
	Strict                  bool                   `json:"strict"`
	RequiredChecks          []string               `json:"required_checks"`
	RequiredCheckBindings   []requiredCheckBinding `json:"required_check_bindings"`
	EnforceAdmins           bool                   `json:"enforce_admins"`
	RequiredReviews         int64                  `json:"required_reviews"`
	DismissStaleReviews     bool                   `json:"dismiss_stale_reviews"`
	RequireLastPushApproval bool                   `json:"require_last_push_approval"`
	LinearHistory           bool                   `json:"linear_history"`
	AllowForcePushes        bool                   `json:"allow_force_pushes"`
	AllowDeletions          bool                   `json:"allow_deletions"`
	MissingReason           string                 `json:"missing_reason"`
	BaseSHA                 string                 `json:"base_sha"`
	HeadSHA                 string                 `json:"head_sha"`
	RunID                   int64                  `json:"run_id"`
	RunAttempt              int64                  `json:"run_attempt"`
	WorkflowSHA             string                 `json:"workflow_sha"`
	Digest                  string                 `json:"digest_sha256"`
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

type guardianEvidence struct {
	Schema                string           `json:"schema"`
	Route                 string           `json:"route"`
	CheckName             string           `json:"check_name"`
	Repository            string           `json:"repository"`
	PRNumber              int64            `json:"pr_number"`
	Action                string           `json:"action"`
	BaseRepo              string           `json:"base_repo"`
	BaseRef               string           `json:"base_ref"`
	BaseSHA               string           `json:"base_sha"`
	HeadRepo              string           `json:"head_repo"`
	HeadRef               string           `json:"head_ref"`
	HeadSHA               string           `json:"head_sha"`
	WorkflowRef           string           `json:"workflow_ref"`
	WorkflowSHA           string           `json:"workflow_sha"`
	RuntimeRef            string           `json:"runtime_ref"`
	RuntimeSHA            string           `json:"runtime_sha"`
	EventRef              string           `json:"event_ref"`
	DefaultBranch         string           `json:"default_branch"`
	RunID                 int64            `json:"run_id"`
	RunAttempt            int64            `json:"run_attempt"`
	WorkflowID            int64            `json:"workflow_id"`
	WorkflowPath          string           `json:"workflow_path"`
	RunEvent              string           `json:"run_event"`
	RunStatus             string           `json:"run_status"`
	RunConclusion         string           `json:"run_conclusion"`
	RunCreatedAt          string           `json:"run_created_at"`
	RunNumber             int64            `json:"run_number"`
	LiveRefsBefore        guardianLiveRefs `json:"live_refs_before"`
	LiveRefsAfter         guardianLiveRefs `json:"live_refs_after"`
	Topology              guardianTopology `json:"topology"`
	ArtifactID            int64            `json:"artifact_id"`
	ArtifactName          string           `json:"artifact_name"`
	ArtifactSize          int64            `json:"artifact_size"`
	ArtifactExpired       bool             `json:"artifact_expired"`
	ArtifactDigest        string           `json:"artifact_digest"`
	ManifestBundleSHA     string           `json:"manifest_bundle_sha256"`
	GuardianJobID         int64            `json:"guardian_job_id"`
	GuardianJobName       string           `json:"guardian_job_name"`
	GuardianJobStatus     string           `json:"guardian_job_status"`
	GuardianJobConclusion string           `json:"guardian_job_conclusion"`
	GuardianJobHeadSHA    string           `json:"guardian_job_head_sha"`
	CheckRunID            int64            `json:"check_run_id"`
	CheckRunName          string           `json:"check_run_name"`
	CheckRunAppID         int64            `json:"check_run_app_id"`
	CheckRunStatus        string           `json:"check_run_status"`
	CheckRunConclusion    string           `json:"check_run_conclusion"`
	CheckRunHeadSHA       string           `json:"check_run_head_sha"`
	CheckSuiteID          int64            `json:"check_suite_id"`
	Decision              string           `json:"decision"`
	Code                  *string          `json:"code"`
	HeadBindingStatus     string           `json:"head_binding_status"`
	BundleSHA             string           `json:"bundle_sha256"`
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
	BranchProtection       branchProtection        `json:"branch_protection"`
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
