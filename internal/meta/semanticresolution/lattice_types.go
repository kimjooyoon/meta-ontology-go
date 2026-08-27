package semanticresolution

type LatticeStage string

const (
	StageExact              LatticeStage = "EXACT"
	StagePartialObservation LatticeStage = "PARTIAL_OBSERVATION"
	StageInvariantOnly      LatticeStage = "INVARIANT_ONLY"
)

type ProofLevel string

const (
	ProofLevelFoundation ProofLevel = "FOUNDATION"
	ProofLevelCoherence  ProofLevel = "COHERENCE"
	ProofLevelRegression ProofLevel = "REGRESSION"
)

type ClaimState string

const (
	ClaimOpen       ClaimState = "OPEN"
	ClaimDischarged ClaimState = "DISCHARGED"
	ClaimRefuted    ClaimState = "REFUTED"
)

type PartialObservation struct {
	Required          int    `json:"required"`
	Observed          int    `json:"observed"`
	Reason            string `json:"reason"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type UnknownValue struct {
	Stage  LatticeStage `json:"stage"`
	Step   int          `json:"step"`
	Reason string       `json:"reason"`
}

type LatticeTransition struct {
	FromResolution    Resolution    `json:"from_resolution"`
	ToResolution      Resolution    `json:"to_resolution,omitempty"`
	Decision          string        `json:"decision"`
	Reason            string        `json:"reason"`
	Unknown           *UnknownValue `json:"unknown,omitempty"`
	RepositoryWrites  int           `json:"repository_writes"`
	MutationAuthority bool          `json:"mutation_authority"`
}

type LatticeCase struct {
	ID          string             `json:"id"`
	Decision    string             `json:"decision"`
	Observation PartialObservation `json:"observation"`
	Transition  LatticeTransition  `json:"transition"`
	ClaimID     string             `json:"claim_id"`
}

type ClaimRecord struct {
	ID          string     `json:"id"`
	State       ClaimState `json:"state"`
	BeforeState ClaimState `json:"before_state"`
	AfterState  ClaimState `json:"after_state"`
	Preserved   bool       `json:"preserved"`
}

type LatticeMetric struct {
	ID            string     `json:"id"`
	Class         string     `json:"class"`
	Numerator     int        `json:"numerator"`
	Denominator   int        `json:"denominator"`
	Unit          string     `json:"unit"`
	Relation      string     `json:"relation"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	Proof         ProofLevel `json:"proof"`
}

type LatticeCounts struct {
	CasesTotal int `json:"cases_total"`
	Pass       int `json:"pass"`
	FailClosed int `json:"fail_closed"`
	Unknown    int `json:"unknown"`
}

type LatticeCounterfactual struct {
	ID                     string            `json:"id"`
	Kind                   string            `json:"kind"`
	BaselineDecision       string            `json:"baseline_decision"`
	VariantDecision        string            `json:"variant_decision"`
	BaselineTransition     LatticeTransition `json:"baseline_transition"`
	VariantTransition      LatticeTransition `json:"variant_transition"`
	BaselineClaim          ClaimRecord       `json:"baseline_claim"`
	VariantClaim           ClaimRecord       `json:"variant_claim"`
	SourceDigestChanged    bool              `json:"source_digest_changed"`
	SemanticDigestChanged  bool              `json:"semantic_digest_changed"`
	DecisionChanged        bool              `json:"decision_changed"`
	ClaimTransitionChanged bool              `json:"claim_transition_changed"`
}

type LatticeReceipt struct {
	Schema            string                  `json:"schema"`
	Source            string                  `json:"source"`
	SourceSHA256      string                  `json:"source_sha256"`
	SemanticDigest    string                  `json:"semantic_digest"`
	RepositoryWrites  int                     `json:"repository_writes"`
	MutationAuthority bool                    `json:"mutation_authority"`
	CaseDenominator   int                     `json:"case_denominator"`
	Counts            LatticeCounts           `json:"counts"`
	Cases             []LatticeCase           `json:"cases"`
	Claims            []ClaimRecord           `json:"claims"`
	Counterfactuals   []LatticeCounterfactual `json:"counterfactuals"`
	Metrics           []LatticeMetric         `json:"metrics"`
}
