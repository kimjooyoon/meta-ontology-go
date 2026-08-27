package nonmonotonicrefutationoracle

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type revisionPolicy struct {
	ID                          string `json:"id"`
	CorrectionRelation          string `json:"correction_relation"`
	CorrectionTarget            string `json:"correction_target"`
	UnknownAction               string `json:"unknown_action"`
	InsufficientAction          string `json:"insufficient_action"`
	OrdinarySupportAfterRefuted string `json:"ordinary_support_after_refuted"`
	FoundationRule              string `json:"foundation_rule"`
	CoherenceRule               string `json:"coherence_rule"`
	RegressionRule              string `json:"regression_rule"`
	FixtureClass                string `json:"fixture_class"`
	PolicyDigest                string `json:"policy_digest"`
}

type sourceClaim struct {
	ID            string `json:"id"`
	Proposition   string `json:"proposition"`
	Subject       string `json:"subject"`
	Input         string `json:"input"`
	Predicate     string `json:"predicate"`
	ExpectedValue string `json:"expected_value"`
}

type evidenceMaterial struct {
	ClaimID                  string `json:"claim_id"`
	Proposition              string `json:"proposition"`
	TargetAddress            string `json:"target_address"`
	ObservedMaterial         string `json:"observed_material"`
	ObservedValue            string `json:"observed_value"`
	ObservationQuality       string `json:"observation_quality"`
	ProviderClass            string `json:"provider_class"`
	Sequence                 int    `json:"sequence"`
	SupersededEvidenceDigest string `json:"superseded_evidence_digest"`
	SupersededClaimID        string `json:"superseded_claim_id"`
}

type sourceObservation struct {
	ID                       string     `json:"id"`
	Activity                 string     `json:"activity"`
	ClaimID                  string     `json:"claim_id"`
	Sequence                 int        `json:"sequence"`
	Proposition              string     `json:"proposition"`
	Subject                  string     `json:"subject"`
	Input                    string     `json:"input"`
	Predicate                string     `json:"predicate"`
	ExpectedValue            string     `json:"expected_value"`
	ObservedValue            string     `json:"observed_value"`
	ObservedMaterial         string     `json:"observed_material"`
	ObservationQuality       string     `json:"observation_quality"`
	ProviderClass            string     `json:"provider_class"`
	Provenance               string     `json:"provenance"`
	EvidenceDigest           string     `json:"evidence_digest"`
	RevisionRelation         string     `json:"revision_relation"`
	SupersedesEvidenceDigest string     `json:"supersedes_evidence_digest"`
	SupersedesClaimID        string     `json:"supersedes_claim_id"`
	PolicyID                 string     `json:"policy_id"`
	PolicyDigest             string     `json:"policy_digest"`
	Producer                 string     `json:"producer"`
	Consumer                 string     `json:"consumer"`
	MetaOperation            string     `json:"meta_operation"`
	ProofChoice              string     `json:"proof_choice"`
	Coordinate               coordinate `json:"coordinate"`
	TargetAddress            string     `json:"target_address"`
}

type sourceContract struct {
	Schema                string              `json:"schema"`
	Policy                revisionPolicy      `json:"policy"`
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
	PolicyID       string `json:"policy_id"`
	PolicyDigest   string `json:"policy_digest"`
}

type effects struct {
	NetRepositoryStatusUnchanged bool   `json:"net_repository_status_unchanged"`
	RepositoryWriteObservation   string `json:"repository_write_observation"`
	MutationAuthorityResolution  string `json:"mutation_authority_resolution"`
	PromotionOperationsObserved  int    `json:"promotion_operations_observed"`
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
	Sequence                         int        `json:"sequence"`
	CaseID                           string     `json:"case_id"`
	ClaimID                          string     `json:"claim_id"`
	Before                           string     `json:"before"`
	After                            string     `json:"after"`
	Accepted                         bool       `json:"accepted"`
	EvidenceID                       string     `json:"evidence_id"`
	Relation                         string     `json:"relation"`
	RevisionRelation                 string     `json:"revision_relation"`
	SupersedesEvidenceDigest         string     `json:"supersedes_evidence_digest"`
	CorrectionTargetClaimID          string     `json:"correction_target_claim_id"`
	CorrectionTargetTransitionDigest string     `json:"correction_target_transition_digest"`
	CorrectionTargetStatus           string     `json:"correction_target_status"`
	CorrectionTargetActive           bool       `json:"correction_target_active"`
	EvidenceBasis                    string     `json:"evidence_basis"`
	EvidenceDigest                   string     `json:"evidence_digest"`
	EvidenceProvenance               string     `json:"evidence_provenance"`
	ProviderClass                    string     `json:"provider_class"`
	ProofChoice                      string     `json:"proof_choice"`
	ProofAdmitted                    bool       `json:"proof_admitted"`
	ProofAdmission                   string     `json:"proof_admission"`
	Coordinate                       coordinate `json:"coordinate"`
	PreviousDigest                   string     `json:"previous_digest,omitempty"`
	TransitionDigest                 string     `json:"transition_digest"`
}

type CaseResult struct {
	ID                       string   `json:"id"`
	ClaimID                  string   `json:"claim_id"`
	Proposition              string   `json:"proposition"`
	Subject                  string   `json:"subject"`
	Input                    string   `json:"input"`
	FixtureKnowledge         string   `json:"fixture_knowledge"`
	CurrentEvidenceID        string   `json:"current_evidence_id"`
	CurrentEvidenceDigest    string   `json:"current_evidence_digest"`
	InitialStatus            string   `json:"initial_status"`
	CurrentStatus            string   `json:"current_status"`
	StatusHistory            []string `json:"status_history"`
	HistoryRetained          bool     `json:"history_retained"`
	ObservationTotal         int      `json:"observation_total"`
	RejectedObservationTotal int      `json:"rejected_observation_total"`
	RefutationObserved       bool     `json:"refutation_observed"`
}

type Metrics struct {
	FixedCaseTotal               int `json:"fixed_case_total"`
	FixedClaimTotal              int `json:"fixed_claim_total"`
	FixedObservationTotal        int `json:"fixed_observation_total"`
	FixedLedgerRowTotal          int `json:"fixed_ledger_row_total"`
	InScopeClaimTotal            int `json:"in_scope_claim_total"`
	ObservationAttemptTotal      int `json:"observation_attempt_total"`
	TransitionTotal              int `json:"transition_total"`
	AcceptedStateTransitionTotal int `json:"accepted_state_transition_total"`
	RejectedObservationTotal     int `json:"rejected_observation_total"`
	SupportsTotal                int `json:"supports_total"`
	ContradictsTotal             int `json:"contradicts_total"`
	InsufficientTotal            int `json:"insufficient_total"`
	UnknownTotal                 int `json:"unknown_total"`
	OpenToDischargedTotal        int `json:"open_to_discharged_total"`
	DischargedToRefutedTotal     int `json:"discharged_to_refuted_total"`
	RefutedToDischargedTotal     int `json:"refuted_to_discharged_total"`
	CurrentDischargedTotal       int `json:"current_discharged_total"`
	CurrentRefutedTotal          int `json:"current_refuted_total"`
	CurrentOpenTotal             int `json:"current_open_total"`
	RetainedStateTotal           int `json:"retained_state_total"`
	NonMonotonicRevisionTotal    int `json:"non_monotonic_revision_total"`
	CurrentDischargeBasisPoints  int `json:"current_discharge_basis_points"`
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

type Vocabulary struct {
	FixtureKnowledge string `json:"fixture_knowledge"`
	CurrentEvidence  string `json:"current_evidence"`
	UnknownEvidence  string `json:"unknown_evidence"`
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
	Policy                revisionPolicy    `json:"policy"`
	Vocabulary            Vocabulary        `json:"vocabulary"`
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
