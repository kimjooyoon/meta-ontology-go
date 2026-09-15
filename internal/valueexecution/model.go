package valueexecution

const (
	ReportSchema         = "gooo.language.value-witness/v2"
	DecisionProven       = "VALUE_WITNESS_PROVEN"
	DecisionFailClosed   = "FAIL_CLOSED"
	ReasonExactWitness   = "VALUE_WITNESS_EXACT"
	ResolutionCoreValue  = "CORE_IR_ACTIVITY_VALUE_PROGRAM"
	ResolutionBidirValue = "BIDIR_ACTIVITY_SEMANTIC"
	ResolutionSyntaxOnly = "SYNTAX_ONLY"
	ValueIndicatorCount  = 18
)

const RegisteredValueOperationScope = "REGISTERED_VALUE_OPERATION"

type Report struct {
	Schema              string                 `json:"schema"`
	Scope               string                 `json:"scope"`
	Decision            string                 `json:"decision"`
	Reason              string                 `json:"reason"`
	Resolution          string                 `json:"resolution"`
	HeadSHA             string                 `json:"head_sha"`
	SourcePath          string                 `json:"source_path"`
	SourceDigest        string                 `json:"source_digest"`
	SourceBytes         int                    `json:"source_bytes"`
	SourceLines         int                    `json:"source_lines"`
	SemanticFingerprint string                 `json:"semantic_fingerprint"`
	CoreIRFingerprint   string                 `json:"core_ir_fingerprint"`
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

func DefaultNonClaims() []string {
	return []string{
		"general expression language", "arbitrary value types", "core IR execution or code generation",
		"runtime memory or performance bounds", "repository mutation, promotion, or automatic adoption",
		"handwritten Go-body execution", "external effects",
	}
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
