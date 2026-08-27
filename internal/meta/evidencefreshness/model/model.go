package model

const (
	ContractSchema     = "gooo/evidence-freshness-contract/v2"
	PolicySchema       = "gooo/evidence-freshness-policy/v1"
	ReceiptSchema      = "gooo/evidence-freshness-receipt/v2"
	ReportSchema       = "gooo/evidence-freshness-report/v2"
	ContextSchema      = "gooo/evidence-freshness-context/v2"
	VerdictSchema      = "gooo/evidence-freshness-verdict/v2"
	LedgerSchema       = "gooo/evidence-freshness-ledger/v1"
	IndependenceSchema = "gooo/evidence-freshness-independence/v1"
	Scope              = "CLAIM_JUSTIFICATION_BOUNDARY_ONLY"

	ProducerID         = "evidence-freshness-producer/v2"
	ConsumerID         = "evidence-freshness-decider/v2"
	MetaOperationID    = "measure-claim-freshness-boundary"
	DefaultProofChoice = "FOUNDATION"

	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"

	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"

	StateFresh   = "FRESH"
	StateStale   = "STALE"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"

	ObservationCurrent   = "CURRENT_EVIDENCE"
	ObservationSynthetic = "SYNTHETIC_COUNTEREXAMPLE"

	ClaimOpen       = "OPEN"
	ClaimDischarged = "DISCHARGED"
	ClaimRefuted    = "REFUTED"

	StageSubject     = "SUBJECT_BINDING"
	StageMaterial    = "MATERIAL_CLOSURE"
	StageRecipe      = "RECIPE_RESOLUTION"
	StageEnvironment = "ENVIRONMENT_CAPTURE"
	StageRunner      = "RUNNER_EXECUTION"
	StageVerifier    = "VERIFIER_JUDGMENT"

	CaseTotal                    = 10
	CurrentEvidenceTotal         = 1
	SyntheticCounterexampleTotal = 9
	AxisTotal                    = 6
	TransitionTotal              = 10
	MetricTotal                  = 13
	CheckTotal                   = 10
	IndependenceContractTotal    = 1
)

type FreshnessPolicy struct {
	Schema            string       `json:"schema"`
	Axes              []AxisPolicy `json:"axes"`
	ComparisonPolicy  string       `json:"comparison_policy"`
	PriorClaimState   string       `json:"prior_claim_state"`
	BoundaryPolicy    string       `json:"boundary_policy"`
	RawMaterialPolicy string       `json:"raw_material_policy"`
	SemanticPolicy    string       `json:"semantic_policy"`
	ClaimLedgerPolicy string       `json:"claim_ledger_policy"`
	EffectPolicy      string       `json:"effect_policy"`
}

type AxisPolicy struct {
	Name   string `json:"name"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type EvidenceTuple struct {
	Subject     string         `json:"subject"`
	Material    MaterialDigest `json:"material"`
	Recipe      string         `json:"recipe"`
	Environment string         `json:"environment"`
	Runner      string         `json:"runner"`
	Verifier    string         `json:"verifier"`
}

type MaterialDigest struct {
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type TemporalBoundary struct {
	ObservationEpoch    int    `json:"observation_epoch"`
	ValidThroughEpoch   int    `json:"valid_through_epoch"`
	EnvironmentBoundary string `json:"environment_boundary"`
}

type WriteSetObservation struct {
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	BeforeCount  int    `json:"before_count"`
	AfterCount   int    `json:"after_count"`
}

type Context struct {
	Schema              string        `json:"schema"`
	PolicyDigest        string        `json:"policy_digest"`
	Tuple               EvidenceTuple `json:"tuple"`
	CurrentEpoch        int           `json:"current_epoch"`
	EnvironmentBoundary string        `json:"environment_boundary"`
	Consumer            string        `json:"consumer"`
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

func DefaultWriteSetObservation() WriteSetObservation {
	emptyDigest := DigestBytes(nil)
	return WriteSetObservation{BeforeDigest: emptyDigest, AfterDigest: emptyDigest}
}

type Receipt struct {
	Schema          string               `json:"schema"`
	HeadSHA         string               `json:"head_sha"`
	ClaimID         string               `json:"claim_id"`
	Producer        string               `json:"producer"`
	Consumer        string               `json:"consumer"`
	MetaOperation   string               `json:"meta_operation"`
	ProofChoice     string               `json:"proof_choice"`
	SourcePath      string               `json:"source_path"`
	PolicyDigest    string               `json:"policy_digest"`
	SourceDigest    string               `json:"source_digest"`
	SemanticDigest  string               `json:"semantic_digest"`
	PriorClaimState string               `json:"prior_claim_state"`
	Tuple           EvidenceTuple        `json:"tuple"`
	Boundary        TemporalBoundary     `json:"boundary"`
	Independence    IndependenceEvidence `json:"independence"`
	WriteSet        WriteSetObservation  `json:"write_set"`
	Digest          string               `json:"digest"`
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
	Metrics     []MetricDefinition `json:"metrics"`
	NotClaimed  []string           `json:"not_claimed"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

// ClaimTransition is the freshness observation. It is deliberately distinct
// from the canonical claim ledger state transition.
type ClaimTransition struct {
	ClaimID        string     `json:"claim_id"`
	From           string     `json:"from"`
	To             string     `json:"to"`
	Preservation   string     `json:"preservation"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
}

type ClaimLedgerEntry struct {
	Schema               string   `json:"schema"`
	Sequence             int      `json:"sequence"`
	ClaimID              string   `json:"claim_id"`
	PriorState           string   `json:"prior_state"`
	NextState            string   `json:"next_state"`
	Preservation         string   `json:"preservation"`
	FreshnessObservation string   `json:"freshness_observation"`
	ReceiptDigest        string   `json:"receipt_digest"`
	SourceDigest         string   `json:"source_digest"`
	SemanticDigest       string   `json:"semantic_digest"`
	PreviousDigest       string   `json:"previous_digest"`
	Provenance           []string `json:"provenance"`
	Digest               string   `json:"digest"`
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
	RawFreshness      string          `json:"raw_freshness"`
	SemanticFreshness string          `json:"semantic_freshness"`
	SourceDigest      string          `json:"source_digest"`
	SemanticDigest    string          `json:"semantic_digest"`
	ReceiptDigest     string          `json:"receipt_digest"`
	ContextDigest     string          `json:"context_digest"`
	Transition        ClaimTransition `json:"transition"`
	Checks            []CheckResult   `json:"checks"`
	Digest            string          `json:"digest"`
}

type CaseResult struct {
	ID                 string           `json:"id"`
	ObservationClass   string           `json:"observation_class"`
	Mutation           string           `json:"mutation"`
	Status             string           `json:"status"`
	ObservedState      string           `json:"observed_state"`
	ObservedDecision   string           `json:"observed_decision"`
	ObservedResolution string           `json:"observed_resolution"`
	ObservedReason     string           `json:"observed_reason"`
	RawFreshness       string           `json:"raw_freshness"`
	SemanticFreshness  string           `json:"semantic_freshness"`
	SourceAvailable    bool             `json:"source_available"`
	SourceDigest       string           `json:"source_digest"`
	SemanticDigest     string           `json:"semantic_digest"`
	Coordinate         Coordinate       `json:"coordinate"`
	Context            Context          `json:"context"`
	ChangedDimensions  []string         `json:"changed_dimensions"`
	Transition         ClaimTransition  `json:"transition"`
	ClaimLedger        ClaimLedgerEntry `json:"claim_ledger"`
	Checks             []CheckResult    `json:"checks"`
}

type Summary struct {
	CasesObserved            int            `json:"cases_observed"`
	CurrentEvidenceCases     int            `json:"current_evidence_cases"`
	SyntheticCounterexamples int            `json:"synthetic_counterexample_cases"`
	AxisChangesObserved      int            `json:"axis_changes_observed"`
	FixedAxisDenominator     int            `json:"fixed_axis_denominator"`
	RawFreshCases            int            `json:"raw_fresh_cases"`
	RawStaleCases            int            `json:"raw_stale_cases"`
	RawUnknownCases          int            `json:"raw_unknown_cases"`
	SemanticFreshCases       int            `json:"semantic_fresh_cases"`
	SemanticStaleCases       int            `json:"semantic_stale_cases"`
	SemanticUnknownCases     int            `json:"semantic_unknown_cases"`
	ClaimFreshCases          int            `json:"claim_fresh_cases"`
	ClaimStaleCases          int            `json:"claim_stale_cases"`
	ClaimUnknownCases        int            `json:"claim_unknown_cases"`
	RawStaleByStage          map[string]int `json:"raw_stale_by_stage"`
	StaleByStage             map[string]int `json:"stale_by_stage"`
	UnknownByStage           map[string]int `json:"unknown_by_stage"`
	FreshnessTransitions     int            `json:"freshness_transitions"`
	ClaimLedgerEntries       int            `json:"claim_ledger_entries"`
	ClaimDischarged          int            `json:"claim_discharged"`
	ClaimOpenPreserved       int            `json:"claim_open_preserved"`
	ClaimRefuted             int            `json:"claim_refuted"`
	SourceReconstructedCases int            `json:"source_reconstructed_cases"`
	SourceUnavailableCases   int            `json:"source_unavailable_cases"`
	ForbiddenDependencyCount int            `json:"forbidden_dependency_count"`
	IndependenceContract     FixedMetric    `json:"independence_contract"`
	ReadOnlyBeforeCount      int            `json:"read_only_before_count"`
	ReadOnlyAfterCount       int            `json:"read_only_after_count"`
	ReadOnlyWriteSetStable   bool           `json:"read_only_write_set_stable"`
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
	Policy             FreshnessPolicy      `json:"policy"`
	PolicyDigest       string               `json:"policy_digest"`
	ContractDigest     string               `json:"contract_digest"`
	SourceDigest       string               `json:"source_digest"`
	SemanticDigest     string               `json:"semantic_digest"`
	Receipt            Receipt              `json:"receipt"`
	ReceiptDigest      string               `json:"receipt_digest"`
	Independence       IndependenceEvidence `json:"independence"`
	IndependenceDigest string               `json:"independence_digest"`
	Cases              []CaseResult         `json:"cases"`
	Summary            Summary              `json:"summary"`
	Indicators         []Indicator          `json:"indicators"`
	ClaimLedger        []ClaimLedgerEntry   `json:"claim_ledger"`
	ClaimLedgerDigest  string               `json:"claim_ledger_digest"`
	NotClaimed         []string             `json:"not_claimed"`
	WriteSet           WriteSetObservation  `json:"write_set"`
	Digest             string               `json:"digest"`
}
