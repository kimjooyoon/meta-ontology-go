package main

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
	Schema                 string                  `json:"schema"`
	Repository             string                  `json:"repository"`
	Event                  string                  `json:"event"`
	PRNumber               int64                   `json:"pr_number"`
	BaseRef                string                  `json:"base_ref"`
	BaseSHA                string                  `json:"base_sha"`
	HeadSHA                string                  `json:"head_sha"`
	Ref                    string                  `json:"ref"`
	EventRef               string                  `json:"event_ref"`
	CheckoutRef            string                  `json:"checkout_ref"`
	RunID                  int64                   `json:"run_id"`
	RunAttempt             int64                   `json:"run_attempt"`
	WorkflowSHA            string                  `json:"workflow_sha"`
	Jobs                   []jobInput              `json:"jobs"`
	Actors                 actorRoles              `json:"actors"`
	BranchProtection       branchProtection        `json:"branch_protection"`
	DevBranchProtection    branchProtection        `json:"dev_branch_protection"`
	DomainEvidence         domainEvidence          `json:"domain_evidence"`
	Scope                  scopeResult             `json:"scope"`
	Fixtures               fixtureResult           `json:"fixtures"`
	Artifacts              []artifactInput         `json:"artifacts"`
	Cache                  cacheInput              `json:"cache"`
	Digests                proofDigests            `json:"digests"`
	WriteEffect            string                  `json:"write_effect"`
	Decision               string                  `json:"decision"`
	NoWrite                bool                    `json:"no_write_outside_generated"`
	Rejections             []string                `json:"rejections"`
	Predecessors           []string                `json:"predecessors"`
	MissingReasons         missingReasons          `json:"missing_reasons"`
	GuardianEvidence       *guardianEvidence       `json:"guardian_evidence,omitempty"`
	PromotionObservation   *promotionObservation   `json:"promotion_observation,omitempty"`
	PromotionAuthorization *promotionAuthorization `json:"promotion_authorization,omitempty"`
}
type actorRoles struct {
	Actor   string `json:"actor"`
	Builder string `json:"builder"`
	Gate    string `json:"gate"`
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
