package main

const failureSchema = "gooo/ci-failure/v1"
const failureCatalogPath = "scripts/ci-proof/docs/failure-reasons.md"

type failureJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}
type failureInput struct {
	Code                 string          `json:"code"`
	FailureCodes         []string        `json:"failure_codes"`
	Message              string          `json:"message"`
	Remediation          string          `json:"remediation"`
	OwnerBranch          string          `json:"owner_branch"`
	Rejections           []string        `json:"rejections"`
	MissingReasons       missingReasons  `json:"missing_reasons"`
	Artifacts            []artifactInput `json:"artifacts"`
	ProofArtifact        *artifactInput  `json:"proof_artifact"`
	ArtifactStatus       string          `json:"artifact_status"`
	ArtifactReason       string          `json:"artifact_reason"`
	TerminalFailures     []failureJob    `json:"terminal_failures"`
	TerminalFailureCodes []string        `json:"terminal_failure_codes"`
	Job                  failureJob      `json:"job"`
}
type failureBinding struct {
	Repository  string
	Event       string
	EventRef    string
	CheckoutRef string
	BaseRef     string
	BaseSHA     string
	HeadSHA     string
	WorkflowSHA string
	PRNumber    int64
	RunID       int64
	RunAttempt  int64
	Actor       string
	OwnerBranch string
}
type failureProvenance struct {
	WasGeneratedBy    string   `json:"wasGeneratedBy"`
	WasAssociatedWith string   `json:"wasAssociatedWith"`
	WasDerivedFrom    []string `json:"wasDerivedFrom"`
	HadPrimarySource  []string `json:"hadPrimarySource"`
}
