package ambiguitybudget

const (
	ContractSchema = "gooo/ambiguity-budget-contract/v1"
	ReceiptSchema  = "gooo/ambiguity-budget-receipt/v1"

	Producer           = "gooo://meta/ambiguity-budget/producer"
	Consumer           = "gooo://meta/ambiguity-budget/independent-verifier"
	MetaOperation      = "measure-deterministic-ambiguity-budget"
	FoundationProof    = "FOUNDATION"
	CoherenceProof     = "COHERENCE"
	RegressionProof    = "REGRESSION"
	ExpectedCaseTotal  = 4
	CoordinatesPerCase = 3
	FixedDenominator   = ExpectedCaseTotal * CoordinatesPerCase
)

type IntegerSet struct {
	InterpretationCandidates int `json:"interpretation_candidates"`
	UnresolvedBranches       int `json:"unresolved_branches"`
	EvidencePaths            int `json:"evidence_paths"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	CaseID string `json:"case_id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type CaseSpec struct {
	ID                 string          `json:"id"`
	Class              string          `json:"class"`
	InputState         string          `json:"input_state"`
	Counts             IntegerSet      `json:"counts"`
	ExpectedDecision   string          `json:"expected_decision"`
	ExpectedResolution string          `json:"expected_resolution"`
	ExpectedReason     string          `json:"expected_reason"`
	Coordinate         Coordinate      `json:"coordinate"`
	Claim              ClaimTransition `json:"claim"`
}

type Contract struct {
	Schema           string     `json:"schema"`
	ID               string     `json:"id"`
	SourcePath       string     `json:"source_path"`
	SourcePackage    string     `json:"source_package"`
	SourceNamespace  string     `json:"source_namespace"`
	SourceEntities   int        `json:"source_entities"`
	SourceActivities int        `json:"source_activities"`
	Budget           IntegerSet `json:"budget"`
	FixedDenominator int        `json:"fixed_denominator"`
	Cases            []CaseSpec `json:"cases"`
	NotClaimed       []string   `json:"not_claimed"`
}

type Input struct {
	SubjectSHA string
	Contract   Contract
	Source     []byte
}

type SourceObservation struct {
	Path       string `json:"path"`
	Digest     string `json:"digest"`
	Package    string `json:"package"`
	Namespace  string `json:"namespace"`
	Entities   int    `json:"entities"`
	Activities int    `json:"activities"`
}

type CaseReceipt struct {
	ID                 string          `json:"id"`
	Class              string          `json:"class"`
	InputState         string          `json:"input_state"`
	Counts             IntegerSet      `json:"counts"`
	ExpectedDecision   string          `json:"expected_decision"`
	ExpectedResolution string          `json:"expected_resolution"`
	ExpectedReason     string          `json:"expected_reason"`
	Decision           string          `json:"decision"`
	Resolution         string          `json:"resolution"`
	Reason             string          `json:"reason"`
	Coordinate         Coordinate      `json:"coordinate"`
	Claim              ClaimTransition `json:"claim"`
	Status             string          `json:"status"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	CaseID        string `json:"case_id"`
	Dimension     string `json:"dimension"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Summary struct {
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	CoordinatesSatisfied int `json:"coordinates_satisfied"`
	CoordinatesTotal     int `json:"coordinates_total"`
	FixedDenominator     int `json:"fixed_denominator"`
	ZeroAmbiguityCases   int `json:"zero_ambiguity_cases"`
	BoundaryCases        int `json:"boundary_cases"`
	OverBudgetCases      int `json:"over_budget_cases"`
	UnknownCases         int `json:"unknown_cases"`
	LowerResolutionCases int `json:"lower_resolution_cases"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema        string            `json:"schema"`
	SubjectSHA    string            `json:"subject_sha"`
	ContractID    string            `json:"contract_id"`
	Source        SourceObservation `json:"source"`
	Budget        IntegerSet        `json:"budget"`
	Producer      string            `json:"producer"`
	Consumer      string            `json:"consumer"`
	MetaOperation string            `json:"meta_operation"`
	ProofChoice   string            `json:"proof_choice"`
	Decision      string            `json:"decision"`
	Resolution    string            `json:"resolution"`
	Reason        string            `json:"reason"`
	Coordinate    Coordinate        `json:"coordinate"`
	Cases         []CaseReceipt     `json:"cases"`
	Claims        []ClaimTransition `json:"claims"`
	Indicators    []Indicator       `json:"indicators"`
	Proofs        []Proof           `json:"proofs"`
	Summary       Summary           `json:"summary"`
	NotClaimed    []string          `json:"not_claimed"`
	Effects       Effects           `json:"effects"`
	FactsDigest   string            `json:"facts_digest"`
	Digest        string            `json:"digest"`
}
