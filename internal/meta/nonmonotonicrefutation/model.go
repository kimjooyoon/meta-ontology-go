package nonmonotonicrefutation

const (
	ProducerSchema = "gooo/meta-nonmonotonic-refutation-producer/v2"

	StatusOpen       = "OPEN"
	StatusDischarged = "DISCHARGED"
	StatusRefuted    = "REFUTED"

	EvidenceSupport       = "SUPPORT"
	EvidenceContradicting = "CONTRADICTING"

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

// Claim is reconstructed from a claim entity and its first observation. It
// intentionally contains no expected final status: that is an oracle result,
// not source input.
type Claim struct {
	ID             string `json:"id"`
	Predicate      string `json:"predicate"`
	ExpectedValue  string `json:"expected_value"`
	InitialStatus  string `json:"initial_status"`
	RevisionPolicy string `json:"revision_policy"`
}

// Observation is the source-backed witness. Kind and transition reason are
// absent because the consumer must classify them from these values.
type Observation struct {
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
	Coordinate     Coordinate `json:"coordinate"`
}

// Contract is a source-derived wire model, not a consumer contract. The
// counts are measured from the parsed source and serve as fixed denominators.
type Contract struct {
	Schema                string        `json:"schema"`
	FixedClaimTotal       int           `json:"fixed_claim_total"`
	FixedObservationTotal int           `json:"fixed_observation_total"`
	FixedTransitionTotal  int           `json:"fixed_transition_total"`
	Claims                []Claim       `json:"claims"`
	Observations          []Observation `json:"observations"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type ProducerReport struct {
	Schema               string   `json:"schema"`
	Contract             Contract `json:"contract"`
	SourcePath           string   `json:"source_path"`
	SourceDigest         string   `json:"source_digest"`
	SourceSemanticDigest string   `json:"source_semantic_digest"`
	SourceModelDigest    string   `json:"source_model_digest"`
	Producer             string   `json:"producer"`
	Consumer             string   `json:"consumer"`
	MetaOperation        string   `json:"meta_operation"`
	ProofChoice          string   `json:"proof_choice"`
	Effects              Effects  `json:"effects"`
	NotClaimed           []string `json:"not_claimed"`
	ReceiptDigest        string   `json:"receipt_digest"`
}
