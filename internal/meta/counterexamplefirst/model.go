package counterexamplefirst

import "reflect"

const (
	ContractSchema     = "gooo/counterexample-first-compiler-contract/v1"
	ReceiptSchema      = "gooo/counterexample-first-compiler-receipt/v1"
	ReportSchema       = "gooo/counterexample-first-judge-report/v1"
	CorpusSchema       = "gooo/counterexample-first-scenarios/v1"
	DenominatorVersion = "counterexample-first-denominator/v1"
	ContractID         = "counterexample-first-meta-compilation-v1"
	ProducerID         = "counterexample-first-compiler"
	ConsumerID         = "independent-counterexample-first-judge"
	MetaOperationID    = "compile-after-counterexample-resolution"
	CaseCount          = 5
	IndicatorCount     = 10
	TransitionCount    = 15
)

type FixedDenominator struct {
	Version            string `json:"version"`
	Cases              int    `json:"cases"`
	Indicators         int    `json:"indicators"`
	ClaimTransitions   int    `json:"claim_transitions"`
	UnknownCoordinates int    `json:"unknown_coordinates"`
}

type Contract struct {
	Schema        string           `json:"schema"`
	ID            string           `json:"id"`
	Version       int              `json:"version"`
	SourcePath    string           `json:"source_path"`
	Package       string           `json:"package"`
	Namespace     string           `json:"namespace"`
	Producer      string           `json:"producer"`
	Consumer      string           `json:"consumer"`
	MetaOperation string           `json:"meta_operation"`
	Fixed         FixedDenominator `json:"fixed_denominator"`
	Cases         []CaseSpec       `json:"cases"`
	NotClaimed    []string         `json:"not_claimed"`
}

type CaseSpec struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	ExpectedStage      string `json:"expected_stage"`
	ExpectedStep       string `json:"expected_step"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}

type ScenarioCorpus struct {
	Schema    string     `json:"schema"`
	Version   int        `json:"version"`
	Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
	ID             string              `json:"id"`
	Candidate      Candidate           `json:"candidate"`
	Counterexample *Counterexample     `json:"counterexample"`
	Resolution     *ResolutionEvidence `json:"resolution"`
}

type Candidate struct {
	ID             string `json:"id"`
	Claim          string `json:"claim"`
	SuccessExample string `json:"success_example"`
}

type ShrinkStep struct {
	FromSize         int  `json:"from_size"`
	ToSize           int  `json:"to_size"`
	PreservesFailure bool `json:"preserves_failure"`
}

type Counterexample struct {
	ID          string       `json:"id"`
	Stage       string       `json:"stage"`
	Step        string       `json:"step"`
	Reason      string       `json:"reason"`
	Failing     bool         `json:"failing"`
	Size        int          `json:"size"`
	Minimal     bool         `json:"minimal"`
	ShrinkTrace []ShrinkStep `json:"shrink_trace"`
}

type ResolutionEvidence struct {
	ID               string `json:"id"`
	CounterexampleID string `json:"counterexample_id"`
	Stage            string `json:"stage"`
	Step             string `json:"step"`
	Reason           string `json:"reason"`
	ProofChoice      string `json:"proof_choice"`
	MetaOperation    string `json:"meta_operation"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	Accepted         bool   `json:"accepted"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type DecisionInput struct {
	CandidateID           string `json:"candidate_id"`
	SuccessExampleDigest  string `json:"success_example_digest"`
	CounterexampleID      string `json:"counterexample_id"`
	CounterexampleDigest  string `json:"counterexample_digest"`
	ResolutionID          string `json:"resolution_id"`
	ResolutionDigest      string `json:"resolution_digest"`
	RequiredBeforeCompile bool   `json:"required_before_compile"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	From           string `json:"from"`
	To             string `json:"to"`
	Status         string `json:"status"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type DecisionReceipt struct {
	Schema           string            `json:"schema"`
	ContractID       string            `json:"contract_id"`
	HeadSHA          string            `json:"head_sha"`
	SourcePath       string            `json:"source_path"`
	SourceDigest     string            `json:"source_digest"`
	ScenarioID       string            `json:"scenario_id"`
	Producer         string            `json:"producer"`
	Consumer         string            `json:"consumer"`
	MetaOperation    string            `json:"meta_operation"`
	ProofChoice      string            `json:"proof_choice"`
	Decision         string            `json:"decision"`
	Resolution       string            `json:"resolution"`
	Reason           string            `json:"reason"`
	Coordinate       Coordinate        `json:"coordinate"`
	DecisionInput    DecisionInput     `json:"decision_input"`
	ClaimTransitions []ClaimTransition `json:"claim_transitions"`
	Effects          Effects           `json:"effects"`
	Digest           string            `json:"digest"`
}

type Summary struct {
	CasesSatisfied              int  `json:"cases_satisfied"`
	CasesTotal                  int  `json:"cases_total"`
	CounterexamplesRequired     int  `json:"counterexamples_required"`
	CounterexamplesObserved     int  `json:"counterexamples_observed"`
	MinimalCounterexamples      int  `json:"minimal_counterexamples"`
	ResolutionsObserved         int  `json:"resolutions_observed"`
	PromotionsAfterResolution   int  `json:"promotions_after_resolution"`
	SuccessOnlyBlocks           int  `json:"success_only_blocks"`
	UnknownCoordinatesPreserved int  `json:"unknown_coordinates_preserved"`
	ClaimTransitionsPreserved   int  `json:"claim_transitions_preserved"`
	ReceiptsVerified            int  `json:"receipts_verified"`
	DeterministicReplays        int  `json:"deterministic_replays"`
	ProducerDependencies        int  `json:"producer_dependencies"`
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthority           bool `json:"mutation_authority"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Denominator   int    `json:"denominator"`
	Satisfied     bool   `json:"satisfied"`
}

type JudgeInput struct {
	Contract             Contract
	HeadSHA              string
	SourcePath           string
	Source               []byte
	Corpus               ScenarioCorpus
	Receipts             []DecisionReceipt
	ProducerDependencies int
}

type Report struct {
	Schema           string            `json:"schema"`
	ContractID       string            `json:"contract_id"`
	HeadSHA          string            `json:"head_sha"`
	Decision         string            `json:"decision"`
	Resolution       string            `json:"resolution"`
	Reason           string            `json:"reason"`
	Denominator      FixedDenominator  `json:"fixed_denominator"`
	Summary          Summary           `json:"summary"`
	Indicators       []Indicator       `json:"indicators"`
	Receipts         []DecisionReceipt `json:"receipts"`
	NotClaimed       []string          `json:"not_claimed"`
	TamperedRejected int               `json:"tampered_rejected"`
	Digest           string            `json:"digest"`
}

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, ID: ContractID, Version: 1,
		SourcePath: "examples/counterexample-first-compiler/main.gooo",
		Package:    "counterexamplefirst", Namespace: "counterexamplefirst",
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID,
		Fixed: FixedDenominator{Version: DenominatorVersion, Cases: CaseCount,
			Indicators: IndicatorCount, ClaimTransitions: TransitionCount, UnknownCoordinates: 1},
		Cases: []CaseSpec{
			{ID: "resolved-minimal-counterexample", ExpectedDecision: "PASS", ExpectedResolution: "EXACT", ExpectedReason: "COUNTEREXAMPLE_RESOLVED", ExpectedStage: "COMPILE_DECISION", ExpectedStep: "promote-after-resolution", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "promote-after-resolution"},
			{ID: "success-example-only", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "COUNTEREXAMPLE_REQUIRED", ExpectedStage: "COUNTEREXAMPLE", ExpectedStep: "minimum-required", ProofChoice: "COUNTEREXAMPLE_REQUIRED", MetaOperation: "require-counterexample-before-compile"},
			{ID: "unresolved-minimal-counterexample", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "COUNTEREXAMPLE_UNRESOLVED", ExpectedStage: "RESOLUTION", ExpectedStep: "await-proof", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "block-unresolved-counterexample"},
			{ID: "non-minimal-counterexample", ExpectedDecision: "FAIL_CLOSED", ExpectedResolution: "EXACT", ExpectedReason: "COUNTEREXAMPLE_NOT_MINIMAL", ExpectedStage: "COUNTEREXAMPLE", ExpectedStep: "shrink", ProofChoice: "COUNTEREXAMPLE_SHRINKING", MetaOperation: "reject-nonminimal-counterexample"},
			{ID: "unknown-counterexample-coordinate", ExpectedDecision: "UNKNOWN", ExpectedResolution: "LOWER_RESOLUTION", ExpectedReason: "COUNTEREXAMPLE_COORDINATE_UNKNOWN", ExpectedStage: "UNKNOWN", ExpectedStep: "UNKNOWN", ProofChoice: "UNKNOWN_PRESERVATION", MetaOperation: "preserve-unknown-coordinate"},
		},
		NotClaimed: []string{"general compiler correctness", "global minimality", "theorem proving", "performance beyond the fixed corpus", "repository mutation"},
	}
}

func ValidContract(value Contract) bool {
	return reflect.DeepEqual(value, CanonicalContract())
}
