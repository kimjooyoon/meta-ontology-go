package nonmonotonicrefutation

const (
	ProducerSchema = "gooo/meta-nonmonotonic-refutation-producer/v3"

	StatusOpen       = "OPEN"
	StatusDischarged = "DISCHARGED"
	StatusRefuted    = "REFUTED"

	RelationSupports     = "SUPPORTS"
	RelationContradicts  = "CONTRADICTS"
	RelationInsufficient = "INSUFFICIENT"
	RelationUnknown      = "UNKNOWN"

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

// Observation is a source-backed request. It contains no conclusion label,
// prior state, or revision policy. The consumer computes the relation.
type Observation struct {
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
	Coordinate     Coordinate `json:"coordinate"`
}

// Contract is reconstructed from the source and is not a consumer authority.
// Denominators are measured counts, not an expected result table.
type Contract struct {
	Schema                string        `json:"schema"`
	FixedCaseTotal        int           `json:"fixed_case_total"`
	FixedClaimTotal       int           `json:"fixed_claim_total"`
	FixedObservationTotal int           `json:"fixed_observation_total"`
	FixedLedgerRowTotal   int           `json:"fixed_ledger_row_total"`
	Claims                []Claim       `json:"claims"`
	Observations          []Observation `json:"observations"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
	PromotionCount    int  `json:"promotion_count"`
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
