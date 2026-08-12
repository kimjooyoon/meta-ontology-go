package main

const (
	proofSchema   = "gooo/ci-proof/v1"
	receiptSchema = "gooo/provenance-receipt/v1"
)

var proofJobs = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type proofInputs struct {
	Governance governanceInput
	Evidence   evidenceInput
	Jobs       []jobInput
	Context    contextInput
}

type governanceInput struct {
	Schema    string         `json:"schema"`
	Promotion promotionInput `json:"promotion"`
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
	Status     string `json:"status,omitempty"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

type contextInput struct {
	Repository       string           `json:"repository"`
	Event            string           `json:"event"`
	Ref              string           `json:"ref"`
	EventRef         string           `json:"event_ref"`
	CheckoutRef      string           `json:"checkout_ref"`
	BaseRef          string           `json:"base_ref"`
	BaseSHA          string           `json:"base_sha"`
	HeadSHA          string           `json:"head_sha"`
	WorkflowSHA      string           `json:"workflow_sha"`
	PRNumber         int64            `json:"pr_number"`
	RunID            int64            `json:"run_id"`
	RunAttempt       int64            `json:"run_attempt"`
	Actor            string           `json:"actor"`
	Builder          string           `json:"builder"`
	Guardian         string           `json:"guardian"`
	Approver         string           `json:"approver"`
	Gate             string           `json:"gate"`
	BranchProtected  bool             `json:"branch_protected"`
	BranchProtection branchProtection `json:"branch_protection"`
	ScopeDecision    string           `json:"scope_decision"`
	FixtureStatus    string           `json:"fixture_status"`
	SourceStatus     string           `json:"source_status"`
	SemanticStatus   string           `json:"semantic_status"`
	ProvenanceStatus string           `json:"provenance_status"`
	ArtifactsStatus  string           `json:"artifacts_status"`
	ApprovalsStatus  string           `json:"approvals_status"`
	WriteEffect      string           `json:"write_effect"`
	NoWrite          bool             `json:"no_write_outside_generated"`
	Approvals        []approvalInput  `json:"approvals"`
	Artifacts        []artifactInput  `json:"artifacts"`
	FixturePaths     []string         `json:"fixture_paths"`
	Cache            cacheInput       `json:"cache"`
	DiagnosticIDs    []string         `json:"diagnostic_ids"`
	RepairIDs        []string         `json:"repair_ids"`
	Predecessors     []string         `json:"predecessors"`
}

type branchProtection struct {
	Repository              string   `json:"repository"`
	Branch                  string   `json:"branch"`
	PolicySHA               string   `json:"policy_sha256"`
	EventRef                string   `json:"event_ref"`
	CheckoutRef             string   `json:"checkout_ref"`
	TokenSource             string   `json:"token_source"`
	ReadStatus              string   `json:"read_status"`
	Exists                  bool     `json:"exists"`
	Strict                  bool     `json:"strict"`
	RequiredChecks          []string `json:"required_checks"`
	EnforceAdmins           bool     `json:"enforce_admins"`
	RequiredReviews         int64    `json:"required_reviews"`
	DismissStaleReviews     bool     `json:"dismiss_stale_reviews"`
	RequireLastPushApproval bool     `json:"require_last_push_approval"`
	LinearHistory           bool     `json:"linear_history"`
	AllowForcePushes        bool     `json:"allow_force_pushes"`
	AllowDeletions          bool     `json:"allow_deletions"`
	BaseSHA                 string   `json:"base_sha"`
	HeadSHA                 string   `json:"head_sha"`
	RunID                   int64    `json:"run_id"`
	RunAttempt              int64    `json:"run_attempt"`
	WorkflowSHA             string   `json:"workflow_sha"`
	Digest                  string   `json:"digest_sha256"`
}

type approvalInput struct {
	Actor string `json:"actor"`
	State string `json:"state"`
}

type artifactInput struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Size    int64  `json:"size_bytes"`
	Expired bool   `json:"expired"`
}

type cacheInput struct {
	Key                   string   `json:"cache_key"`
	Outcome               string   `json:"outcome"`
	Status                string   `json:"status"`
	ArtifactKind          string   `json:"artifact_kind"`
	SemanticClosureDigest string   `json:"semantic_closure_digest"`
	DependencyRoot        string   `json:"dependency_root"`
	DirectDependencies    []string `json:"direct_dependencies"`
	PolicyDigest          string   `json:"policy_digest"`
	SchemaDigest          string   `json:"schema_digest"`
	ToolchainDigest       string   `json:"toolchain_digest"`
	TargetDigest          string   `json:"target_digest"`
	BuildTags             []string `json:"build_tags"`
	EvidenceRefs          []string `json:"evidence_refs"`
	ProducerHost          string   `json:"producer_host"`
	ContentSize           int64    `json:"content_size"`
	HitContentSize        int64    `json:"hit_content_size"`
	Predecessor           string   `json:"predecessor"`
}

type proofBundle struct {
	Schema           string           `json:"schema"`
	Repository       string           `json:"repository"`
	Event            string           `json:"event"`
	PRNumber         int64            `json:"pr_number"`
	BaseRef          string           `json:"base_ref"`
	BaseSHA          string           `json:"base_sha"`
	HeadSHA          string           `json:"head_sha"`
	Ref              string           `json:"ref"`
	EventRef         string           `json:"event_ref"`
	CheckoutRef      string           `json:"checkout_ref"`
	RunID            int64            `json:"run_id"`
	RunAttempt       int64            `json:"run_attempt"`
	WorkflowSHA      string           `json:"workflow_sha"`
	Jobs             []jobInput       `json:"jobs"`
	Actors           actorRoles       `json:"actors"`
	BranchProtection branchProtection `json:"branch_protection"`
	Scope            scopeResult      `json:"scope"`
	Fixtures         fixtureResult    `json:"fixtures"`
	Artifacts        []artifactInput  `json:"artifacts"`
	Approvals        []approvalInput  `json:"approvals"`
	Cache            cacheInput       `json:"cache"`
	Digests          proofDigests     `json:"digests"`
	WriteEffect      string           `json:"write_effect"`
	Decision         string           `json:"decision"`
	NoWrite          bool             `json:"no_write_outside_generated"`
	Rejections       []string         `json:"rejections"`
	Predecessors     []string         `json:"predecessors"`
}

type actorRoles struct {
	Actor    string `json:"actor"`
	Builder  string `json:"builder"`
	Guardian string `json:"guardian"`
	Approver string `json:"approver"`
	Gate     string `json:"gate"`
}

type scopeResult struct {
	Decision string `json:"decision"`
	Status   string `json:"status"`
}

type fixtureResult struct {
	Paths      []string `json:"paths"`
	Status     string   `json:"status"`
	Source     string   `json:"source_status"`
	Semantic   string   `json:"semantic_status"`
	Provenance string   `json:"provenance_status"`
}

type proofDigests struct {
	Source     string `json:"source_sha256"`
	Semantic   string `json:"semantic_sha256"`
	Provenance string `json:"provenance_sha256"`
	Projection string `json:"projection_sha256"`
	Build      string `json:"build_sha256"`
	Policy     string `json:"policy_sha256"`
	Schema     string `json:"schema_sha256"`
	Toolchain  string `json:"toolchain_sha256"`
	Target     string `json:"target_sha256"`
	Bundle     string `json:"bundle_sha256"`
}

type provenanceReceipt struct {
	Schema           string           `json:"schema"`
	Operation        string           `json:"operation"`
	Relation         string           `json:"relation"`
	Delta            string           `json:"delta"`
	AllowedIntent    string           `json:"allowed_intent"`
	Locality         string           `json:"locality"`
	Event            string           `json:"event"`
	BaseRef          string           `json:"base_ref"`
	BaseSHA          string           `json:"base_sha"`
	HeadSHA          string           `json:"head_sha"`
	Ref              string           `json:"ref"`
	EventRef         string           `json:"event_ref"`
	CheckoutRef      string           `json:"checkout_ref"`
	PRNumber         int64            `json:"pr_number"`
	RunID            int64            `json:"run_id"`
	RunAttempt       int64            `json:"run_attempt"`
	WorkflowSHA      string           `json:"workflow_sha"`
	BranchProtection branchProtection `json:"branch_protection"`
	Jobs             []jobInput       `json:"jobs"`
	Artifacts        []artifactInput  `json:"artifacts"`
	Digests          receiptDigests   `json:"digests"`
	Cache            cacheReceipt     `json:"cache"`
	DiagnosticIDs    []string         `json:"diagnostic_ids"`
	RepairIDs        []string         `json:"repair_ids"`
	WriteEffect      string           `json:"write_effect"`
	Producer         string           `json:"producer"`
	Role             string           `json:"role"`
	Predecessors     []string         `json:"predecessors"`
	Decision         string           `json:"decision"`
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
