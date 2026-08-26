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
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Improvement struct {
	ID             string     `json:"id"`
	Before         Coordinate `json:"before"`
	After          Coordinate `json:"after"`
	BeforeEvidence string     `json:"before_evidence"`
	AfterEvidence  string     `json:"after_evidence"`
}
