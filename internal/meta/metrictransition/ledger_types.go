package metrictransition

// StateReference identifies a measured or proof-derived repository state.
type StateReference struct {
	Schema          string `json:"schema"`
	Digest          string `json:"digest"`
	CommitSHA       string `json:"commit_sha"`
	Materialization string `json:"materialization"`
	BasisDigest     string `json:"basis_digest,omitempty"`
}

type ArtifactEvidence struct {
	Role   string `json:"role"`
	Digest string `json:"digest"`
}

type EffectEvidence struct {
	Verifier          string                    `json:"verifier"`
	Artifacts         []ArtifactEvidence        `json:"artifacts"`
	SetDigest         string                    `json:"set_digest"`
	Outcome           string                    `json:"outcome"`
	ReceiptDecision   string                    `json:"receipt_decision"`
	ReceiptCount      int                       `json:"receipt_count"`
	FailureCount      int                       `json:"failure_count"`
	UnknownCount      int                       `json:"unknown_count"`
	OperationEvidence []OperationEffectEvidence `json:"operation_evidence"`
}

type OperationEffectEvidence struct {
	ActionIndicatorID string `json:"action_indicator_id"`
	Operation         string `json:"operation"`
	Subject           string `json:"subject"`
	Executor          string `json:"executor"`
	Evaluator         string `json:"evaluator"`
	Status            string `json:"status"`
}

type MetricDelta struct {
	Source               Counts `json:"source"`
	Storage              Counts `json:"storage"`
	ChangedLanguageFiles int    `json:"changed_language_files"`
	ChangedDirectories   int    `json:"changed_directories"`
}

type Indicator struct {
	ID             string `json:"id"`
	Family         string `json:"family"`
	ProofChoice    string `json:"proof_choice"`
	Satisfied      bool   `json:"satisfied"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
}

type TransitionLedger struct {
	Schema              string         `json:"schema"`
	Status              string         `json:"status"`
	Decision            string         `json:"decision"`
	Reason              string         `json:"reason"`
	Repository          string         `json:"repository"`
	CommitSHA           string         `json:"commit_sha"`
	CIRunID             string         `json:"ci_run_id"`
	Before              StateReference `json:"before"`
	After               StateReference `json:"after"`
	Delta               MetricDelta    `json:"delta"`
	Effect              EffectEvidence `json:"effect"`
	RootPolicy          RootPolicy     `json:"root_policy"`
	Indicators          []Indicator    `json:"indicators"`
	PromotionAuthorized bool           `json:"promotion_authorized"`
	Digest              string         `json:"digest"`
}

type Result struct {
	State  RepositoryState
	Ledger TransitionLedger
}
