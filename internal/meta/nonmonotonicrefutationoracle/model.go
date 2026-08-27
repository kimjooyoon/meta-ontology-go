package nonmonotonicrefutationoracle

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type evidence struct {
	ID            string     `json:"id"`
	ClaimID       string     `json:"claim_id"`
	Kind          string     `json:"kind"`
	Basis         string     `json:"basis"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    coordinate `json:"coordinate"`
}

type caseDefinition struct {
	ID                  string     `json:"id"`
	ClaimID             string     `json:"claim_id"`
	InitialStatus       string     `json:"initial_status"`
	ExpectedFinalStatus string     `json:"expected_final_status"`
	Producer            string     `json:"producer"`
	Consumer            string     `json:"consumer"`
	MetaOperation       string     `json:"meta_operation"`
	ProofChoice         string     `json:"proof_choice"`
	Evidence            []evidence `json:"evidence"`
}

type contract struct {
	Schema               string           `json:"schema"`
	FixedClaimTotal      int              `json:"fixed_claim_total"`
	FixedTransitionTotal int              `json:"fixed_transition_total"`
	Cases                []caseDefinition `json:"cases"`
}

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type producerInput struct {
	Schema        string   `json:"schema"`
	Contract      contract `json:"contract"`
	SourcePath    string   `json:"source_path"`
	SourceDigest  string   `json:"source_digest"`
	Producer      string   `json:"producer"`
	Consumer      string   `json:"consumer"`
	MetaOperation string   `json:"meta_operation"`
	ProofChoice   string   `json:"proof_choice"`
	Effects       effects  `json:"effects"`
	NotClaimed    []string `json:"not_claimed"`
	ReceiptDigest string   `json:"receipt_digest"`
}

type Transition struct {
	Sequence         int        `json:"sequence"`
	CaseID           string     `json:"case_id"`
	ClaimID          string     `json:"claim_id"`
	Before           string     `json:"before"`
	After            string     `json:"after"`
	EvidenceID       string     `json:"evidence_id"`
	EvidenceKind     string     `json:"evidence_kind"`
	Coordinate       coordinate `json:"coordinate"`
	PreviousDigest   string     `json:"previous_digest,omitempty"`
	TransitionDigest string     `json:"transition_digest"`
}

type CaseResult struct {
	ID                  string   `json:"id"`
	ClaimID             string   `json:"claim_id"`
	InitialStatus       string   `json:"initial_status"`
	ExpectedFinalStatus string   `json:"expected_final_status"`
	CurrentStatus       string   `json:"current_status"`
	StatusHistory       []string `json:"status_history"`
	HistoryRetained     bool     `json:"history_retained"`
	RefutationObserved  bool     `json:"refutation_observed"`
	Producer            string   `json:"producer"`
	Consumer            string   `json:"consumer"`
	MetaOperation       string   `json:"meta_operation"`
	ProofChoice         string   `json:"proof_choice"`
}

type Metrics struct {
	FixedClaimTotal             int `json:"fixed_claim_total"`
	InScopeClaimTotal           int `json:"in_scope_claim_total"`
	TransitionTotal             int `json:"transition_total"`
	OpenToDischargedTotal       int `json:"open_to_discharged_total"`
	DischargedToRefutedTotal    int `json:"discharged_to_refuted_total"`
	RefutedToDischargedTotal    int `json:"refuted_to_discharged_total"`
	CurrentDischargedTotal      int `json:"current_discharged_total"`
	CurrentRefutedTotal         int `json:"current_refuted_total"`
	RetainedStateTotal          int `json:"retained_state_total"`
	NonMonotonicRevisionTotal   int `json:"non_monotonic_revision_total"`
	CurrentDischargeBasisPoints int `json:"current_discharge_basis_points"`
}

type Report struct {
	Schema                string       `json:"schema"`
	ProducerSchema        string       `json:"producer_schema"`
	ProducerReceiptDigest string       `json:"producer_receipt_digest"`
	SourcePath            string       `json:"source_path"`
	SourceDigest          string       `json:"source_digest"`
	Producer              string       `json:"producer"`
	Consumer              string       `json:"consumer"`
	MetaOperation         string       `json:"meta_operation"`
	ProofChoice           string       `json:"proof_choice"`
	Decision              string       `json:"decision"`
	Resolution            string       `json:"resolution"`
	Reason                string       `json:"reason"`
	MetaValue             string       `json:"meta_value"`
	FalsifiablePrediction string       `json:"falsifiable_prediction"`
	Effects               effects      `json:"effects"`
	Metrics               Metrics      `json:"metrics"`
	Cases                 []CaseResult `json:"cases"`
	Transitions           []Transition `json:"transitions"`
	ReportDigest          string       `json:"report_digest"`
}
