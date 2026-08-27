package nonmonotonicrefutationoracle

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type sourceClaim struct {
	ID            string `json:"id"`
	Proposition   string `json:"proposition"`
	Subject       string `json:"subject"`
	Input         string `json:"input"`
	Predicate     string `json:"predicate"`
	ExpectedValue string `json:"expected_value"`
}

type sourceObservation struct {
	ID             string     `json:"id"`
	Activity       string     `json:"activity"`
	ClaimID        string     `json:"claim_id"`
	Sequence       int        `json:"sequence"`
	Proposition    string     `json:"proposition"`
	Subject        string     `json:"subject"`
	Input          string     `json:"input"`
	Predicate      string     `json:"predicate"`
	ExpectedValue  string     `json:"expected_value"`
	ObservedValue  string     `json:"observed_value"`
	Provenance     string     `json:"provenance"`
	EvidenceDigest string     `json:"evidence_digest"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	MetaOperation  string     `json:"meta_operation"`
	ProofChoice    string     `json:"proof_choice"`
	Coordinate     coordinate `json:"coordinate"`
}

type sourceContract struct {
	Schema                string              `json:"schema"`
	FixedCaseTotal        int                 `json:"fixed_case_total"`
	FixedClaimTotal       int                 `json:"fixed_claim_total"`
	FixedObservationTotal int                 `json:"fixed_observation_total"`
	FixedLedgerRowTotal   int                 `json:"fixed_ledger_row_total"`
	Claims                []sourceClaim       `json:"claims"`
	Observations          []sourceObservation `json:"observations"`
}

type sourceModel struct {
	Contract       sourceContract
	SemanticDigest string
}

type sourceBinding struct {
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
	PromotionCount    int  `json:"promotion_count"`
}

type producerInput struct {
	Schema               string         `json:"schema"`
	Contract             sourceContract `json:"contract"`
	SourcePath           string         `json:"source_path"`
	SourceDigest         string         `json:"source_digest"`
	SourceSemanticDigest string         `json:"source_semantic_digest"`
	SourceBindingDigest  string         `json:"source_binding_digest"`
	SourceModelDigest    string         `json:"source_model_digest"`
	Producer             string         `json:"producer"`
	Consumer             string         `json:"consumer"`
	MetaOperation        string         `json:"meta_operation"`
	ProofChoice          string         `json:"proof_choice"`
	Effects              effects        `json:"effects"`
	NotClaimed           []string       `json:"not_claimed"`
	ReceiptDigest        string         `json:"receipt_digest"`
}

type Transition struct {
	Sequence           int        `json:"sequence"`
	CaseID             string     `json:"case_id"`
	ClaimID            string     `json:"claim_id"`
	Before             string     `json:"before"`
	After              string     `json:"after"`
	Accepted           bool       `json:"accepted"`
	EvidenceID         string     `json:"evidence_id"`
	Relation           string     `json:"relation"`
	EvidenceBasis      string     `json:"evidence_basis"`
	EvidenceDigest     string     `json:"evidence_digest"`
	EvidenceProvenance string     `json:"evidence_provenance"`
	ProofChoice        string     `json:"proof_choice"`
	Coordinate         coordinate `json:"coordinate"`
	PreviousDigest     string     `json:"previous_digest,omitempty"`
	TransitionDigest   string     `json:"transition_digest"`
}

type CaseResult struct {
	ID                 string   `json:"id"`
	ClaimID            string   `json:"claim_id"`
	Proposition        string   `json:"proposition"`
	Subject            string   `json:"subject"`
	Input              string   `json:"input"`
	InitialStatus      string   `json:"initial_status"`
	CurrentStatus      string   `json:"current_status"`
	StatusHistory      []string `json:"status_history"`
	HistoryRetained    bool     `json:"history_retained"`
	ObservationTotal   int      `json:"observation_total"`
	RefutationObserved bool     `json:"refutation_observed"`
}

type Metrics struct {
	FixedCaseTotal              int `json:"fixed_case_total"`
	FixedClaimTotal             int `json:"fixed_claim_total"`
	FixedObservationTotal       int `json:"fixed_observation_total"`
	FixedLedgerRowTotal         int `json:"fixed_ledger_row_total"`
	InScopeClaimTotal           int `json:"in_scope_claim_total"`
	TransitionTotal             int `json:"transition_total"`
	SupportsTotal               int `json:"supports_total"`
	ContradictsTotal            int `json:"contradicts_total"`
	InsufficientTotal           int `json:"insufficient_total"`
	UnknownTotal                int `json:"unknown_total"`
	OpenToDischargedTotal       int `json:"open_to_discharged_total"`
	DischargedToRefutedTotal    int `json:"discharged_to_refuted_total"`
	RefutedToDischargedTotal    int `json:"refuted_to_discharged_total"`
	CurrentDischargedTotal      int `json:"current_discharged_total"`
	CurrentRefutedTotal         int `json:"current_refuted_total"`
	CurrentOpenTotal            int `json:"current_open_total"`
	RetainedStateTotal          int `json:"retained_state_total"`
	NonMonotonicRevisionTotal   int `json:"non_monotonic_revision_total"`
	CurrentDischargeBasisPoints int `json:"current_discharge_basis_points"`
}

type Conformance struct {
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type SubjectResolution struct {
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type Report struct {
	Schema                string            `json:"schema"`
	ProducerSchema        string            `json:"producer_schema"`
	ProducerReceiptDigest string            `json:"producer_receipt_digest"`
	SourcePath            string            `json:"source_path"`
	SourceDigest          string            `json:"source_digest"`
	SourceSemanticDigest  string            `json:"source_semantic_digest"`
	SourceBindingDigest   string            `json:"source_binding_digest"`
	SourceModelDigest     string            `json:"source_model_digest"`
	Producer              string            `json:"producer"`
	Consumer              string            `json:"consumer"`
	MetaOperation         string            `json:"meta_operation"`
	ProofChoice           string            `json:"proof_choice"`
	Decision              string            `json:"decision"`
	Resolution            string            `json:"resolution"`
	Reason                string            `json:"reason"`
	Conformance           Conformance       `json:"conformance"`
	SubjectResolution     SubjectResolution `json:"subject_resolution"`
	MetaValue             string            `json:"meta_value"`
	FalsifiablePrediction string            `json:"falsifiable_prediction"`
	Effects               effects           `json:"effects"`
	Metrics               Metrics           `json:"metrics"`
	Cases                 []CaseResult      `json:"cases"`
	Transitions           []Transition      `json:"transitions"`
	ReportDigest          string            `json:"report_digest"`
}
