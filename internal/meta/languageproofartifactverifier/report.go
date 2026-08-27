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
	ID                 string        `json:"id"`
	Status             string        `json:"status"`
	ExpectedDecision   string        `json:"expected_decision"`
	ExpectedResolution string        `json:"expected_resolution"`
	ExpectedReason     string        `json:"expected_reason"`
	ObservedDecision   string        `json:"observed_decision"`
	ObservedResolution string        `json:"observed_resolution"`
	ObservedReason     string        `json:"observed_reason"`
	ProofChoice        string        `json:"proof_choice"`
	MetaOperation      string        `json:"meta_operation"`
	Coordinate         Coordinate    `json:"coordinate"`
	Claims             []ClaimResult `json:"claims"`
	ArtifactDigest     string        `json:"artifact_digest"`
	SourceDigest       string        `json:"source_digest"`
	SemanticDigest     string        `json:"semantic_digest"`
	OperationDigest    string        `json:"operation_digest"`
}

type Summary struct {
	CasesSatisfied              int `json:"cases_satisfied"`
	CasesTotal                  int `json:"cases_total"`
	ValidArtifacts              int `json:"valid_artifacts"`
	EvidenceKindsCarried        int `json:"evidence_kinds_carried"`
	ExactEvidenceLinks          int `json:"exact_evidence_links"`
	RecipeMatches               int `json:"recipe_matches"`
	PreservedTransitions        int `json:"preserved_transitions"`
	TransitionTotal             int `json:"transition_total"`
	TamperedRejections          int `json:"tampered_rejections"`
	CoherentTamperRejections    int `json:"coherent_tamper_rejections"`
	MissingEvidenceRejections   int `json:"missing_evidence_rejections"`
	ByteOnlyDenials             int `json:"byte_only_denials"`
	RecipeRejections            int `json:"recipe_rejections"`
	RecipeOnlyRejections        int `json:"recipe_only_rejections"`
	MissingAttachmentRejections int `json:"missing_attachment_rejections"`
	WrongAttachmentRejections   int `json:"wrong_attachment_rejections"`
	UnrelatedEvidenceRejections int `json:"unrelated_evidence_rejections"`
	StaleHeadRejections         int `json:"stale_head_rejections"`
	UnauthorizedConsumerDenials int `json:"unauthorized_consumer_denials"`
	SemanticInterventions       int `json:"semantic_interventions"`
	NonsemanticInterventions    int `json:"nonsemantic_interventions"`
	ReadOnlyAuthorities         int `json:"read_only_authorities"`
	ProducerDependencies        int `json:"producer_dependencies"`
	ProducerImportNumerator     int `json:"producer_import_numerator"`
	ProducerImportDenominator   int `json:"producer_import_denominator"`
	CoreParserDependencies      int `json:"core_parser_dependencies"`
	ClaimTemplates              int `json:"claim_templates"`
	ClaimInstances              int `json:"claim_instances"`
	AcceptedTransitions         int `json:"accepted_transitions"`
	CaseDischargedClaims        int `json:"case_discharged_claims"`
	CaseOpenClaims              int `json:"case_open_claims"`
	CaseRefutedClaims           int `json:"case_refuted_claims"`
	FinalLedgerOpenClaims       int `json:"final_ledger_open_claims"`
	FinalLedgerDischargedClaims int `json:"final_ledger_discharged_claims"`
	NetRepositoryStateUnchanged int `json:"net_repository_state_unchanged"`
	BundleOnlyVerification      int `json:"bundle_only_verification"`
	ConsumerRechecks            int `json:"consumer_rechecks"`
	GeneratedAuthority          int `json:"generated_authority"`
	SemanticClaims              int `json:"semantic_claims"`
	RepositoryWrites            int `json:"repository_writes"`
	MutationAuthorities         int `json:"mutation_authorities"`
	PromotionAuthorities        int `json:"promotion_authorities"`
	SemanticAuthorities         int `json:"semantic_authorities"`
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
	Choice          string   `json:"choice"`
	MetaOperation   string   `json:"meta_operation"`
	TargetDigest    string   `json:"target_digest"`
	Dependency      string   `json:"dependency"`
	EvidenceDigests []string `json:"evidence_digests"`
	ReceiptDigest   string   `json:"receipt_digest"`
	Passed          bool     `json:"passed"`
}

type Report struct {
	Schema                    string               `json:"schema"`
	HeadSHA                   string               `json:"head_sha"`
	Producer                  string               `json:"producer"`
	Consumer                  string               `json:"consumer"`
	ConformanceDecision       string               `json:"conformance_decision"`
	ConformanceResolution     string               `json:"conformance_resolution"`
	ConformanceReason         string               `json:"conformance_reason"`
	SubjectArtifactDecision   string               `json:"subject_artifact_decision"`
	SubjectArtifactResolution string               `json:"subject_artifact_resolution"`
	SubjectArtifactReason     string               `json:"subject_artifact_reason"`
	ContractDigest            string               `json:"contract_digest"`
	RecipeDigest              string               `json:"recipe_digest"`
	RecipeVersion             int                  `json:"recipe_version"`
	IndependenceDigest        string               `json:"independence_digest"`
	ArtifactUseAuthority      string               `json:"artifact_use_authority"`
	AuthorityObservation      string               `json:"authority_observation"`
	BundleDigest              string               `json:"bundle_digest"`
	Checkout                  CheckoutEvidence     `json:"checkout"`
	ConsumerReceipt           ConsumerReceipt      `json:"consumer_receipt"`
	Cases                     []CaseResult         `json:"cases"`
	Summary                   Summary              `json:"summary"`
	Indicators                []Indicator          `json:"indicators"`
	Proofs                    []Proof              `json:"proofs"`
	Transitions               []ClaimTransition    `json:"claim_transitions"`
	PriorLedger               ClaimLedger          `json:"prior_ledger"`
	Ledger                    ClaimLedger          `json:"ledger"`
	WriteSet                  WriteSetObservation  `json:"write_set"`
	Interventions             []InterventionResult `json:"interventions"`
	RepositoryWrites          int                  `json:"repository_writes"`
	MutationAuthority         bool                 `json:"mutation_authority"`
	PromotionAuthority        bool                 `json:"promotion_authority"`
	SemanticAuthority         bool                 `json:"semantic_authority"`
	NotClaimed                []string             `json:"not_claimed"`
	Digest                    string               `json:"digest"`
}
