package main

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
