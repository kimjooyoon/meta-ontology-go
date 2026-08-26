package valueexecution

const (
	ReportSchema         = "gooo.language.value-witness/v1"
	DecisionProven       = "VALUE_WITNESS_PROVEN"
	DecisionFailClosed   = "FAIL_CLOSED"
	ReasonExactWitness   = "VALUE_WITNESS_EXACT"
	ResolutionBidirValue = "BIDIR_ACTIVITY_SEMANTIC"
	ResolutionSyntaxOnly = "SYNTAX_ONLY"
)

type Report struct {
	Schema              string                 `json:"schema"`
	Decision            string                 `json:"decision"`
	Reason              string                 `json:"reason"`
	Resolution          string                 `json:"resolution"`
	HeadSHA             string                 `json:"head_sha"`
	SourcePath          string                 `json:"source_path"`
	SourceDigest        string                 `json:"source_digest"`
	SourceBytes         int                    `json:"source_bytes"`
	SourceLines         int                    `json:"source_lines"`
	SemanticFingerprint string                 `json:"semantic_fingerprint"`
	Activity            string                 `json:"activity"`
	ValueProgram        string                 `json:"value_program"`
	ValueProgramDigest  string                 `json:"value_program_digest"`
	Registry            RegistrySummary        `json:"registry"`
	Improvement         Improvement            `json:"improvement"`
	Cases               []CaseResult           `json:"cases"`
	Counterexamples     []CounterexampleResult `json:"counterexamples"`
	Indicators          []Indicator            `json:"indicators"`
	Views               []View                 `json:"views"`
	Proofs              []Proof                `json:"proofs"`
	Summary             Summary                `json:"summary"`
	NonClaims           []string               `json:"non_claims"`
	Authority           Authority              `json:"authority"`
	Digest              string                 `json:"digest"`
}

type RegistrySummary struct {
	RegisteredOperations int      `json:"registered_operations"`
	InvokedOperations    int      `json:"invoked_operations"`
	OperationIDs         []string `json:"operation_ids"`
}

type Coordinate struct {
	Satisfied  int `json:"satisfied"`
	Total      int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Improvement struct {
	ID             string     `json:"id"`
	Before         Coordinate `json:"before"`
	After          Coordinate `json:"after"`
	BeforeEvidence string     `json:"before_evidence"`
	AfterEvidence  string     `json:"after_evidence"`
}

type CaseResult struct {
	ID             string `json:"id"`
	Input          int64  `json:"input"`
	Expected       int64  `json:"expected"`
	Actual         int64  `json:"actual"`
	Replay         int64  `json:"replay"`
	Passed         bool   `json:"passed"`
	ReplayMatched  bool   `json:"replay_matched"`
}

type CounterexampleResult struct {
	ID             string `json:"id"`
	ExpectedReason string `json:"expected_reason"`
	ActualReason   string `json:"actual_reason"`
	ReplayReason   string `json:"replay_reason"`
	Passed         bool   `json:"passed"`
	ReplayMatched  bool   `json:"replay_matched"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Summary struct {
	ValueCasesPassed       int        `json:"value_cases_passed"`
	ValueCasesTotal        int        `json:"value_cases_total"`
	CounterexamplesPassed  int        `json:"counterexamples_passed"`
	CounterexamplesTotal   int        `json:"counterexamples_total"`
	ValueOutputsObserved   int        `json:"value_outputs_observed"`
	DeterministicReplays   int        `json:"deterministic_replays"`
	RepositoryWrites       int        `json:"repository_writes"`
	CoreIRProgramPreserved Coordinate `json:"core_ir_program_preserved"`
	CoreIRFailClosed       Coordinate `json:"core_ir_fail_closed"`
}

type Authority struct {
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
	PromotionAuthorized          bool `json:"promotion_authorized"`
	AutomaticAdoptionAuthorized  bool `json:"automatic_adoption_authorized"`
}
