package ambiguitybudgetjudge

const (
	ContractSchema = "gooo/ambiguity-budget-contract/v1"
	ReceiptSchema  = "gooo/ambiguity-budget-receipt/v1"
	JudgeSchema    = "gooo/ambiguity-budget-judge/v1"
	Producer       = "gooo://meta/ambiguity-budget/producer"
	Consumer       = "gooo://meta/ambiguity-budget/independent-verifier"
	MetaOperation  = "measure-deterministic-ambiguity-budget"
	CaseTotal      = 4
	Denominator    = 12
)

type integerSet struct {
	InterpretationCandidates int `json:"interpretation_candidates"`
	UnresolvedBranches       int `json:"unresolved_branches"`
	EvidencePaths            int `json:"evidence_paths"`
}

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type transition struct {
	CaseID string `json:"case_id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type caseSpec struct {
	ID                 string     `json:"id"`
	Class              string     `json:"class"`
	InputState         string     `json:"input_state"`
	Counts             integerSet `json:"counts"`
	ExpectedDecision   string     `json:"expected_decision"`
	ExpectedResolution string     `json:"expected_resolution"`
	ExpectedReason     string     `json:"expected_reason"`
	Coordinate         coordinate `json:"coordinate"`
	Claim              transition `json:"claim"`
}

type contract struct {
	Schema           string     `json:"schema"`
	ID               string     `json:"id"`
	SourcePath       string     `json:"source_path"`
	SourcePackage    string     `json:"source_package"`
	SourceNamespace  string     `json:"source_namespace"`
	SourceEntities   int        `json:"source_entities"`
	SourceActivities int        `json:"source_activities"`
	Budget           integerSet `json:"budget"`
	FixedDenominator int        `json:"fixed_denominator"`
	Cases            []caseSpec `json:"cases"`
	NotClaimed       []string   `json:"not_claimed"`
}

type sourceObservation struct {
	Path       string `json:"path"`
	Digest     string `json:"digest"`
	Package    string `json:"package"`
	Namespace  string `json:"namespace"`
	Entities   int    `json:"entities"`
	Activities int    `json:"activities"`
}

type caseReceipt struct {
	ID                 string     `json:"id"`
	Class              string     `json:"class"`
	InputState         string     `json:"input_state"`
	Counts             integerSet `json:"counts"`
	ExpectedDecision   string     `json:"expected_decision"`
	ExpectedResolution string     `json:"expected_resolution"`
	ExpectedReason     string     `json:"expected_reason"`
	Decision           string     `json:"decision"`
	Resolution         string     `json:"resolution"`
	Reason             string     `json:"reason"`
	Coordinate         coordinate `json:"coordinate"`
	Claim              transition `json:"claim"`
	Status             string     `json:"status"`
}

type indicator struct {
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

type proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type summary struct {
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

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type receipt struct {
	Schema        string            `json:"schema"`
	SubjectSHA    string            `json:"subject_sha"`
	ContractID    string            `json:"contract_id"`
	Source        sourceObservation `json:"source"`
	Budget        integerSet        `json:"budget"`
	Producer      string            `json:"producer"`
	Consumer      string            `json:"consumer"`
	MetaOperation string            `json:"meta_operation"`
	ProofChoice   string            `json:"proof_choice"`
	Decision      string            `json:"decision"`
	Resolution    string            `json:"resolution"`
	Reason        string            `json:"reason"`
	Coordinate    coordinate        `json:"coordinate"`
	Cases         []caseReceipt     `json:"cases"`
	Claims        []transition      `json:"claims"`
	Indicators    []indicator       `json:"indicators"`
	Proofs        []proof           `json:"proofs"`
	Summary       summary           `json:"summary"`
	NotClaimed    []string          `json:"not_claimed"`
	Effects       effects           `json:"effects"`
	FactsDigest   string            `json:"facts_digest"`
	Digest        string            `json:"digest"`
}

type Check struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	Coordinate coordinate `json:"coordinate"`
}

type Result struct {
	Schema               string  `json:"schema"`
	SubjectSHA           string  `json:"subject_sha"`
	ContractID           string  `json:"contract_id"`
	Producer             string  `json:"producer"`
	Consumer             string  `json:"consumer"`
	MetaOperation        string  `json:"meta_operation"`
	Decision             string  `json:"decision"`
	Resolution           string  `json:"resolution"`
	Reason               string  `json:"reason"`
	ReportDigest         string  `json:"report_digest"`
	SourceDigest         string  `json:"source_digest"`
	FixedDenominator     int     `json:"fixed_denominator"`
	CasesSatisfied       int     `json:"cases_satisfied"`
	CasesTotal           int     `json:"cases_total"`
	CoordinatesSatisfied int     `json:"coordinates_satisfied"`
	CoordinatesTotal     int     `json:"coordinates_total"`
	Checks               []Check `json:"checks"`
	RepositoryWrites     int     `json:"repository_writes"`
	MutationAuthority    bool    `json:"mutation_authority"`
	Digest               string  `json:"digest"`
}
