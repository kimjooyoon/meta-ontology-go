package model

const (
	ContractSchema = "gooo/evidence-freshness-contract/v1"
	ReceiptSchema  = "gooo/evidence-freshness-receipt/v1"
	ReportSchema   = "gooo/evidence-freshness-report/v1"
	ContextSchema  = "gooo/evidence-freshness-context/v1"
	VerdictSchema  = "gooo/evidence-freshness-verdict/v1"

	IndependenceSchema = "gooo/evidence-freshness-independence/v1"
	Scope              = "CLAIM_JUSTIFICATION_BOUNDARY_ONLY"

	ProducerID         = "evidence-freshness-producer/v1"
	ConsumerID         = "evidence-freshness-decider/v1"
	MetaOperationID    = "measure-claim-freshness-boundary"
	DefaultProofChoice = "FOUNDATION"

	DecisionPass        = "PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"

	StateFresh   = "FRESH"
	StateStale   = "STALE"
	StateUnknown = "UNKNOWN"

	StageSubject     = "SUBJECT_BINDING"
	StageMaterial    = "MATERIAL_CLOSURE"
	StageRecipe      = "RECIPE_RESOLUTION"
	StageEnvironment = "ENVIRONMENT_CAPTURE"
	StageRunner      = "RUNNER_EXECUTION"
	StageVerifier    = "VERIFIER_JUDGMENT"

	CaseTotal                 = 10
	AxisTotal                 = 6
	CheckTotal                = 10
	MetricTotal               = 10
	TransitionTotal           = 10
	IndependenceContractTotal = 1
)

// EvidenceTuple is the identity of the justification boundary. These six
// values deliberately stay separate: equal bytes do not imply equal claims.
type EvidenceTuple struct {
	Subject     string `json:"subject"`
	Material    string `json:"material"`
	Recipe      string `json:"recipe"`
	Environment string `json:"environment"`
	Runner      string `json:"runner"`
	Verifier    string `json:"verifier"`
}

// TemporalBoundary makes the time and environment limits of a claim explicit
// metadata rather than an implicit cache policy.
type TemporalBoundary struct {
	ObservationEpoch    int    `json:"observation_epoch"`
	ValidThroughEpoch   int    `json:"valid_through_epoch"`
	EnvironmentBoundary string `json:"environment_boundary"`
}

type Context struct {
	Schema              string        `json:"schema"`
	Tuple               EvidenceTuple `json:"tuple"`
	CurrentEpoch        int           `json:"current_epoch"`
	EnvironmentBoundary string        `json:"environment_boundary"`
	Consumer            string        `json:"consumer"`
}

type Receipt struct {
	Schema            string               `json:"schema"`
	HeadSHA           string               `json:"head_sha"`
	ClaimID           string               `json:"claim_id"`
	Producer          string               `json:"producer"`
	Consumer          string               `json:"consumer"`
	MetaOperation     string               `json:"meta_operation"`
	ProofChoice       string               `json:"proof_choice"`
	Tuple             EvidenceTuple        `json:"tuple"`
	Boundary          TemporalBoundary     `json:"boundary"`
	Independence      IndependenceEvidence `json:"independence"`
	RepositoryWrites  int                  `json:"repository_writes"`
	MutationAuthority bool                 `json:"mutation_authority"`
	Digest            string               `json:"digest"`
}

type CaseDefinition struct {
	ID                 string `json:"id"`
	Mutation           string `json:"mutation"`
	ExpectedState      string `json:"expected_state"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedStage      string `json:"expected_stage"`
	ExpectedStep       string `json:"expected_step"`
	ExpectedReason     string `json:"expected_reason"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}

type MetricDefinition struct {
	MetricID          string `json:"metric_id"`
	Class             string `json:"class"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	ProofChoice       string `json:"proof_choice"`
	MetaOperation     string `json:"meta_operation"`
	ExpectedNumerator int    `json:"expected_numerator"`
	Denominator       int    `json:"denominator"`
}

type Contract struct {
	Schema      string             `json:"schema"`
	Scope       string             `json:"scope"`
	SourcePath  string             `json:"source_path"`
	BaseContext Context            `json:"base_context"`
	Cases       []CaseDefinition   `json:"cases"`
	Metrics     []MetricDefinition `json:"metrics"`
	NotClaimed  []string           `json:"not_claimed"`
}

type FixedMetric struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

type IndependenceEvidence struct {
	Schema                   string      `json:"schema"`
	ForbiddenDependencyCount int         `json:"forbidden_dependency_count"`
	IndependenceContract     FixedMetric `json:"independence_contract"`
}

func DefaultIndependenceEvidence() IndependenceEvidence {
	return IndependenceEvidence{Schema: IndependenceSchema,
		IndependenceContract: FixedMetric{Numerator: IndependenceContractTotal, Denominator: IndependenceContractTotal}}
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	ClaimID        string     `json:"claim_id"`
	From           string     `json:"from"`
	To             string     `json:"to"`
	Preservation   string     `json:"preservation"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
}

type CheckResult struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Expected      string `json:"expected"`
	Observed      string `json:"observed"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Verdict struct {
	Schema            string          `json:"schema"`
	State             string          `json:"state"`
	Decision          string          `json:"decision"`
	Resolution        string          `json:"resolution"`
	Reason            string          `json:"reason"`
	Coordinate        Coordinate      `json:"coordinate"`
	ChangedDimensions []string        `json:"changed_dimensions"`
	ReceiptDigest     string          `json:"receipt_digest"`
	ContextDigest     string          `json:"context_digest"`
	Transition        ClaimTransition `json:"transition"`
	Checks            []CheckResult   `json:"checks"`
	Digest            string          `json:"digest"`
}

type CaseResult struct {
	ID                 string          `json:"id"`
	Status             string          `json:"status"`
	Mutation           string          `json:"mutation"`
	ExpectedState      string          `json:"expected_state"`
	ExpectedDecision   string          `json:"expected_decision"`
	ExpectedResolution string          `json:"expected_resolution"`
	ExpectedStage      string          `json:"expected_stage"`
	ExpectedStep       string          `json:"expected_step"`
	ExpectedReason     string          `json:"expected_reason"`
	ObservedState      string          `json:"observed_state"`
	ObservedDecision   string          `json:"observed_decision"`
	ObservedResolution string          `json:"observed_resolution"`
	ObservedReason     string          `json:"observed_reason"`
	Coordinate         Coordinate      `json:"coordinate"`
	Context            Context         `json:"context"`
	ChangedDimensions  []string        `json:"changed_dimensions"`
	Transition         ClaimTransition `json:"transition"`
	Checks             []CheckResult   `json:"checks"`
}

type Summary struct {
	CasesSatisfied           int            `json:"cases_satisfied"`
	CasesTotal               int            `json:"cases_total"`
	FreshCases               int            `json:"fresh_cases"`
	StaleCases               int            `json:"stale_cases"`
	UnknownCases             int            `json:"unknown_cases"`
	AxisChangesObserved      int            `json:"axis_changes_observed"`
	FixedAxisDenominator     int            `json:"fixed_axis_denominator"`
	StaleByStage             map[string]int `json:"stale_by_stage"`
	UnknownByStage           map[string]int `json:"unknown_by_stage"`
	PreservationTransitions  int            `json:"preservation_transitions"`
	TemporalBoundaryCases    int            `json:"temporal_boundary_cases"`
	ReadOnlyCases            int            `json:"read_only_cases"`
	ForbiddenDependencyCount int            `json:"forbidden_dependency_count"`
	IndependenceContract     FixedMetric    `json:"independence_contract"`
}

type Indicator struct {
	MetricID          string `json:"metric_id"`
	Class             string `json:"class"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	ProofChoice       string `json:"proof_choice"`
	MetaOperation     string `json:"meta_operation"`
	Numerator         int    `json:"numerator"`
	Denominator       int    `json:"denominator"`
	BasisPoints       int    `json:"basis_points"`
	ExpectedNumerator int    `json:"expected_numerator"`
	Satisfied         bool   `json:"satisfied"`
}

type Report struct {
	Schema             string               `json:"schema"`
	Scope              string               `json:"scope"`
	HeadSHA            string               `json:"head_sha"`
	Decision           string               `json:"decision"`
	Resolution         string               `json:"resolution"`
	Reason             string               `json:"reason"`
	ContractDigest     string               `json:"contract_digest"`
	SourceDigest       string               `json:"source_digest"`
	Receipt            Receipt              `json:"receipt"`
	ReceiptDigest      string               `json:"receipt_digest"`
	Independence       IndependenceEvidence `json:"independence"`
	IndependenceDigest string               `json:"independence_digest"`
	Cases              []CaseResult         `json:"cases"`
	Summary            Summary              `json:"summary"`
	Indicators         []Indicator          `json:"indicators"`
	NotClaimed         []string             `json:"not_claimed"`
	RepositoryWrites   int                  `json:"repository_writes"`
	MutationAuthority  bool                 `json:"mutation_authority"`
	Digest             string               `json:"digest"`
}
