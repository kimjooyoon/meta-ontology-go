package phaseseparation

type Source struct {
	Producer     string
	Consumer     string
	MetaOperation string
	ProofChoice  string
	Cases        []Case
}

type Case struct {
	Name      string
	Values    []Value
	Transfers []Transfer
}

type Value struct {
	Phase   string
	ID      string
	Literal string
}

type Transfer struct {
	FromPhase string
	FromID    string
	ToPhase   string
	ToID      string
	Kind      string
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type CaseResult struct {
	Name             string `json:"name"`
	Class            string `json:"class"`
	Expected         string `json:"expected"`
	Actual           string `json:"actual"`
	Reason           string `json:"reason"`
	Passed           bool   `json:"passed"`
	TransitionCount  int    `json:"transition_count"`
}

type ClaimTransition struct {
	ID             string `json:"id"`
	FromPhase      string `json:"from_phase"`
	ToPhase        string `json:"to_phase"`
	FromClaim      string `json:"from_claim"`
	ToClaim        string `json:"to_claim"`
	FromState      string `json:"from_state"`
	ToState        string `json:"to_state"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Preserved      bool   `json:"preserved"`
}

type Indicator struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience      string `json:"audience"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Satisfied     int    `json:"satisfied"`
	Total         int    `json:"total"`
	BasisPoints   int    `json:"basis_points"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Summary struct {
	CleanCasesPassed         int `json:"clean_cases_passed"`
	CleanCasesTotal          int `json:"clean_cases_total"`
	LeakageCasesCaught       int `json:"leakage_cases_caught"`
	LeakageCasesTotal        int `json:"leakage_cases_total"`
	ClaimTransitionsPreserved int `json:"claim_transitions_preserved"`
	ClaimTransitionsTotal    int `json:"claim_transitions_total"`
	IndicatorsSatisfied       int `json:"indicators_satisfied"`
	IndicatorsTotal           int `json:"indicators_total"`
	UnknownCases              int `json:"unknown_cases"`
	RepositoryWrites          int `json:"repository_writes"`
}

type Authority struct {
	Execution bool `json:"execution"`
	Mutation  bool `json:"mutation"`
	Promotion bool `json:"promotion"`
}

type Report struct {
	Schema          string            `json:"schema"`
	Decision        string            `json:"decision"`
	Reason          string            `json:"reason"`
	Resolution      string            `json:"resolution"`
	HeadSHA         string            `json:"head_sha"`
	Toolchain       string            `json:"toolchain"`
	SourcePath      string            `json:"source_path"`
	SourceDigest    string            `json:"source_digest"`
	LeakSourcePath  string            `json:"leak_source_path"`
	LeakSourceDigest string           `json:"leak_source_digest"`
	Producer        string            `json:"producer"`
	Consumer        string            `json:"consumer"`
	MetaOperation   string            `json:"meta_operation"`
	ProofChoice     string            `json:"proof_choice"`
	Cases           []CaseResult      `json:"cases"`
	Transitions     []ClaimTransition `json:"claim_transitions"`
	Indicators      []Indicator       `json:"indicators"`
	Views           []View            `json:"views"`
	Proofs          []Proof           `json:"proofs"`
	Summary         Summary           `json:"summary"`
	Authority       Authority         `json:"authority"`
	Coordinate      Coordinate        `json:"coordinate"`
	Digest          string            `json:"digest"`
}
