package languageproofartifactverifier

type ClaimResult struct {
	ID              string     `json:"id"`
	Proposition     string     `json:"proposition"`
	TargetDigest    string     `json:"target_digest"`
	Dependencies    []string   `json:"dependencies"`
	Status          string     `json:"status"`
	Resolution      string     `json:"resolution"`
	Reason          string     `json:"reason"`
	ProofChoice     string     `json:"proof_choice"`
	MetaOperation   string     `json:"meta_operation"`
	Coordinate      Coordinate `json:"coordinate"`
	EvidenceDigest  string     `json:"evidence_digest"`
	EvidenceDigests []string   `json:"evidence_digests"`
	StateDigest     string     `json:"state_digest"`
	Provenance      string     `json:"provenance"`
}

type CaseResult struct {
	ID                        string        `json:"id"`
	Status                    string        `json:"status"`
	ExpectedDecision          string        `json:"expected_decision"`
	ExpectedResolution        string        `json:"expected_resolution"`
	ExpectedReason            string        `json:"expected_reason"`
	ObservedDecision          string        `json:"observed_decision"`
	ObservedResolution        string        `json:"observed_resolution"`
	ObservedReason            string        `json:"observed_reason"`
	ProofChoice               string        `json:"proof_choice"`
	MetaOperation             string        `json:"meta_operation"`
	Coordinate                Coordinate    `json:"coordinate"`
	Claims                    []ClaimResult `json:"claims"`
	ArtifactDigest            string        `json:"artifact_digest"`
	SourceDigest              string        `json:"source_digest"`
	SemanticDigest            string        `json:"semantic_digest"`
	OperationDigest           string        `json:"operation_digest"`
	OperationAttachmentDigest string        `json:"operation_attachment_digest,omitempty"`
	RecipeAttachmentDigest    string        `json:"recipe_attachment_digest,omitempty"`
	ConsumerTargetDigest      string        `json:"consumer_target_digest,omitempty"`
	ConsumerOutputDigest      string        `json:"consumer_output_digest,omitempty"`
	ConsumerOutputExists      bool          `json:"consumer_output_exists"`
	ConsumerErrorClass        string        `json:"consumer_error_class,omitempty"`
	ConsumerErrorDigest       string        `json:"consumer_error_digest,omitempty"`
	EnvelopeDigest            string        `json:"envelope_digest"`
}

type Summary struct {
	CasesSatisfied                   int `json:"cases_satisfied"`
	CasesTotal                       int `json:"cases_total"`
	ValidArtifacts                   int `json:"valid_artifacts"`
	EvidenceKindsCarried             int `json:"evidence_kinds_carried"`
	ExactEvidenceLinks               int `json:"exact_evidence_links"`
	RecipeMatches                    int `json:"recipe_matches"`
	PreservedTransitions             int `json:"preserved_transitions"`
	TransitionTotal                  int `json:"transition_total"`
	TamperedRejections               int `json:"tampered_rejections"`
	CoherentTamperRejections         int `json:"coherent_tamper_rejections"`
	MissingEvidenceRejections        int `json:"missing_evidence_rejections"`
	ByteOnlyDenials                  int `json:"byte_only_denials"`
	RecipeRejections                 int `json:"recipe_rejections"`
	RecipeOnlyRejections             int `json:"recipe_only_rejections"`
	MissingAttachmentRejections      int `json:"missing_attachment_rejections"`
	WrongAttachmentRejections        int `json:"wrong_attachment_rejections"`
	UnrelatedEvidenceRejections      int `json:"unrelated_evidence_rejections"`
	StaleHeadRejections              int `json:"stale_head_rejections"`
	UnauthorizedConsumerDenials      int `json:"unauthorized_consumer_denials"`
	CoherentClaimStructureRejections int `json:"coherent_claim_structure_rejections"`
	SemanticInterventions            int `json:"semantic_interventions"`
	NonsemanticInterventions         int `json:"nonsemantic_interventions"`
	ReadOnlyAuthorities              int `json:"read_only_authorities"`
	ProducerDependencies             int `json:"producer_dependencies"`
	ProducerImportNumerator          int `json:"producer_import_numerator"`
	ProducerImportDenominator        int `json:"producer_import_denominator"`
	CoreParserDependencies           int `json:"core_parser_dependencies"`
	ClaimTemplates                   int `json:"claim_templates"`
	ClaimInstances                   int `json:"claim_instances"`
	AcceptedTransitions              int `json:"accepted_transitions"`
	CaseDischargedClaims             int `json:"case_discharged_claims"`
	CaseOpenClaims                   int `json:"case_open_claims"`
	CaseRefutedClaims                int `json:"case_refuted_claims"`
	FinalLedgerOpenClaims            int `json:"final_ledger_open_claims"`
	FinalLedgerDischargedClaims      int `json:"final_ledger_discharged_claims"`
	NetRepositoryStateUnchanged      int `json:"net_repository_state_unchanged"`
	UnknownAuthorityObservations     int `json:"unknown_authority_observations"`
	BundleOnlyVerification           int `json:"bundle_only_verification"`
	ConsumerRechecks                 int `json:"consumer_rechecks"`
	GeneratedAuthority               int `json:"generated_authority"`
	SemanticClaims                   int `json:"semantic_claims"`
	NetChangedPaths                  int `json:"net_changed_paths"`
	MutationAuthorities              int `json:"mutation_authorities"`
	PromotionAuthorities             int `json:"promotion_authorities"`
	SemanticAuthorities              int `json:"semantic_authorities"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Phase             string   `json:"phase"`
	State             string   `json:"state"`
	EvidenceValidated bool     `json:"evidence_validated"`
	Choice            string   `json:"choice"`
	MetaOperation     string   `json:"meta_operation"`
	TargetDigest      string   `json:"target_digest"`
	Dependency        string   `json:"dependency"`
	EvidenceDigests   []string `json:"evidence_digests"`
	ReceiptDigest     string   `json:"receipt_digest"`
	Passed            bool     `json:"passed"`
	ConsumerGateOpen  bool     `json:"consumer_gate_open"`
}

type Report struct {
	Schema                           string               `json:"schema"`
	HeadSHA                          string               `json:"head_sha"`
	Producer                         string               `json:"producer"`
	Consumer                         string               `json:"consumer"`
	ConformanceDecision              string               `json:"conformance_decision"`
	ConformanceResolution            string               `json:"conformance_resolution"`
	ConformanceReason                string               `json:"conformance_reason"`
	ConformanceCoordinate            Coordinate           `json:"conformance_coordinate"`
	PreliminaryDecision              string               `json:"preliminary_decision,omitempty"`
	PreliminaryResolution            string               `json:"preliminary_resolution,omitempty"`
	PreliminaryReason                string               `json:"preliminary_reason,omitempty"`
	PreliminaryCoordinate            Coordinate           `json:"preliminary_coordinate,omitempty"`
	SubjectArtifactDecision          string               `json:"subject_artifact_decision"`
	SubjectArtifactResolution        string               `json:"subject_artifact_resolution"`
	SubjectArtifactReason            string               `json:"subject_artifact_reason"`
	ContractDigest                   string               `json:"contract_digest"`
	RecipeDigest                     string               `json:"recipe_digest"`
	RecipeVersion                    int                  `json:"recipe_version"`
	IndependenceDigest               string               `json:"independence_digest"`
	ArtifactUseAuthority             string               `json:"artifact_use_authority"`
	AuthorityObservation             string               `json:"authority_observation"`
	BundleDigest                     string               `json:"bundle_digest"`
	Checkout                         CheckoutEvidence     `json:"checkout"`
	ConsumerReceipt                  ConsumerReceipt      `json:"consumer_receipt"`
	CheckoutBindingScope             string               `json:"checkout_binding_scope"`
	UnauthorizedConsumerTargetDigest string               `json:"unauthorized_consumer_target_digest"`
	UnauthorizedConsumerOutputDigest string               `json:"unauthorized_consumer_output_digest"`
	UnauthorizedConsumerOutputExists bool                 `json:"unauthorized_consumer_output_exists"`
	UnauthorizedConsumerErrorClass   string               `json:"unauthorized_consumer_error_class"`
	UnauthorizedConsumerErrorDigest  string               `json:"unauthorized_consumer_error_digest"`
	Counterexamples                  []Counterexample     `json:"counterexamples"`
	Cases                            []CaseResult         `json:"cases"`
	Summary                          Summary              `json:"summary"`
	Indicators                       []Indicator          `json:"indicators"`
	Proofs                           []Proof              `json:"proofs"`
	ProofSummary                     ProofSummary         `json:"proof_summary"`
	Transitions                      []ClaimTransition    `json:"claim_transitions"`
	PriorLedger                      ClaimLedger          `json:"prior_ledger"`
	Ledger                           ClaimLedger          `json:"ledger"`
	WriteSet                         WriteSetObservation  `json:"write_set"`
	Interventions                    []InterventionResult `json:"interventions"`
	NetChangedPaths                  int                  `json:"net_changed_paths"`
	CapabilityMutationGranted        bool                 `json:"capability_mutation_granted"`
	PromotionAuthority               bool                 `json:"promotion_authority"`
	SemanticAuthority                bool                 `json:"semantic_authority"`
	NotClaimed                       []string             `json:"not_claimed"`
	Digest                           string               `json:"digest"`
	ValidationFailure                *ValidationFailure   `json:"validation_failure,omitempty"`
}

type ValidationFailure struct {
	Coordinate Coordinate `json:"coordinate"`
	Detail     string     `json:"detail"`
}

type ProofSummary struct {
	Phase                  string `json:"phase"`
	Proofs                 int    `json:"proofs"`
	EvidenceValidated      int    `json:"evidence_validated"`
	EvidenceValidatedTotal int    `json:"evidence_validated_total"`
	ObservedState          int    `json:"observed_state"`
	ObservedStateTotal     int    `json:"observed_state_total"`
	Open                   int    `json:"open"`
	OpenTotal              int    `json:"open_total"`
	Discharged             int    `json:"discharged"`
	DischargedTotal        int    `json:"discharged_total"`
	Authority              int    `json:"authority"`
	AuthorityTotal         int    `json:"authority_total"`
}
