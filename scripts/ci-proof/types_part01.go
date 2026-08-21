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
