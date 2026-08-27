package nonmonotonicrefutation

const (
	ProducerSchema = "gooo/meta-nonmonotonic-refutation-producer/v4"

	StatusOpen       = "OPEN"
	StatusDischarged = "DISCHARGED"
	StatusRefuted    = "REFUTED"

	RelationSupports     = "SUPPORTS"
	RelationContradicts  = "CONTRADICTS"
	RelationInsufficient = "INSUFFICIENT"
	RelationUnknown      = "UNKNOWN"

	RevisionNone       = "NONE"
	RevisionSupersedes = "SUPERSEDES"

	ProviderHistoricalFixture = "HISTORICAL_FIXTURE"

	PolicyUnknownRetain              = "RETAIN_CURRENT"
	PolicyInsufficientRetain         = "RETAIN_CURRENT"
	PolicyOrdinarySupportRetain      = "RETAIN_REFUTED"
	PolicyCorrectionTargetEvidence   = "EVIDENCE_DIGEST"
	PolicyFoundationFirstClaimEvent  = "FIRST_CLAIM_OBSERVATION"
	PolicyCoherenceLaterClaimOpening = "LATER_OBSERVATION_AFTER_FIRST_SOURCE_EVENT"
	PolicyRegressionTargetedHistory  = "SEQUENCE_AT_LEAST_5_WITH_PRIOR_CLAIM"

	ProducerID      = "producer://nonmonotonic-refutation"
	ConsumerID      = "consumer://nonmonotonic-refutation-oracle"
	MetaOperation   = "meta://revise-claim-by-evidence"
	ProofFoundation = "FOUNDATION"
	ProofCoherence  = "COHERENCE"
	ProofRegression = "REGRESSION"
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

// RevisionPolicy is a source-declared meta-semantic object. It describes
// admissibility and revision operations, but it never declares a conclusion.
type RevisionPolicy struct {
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

// Claim is a proposition reconstructed from source observations. State is
// intentionally absent: every consumer ledger starts each claim at OPEN.
type Claim struct {
	ID            string `json:"id"`
	Proposition   string `json:"proposition"`
	Subject       string `json:"subject"`
	Input         string `json:"input"`
	Predicate     string `json:"predicate"`
	ExpectedValue string `json:"expected_value"`
}

// EvidenceMaterial is the canonical material hashed for an observation. The
// source provides a fixture recipe; the producer and consumer independently
// derive this digest and never accept a source-supplied digest as authority.
type EvidenceMaterial struct {
	ClaimID                  string `json:"claim_id"`
	Proposition              string `json:"proposition"`
	TargetAddress            string `json:"target_address"`
	ObservedMaterial         string `json:"observed_material"`
	ObservedValue            string `json:"observed_value"`
	ObservationQuality       string `json:"observation_quality"`
	ProviderClass            string `json:"provider_class"`
	Sequence                 int    `json:"sequence"`
	SupersededEvidenceDigest string `json:"superseded_evidence_digest"`
}

// Observation is a source-backed fixture recipe. It contains no computed
// relation, status, or decision. RevisionRelation is an operation request
// (NONE or SUPERSEDES), not a conclusion; the consumer still classifies the
// evidence independently.
type Observation struct {
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
	PolicyID                 string     `json:"policy_id"`
	PolicyDigest             string     `json:"policy_digest"`
	Producer                 string     `json:"producer"`
	Consumer                 string     `json:"consumer"`
	MetaOperation            string     `json:"meta_operation"`
	ProofChoice              string     `json:"proof_choice"`
	Coordinate               Coordinate `json:"coordinate"`
	TargetAddress            string     `json:"target_address"`
}

// Contract is reconstructed from source and is not a consumer authority.
// Denominators are measured counts, not an expected result table.
type Contract struct {
	Schema                string         `json:"schema"`
	Policy                RevisionPolicy `json:"policy"`
	FixedCaseTotal        int            `json:"fixed_case_total"`
	FixedClaimTotal       int            `json:"fixed_claim_total"`
	FixedObservationTotal int            `json:"fixed_observation_total"`
	FixedLedgerRowTotal   int            `json:"fixed_ledger_row_total"`
	Claims                []Claim        `json:"claims"`
	Observations          []Observation  `json:"observations"`
}

type Effects struct {
	NetRepositoryStatusUnchanged bool   `json:"net_repository_status_unchanged"`
	RepositoryWriteObservation   string `json:"repository_write_observation"`
	MutationAuthorityResolution  string `json:"mutation_authority_resolution"`
	PromotionOperationsObserved  int    `json:"promotion_operations_observed"`
}

type ProducerReport struct {
	Schema               string   `json:"schema"`
	Contract             Contract `json:"contract"`
	SourcePath           string   `json:"source_path"`
	SourceDigest         string   `json:"source_digest"`
	SourceSemanticDigest string   `json:"source_semantic_digest"`
	SourceBindingDigest  string   `json:"source_binding_digest"`
	SourceModelDigest    string   `json:"source_model_digest"`
	Producer             string   `json:"producer"`
	Consumer             string   `json:"consumer"`
	MetaOperation        string   `json:"meta_operation"`
	ProofChoice          string   `json:"proof_choice"`
	Effects              Effects  `json:"effects"`
	NotClaimed           []string `json:"not_claimed"`
	ReceiptDigest        string   `json:"receipt_digest"`
}
