package nonmonotonicrefutationoracle

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type sourceClaim struct {
	ID             string `json:"id"`
	Predicate      string `json:"predicate"`
	ExpectedValue  string `json:"expected_value"`
	InitialStatus  string `json:"initial_status"`
	RevisionPolicy string `json:"revision_policy"`
}

type sourceObservation struct {
	ID             string     `json:"id"`
	Activity       string     `json:"activity"`
	ClaimID        string     `json:"claim_id"`
	Sequence       int        `json:"sequence"`
	Predicate      string     `json:"predicate"`
	ExpectedValue  string     `json:"expected_value"`
	ObservedValue  string     `json:"observed_value"`
	Provenance     string     `json:"provenance"`
	EvidenceDigest string     `json:"evidence_digest"`
	PriorState     string     `json:"prior_state"`
	RevisionPolicy string     `json:"revision_policy"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	MetaOperation  string     `json:"meta_operation"`
	ProofChoice    string     `json:"proof_choice"`
	Coordinate     coordinate `json:"coordinate"`
}

type sourceContract struct {
	Schema                string              `json:"schema"`
	FixedClaimTotal       int                 `json:"fixed_claim_total"`
	FixedObservationTotal int                 `json:"fixed_observation_total"`
	FixedTransitionTotal  int                 `json:"fixed_transition_total"`
	Claims                []sourceClaim       `json:"claims"`
	Observations          []sourceObservation `json:"observations"`
}

type sourceModel struct {
	Contract       sourceContract
	SemanticDigest string
}

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type producerInput struct {
	Schema               string         `json:"schema"`
	Contract             sourceContract `json:"contract"`
	SourcePath           string         `json:"source_path"`
	SourceDigest         string         `json:"source_digest"`
	SourceSemanticDigest string         `json:"source_semantic_digest"`
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
	EvidenceKind       string     `json:"evidence_kind"`
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
	InitialStatus      string   `json:"initial_status"`
	CurrentStatus      string   `json:"current_status"`
	StatusHistory      []string `json:"status_history"`
	HistoryRetained    bool     `json:"history_retained"`
	RefutationObserved bool     `json:"refutation_observed"`
	RevisionPolicy     string   `json:"revision_policy"`
	Producer           string   `json:"producer"`
	Consumer           string   `json:"consumer"`
	MetaOperation      string   `json:"meta_operation"`
	ProofChoice        string   `json:"proof_choice"`
}

type Metrics struct {
	FixedClaimTotal             int `json:"fixed_claim_total"`
	FixedObservationTotal       int `json:"fixed_observation_total"`
	FixedTransitionTotal        int `json:"fixed_transition_total"`
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
