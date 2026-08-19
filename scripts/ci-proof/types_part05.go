package main

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
