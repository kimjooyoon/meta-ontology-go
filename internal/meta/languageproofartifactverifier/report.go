package languageproofartifactverifier

type ClaimResult struct {
	ID             string     `json:"id"`
	Status         string     `json:"status"`
	Resolution     string     `json:"resolution"`
	Reason         string     `json:"reason"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
	Provenance     string     `json:"provenance"`
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
	CasesSatisfied            int `json:"cases_satisfied"`
	CasesTotal                int `json:"cases_total"`
	ValidArtifacts            int `json:"valid_artifacts"`
	EvidenceKindsCarried      int `json:"evidence_kinds_carried"`
	ExactEvidenceLinks        int `json:"exact_evidence_links"`
	RecipeMatches             int `json:"recipe_matches"`
	PreservedTransitions      int `json:"preserved_transitions"`
	TransitionTotal           int `json:"transition_total"`
	TamperedRejections        int `json:"tampered_rejections"`
	CoherentTamperRejections  int `json:"coherent_tamper_rejections"`
	MissingEvidenceRejections int `json:"missing_evidence_rejections"`
	ByteOnlyDenials           int `json:"byte_only_denials"`
	RecipeRejections          int `json:"recipe_rejections"`
	LedgerDischargedClaims    int `json:"ledger_discharged_claims"`
	LedgerOpenClaims          int `json:"ledger_open_claims"`
	LedgerRefutedClaims       int `json:"ledger_refuted_claims"`
	SemanticInterventions     int `json:"semantic_interventions"`
	NonsemanticInterventions  int `json:"nonsemantic_interventions"`
	ReadOnlyAuthorities       int `json:"read_only_authorities"`
	ProducerDependencies      int `json:"producer_dependencies"`
	ProducerImportNumerator   int `json:"producer_import_numerator"`
	ProducerImportDenominator int `json:"producer_import_denominator"`
	CoreParserDependencies    int `json:"core_parser_dependencies"`
	GeneratedAuthority        int `json:"generated_authority"`
	SemanticClaims            int `json:"semantic_claims"`
	RepositoryWrites          int `json:"repository_writes"`
	MutationAuthorities       int `json:"mutation_authorities"`
	PromotionAuthorities      int `json:"promotion_authorities"`
	SemanticAuthorities       int `json:"semantic_authorities"`
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
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
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
