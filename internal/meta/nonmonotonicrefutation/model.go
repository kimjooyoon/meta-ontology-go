package nonmonotonicrefutation

const (
	ContractSchema = "gooo/meta-nonmonotonic-refutation-contract/v1"
	ProducerSchema = "gooo/meta-nonmonotonic-refutation-producer/v1"

	StatusOpen       = "OPEN"
	StatusDischarged = "DISCHARGED"
	StatusRefuted    = "REFUTED"

	EvidenceSupport = "SUPPORT"
	EvidenceRefute  = "REFUTE"

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

type Evidence struct {
	ID            string     `json:"id"`
	ClaimID       string     `json:"claim_id"`
	Kind          string     `json:"kind"`
	Basis         string     `json:"basis"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    Coordinate `json:"coordinate"`
}

type CaseDefinition struct {
	ID                  string     `json:"id"`
	ClaimID             string     `json:"claim_id"`
	InitialStatus       string     `json:"initial_status"`
	ExpectedFinalStatus string     `json:"expected_final_status"`
	Producer            string     `json:"producer"`
	Consumer            string     `json:"consumer"`
	MetaOperation       string     `json:"meta_operation"`
	ProofChoice         string     `json:"proof_choice"`
	Evidence            []Evidence `json:"evidence"`
}

type Contract struct {
	Schema               string           `json:"schema"`
	FixedClaimTotal      int              `json:"fixed_claim_total"`
	FixedTransitionTotal int              `json:"fixed_transition_total"`
	Cases                []CaseDefinition `json:"cases"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type ProducerReport struct {
	Schema        string   `json:"schema"`
	Contract      Contract `json:"contract"`
	SourcePath    string   `json:"source_path"`
	SourceDigest  string   `json:"source_digest"`
	Producer      string   `json:"producer"`
	Consumer      string   `json:"consumer"`
	MetaOperation string   `json:"meta_operation"`
	ProofChoice   string   `json:"proof_choice"`
	Effects       Effects  `json:"effects"`
	NotClaimed    []string `json:"not_claimed"`
	ReceiptDigest string   `json:"receipt_digest"`
}
