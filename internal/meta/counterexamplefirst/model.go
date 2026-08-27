package counterexamplefirst

const (
	ContractSchema     = "gooo/counterexample-first-compiler-contract/v2"
	ReceiptSchema      = "gooo/counterexample-first-compiler-receipt/v2"
	ReportSchema       = "gooo/counterexample-first-judge-report/v2"
	CorpusSchema       = "gooo/counterexample-first-scenarios/v2"
	DenominatorVersion = "counterexample-first-denominator/v2"
	ContractID         = "counterexample-first-meta-compilation-v2"
	ProducerID         = "counterexample-first-compiler"
	ConsumerID         = "independent-counterexample-first-judge"
	MetaOperationID    = "compile-after-counterexample-resolution"
	CaseCount          = 5
	IndicatorCount     = 12
	TransitionCount    = 15
)

const (
	RuleIdentityV1 = "identity:v1"
	RuleIdentityV2 = "identity:v2"
)

type FixedDenominator struct {
	Version            string `json:"version"`
	Cases              int    `json:"cases"`
	Indicators         int    `json:"indicators"`
	ClaimTransitions   int    `json:"claim_transitions"`
	UnknownCoordinates int    `json:"unknown_coordinates"`
	CorpusInputs       int    `json:"corpus_inputs"`
}

// Contract describes the operation and its input shape. It intentionally has
// no decision, resolution, failure, minimality, or acceptance table.
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
	Predicate     PredicateSpec    `json:"predicate"`
	Fixed         FixedDenominator `json:"fixed_denominator"`
	Cases         []CaseSpec       `json:"cases"`
	NotClaimed    []string         `json:"not_claimed"`
}

type PredicateSpec struct {
	ID          string `json:"id"`
	Operation   string `json:"operation"`
	Rule        string `json:"rule"`
	SourceFact  string `json:"source_fact"`
	FailureRule string `json:"failure_rule"`
}

type CaseSpec struct {
	ID            string `json:"id"`
	InputKind     string `json:"input_kind"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

// ScenarioCorpus contains only inputs. Conclusions are generated from the
// parser/lowerer observations and never supplied by this wire format.
type ScenarioCorpus struct {
	Schema    string     `json:"schema"`
	Version   int        `json:"version"`
	Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
	ID         string           `json:"id"`
	Candidate  Candidate        `json:"candidate"`
	Resolution *ResolutionInput `json:"resolution_input,omitempty"`
}

type Candidate struct {
	ID     string  `json:"id"`
	Claim  string  `json:"claim"`
	Source *string `json:"source"`
}

type ResolutionInput struct {
	ID     string  `json:"id"`
	Source *string `json:"source"`
}

type DiagnosticObservation struct {
	Code   string `json:"code"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type NodeObservation struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	ValueProgram string `json:"value_program,omitempty"`
}

// ExecutionObservation is a projection of actual ParseFile -> Lower output.
// It is evidence, not an input assertion.
type ExecutionObservation struct {
	InputID          string                  `json:"input_id"`
	SourceDigest     string                  `json:"source_digest"`
	SourceBytes      int                     `json:"source_bytes"`
	ParseDiagnostics []DiagnosticObservation `json:"parse_diagnostics"`
	ParseOK          bool                    `json:"parse_ok"`
	LowerOK          bool                    `json:"lower_ok"`
	LowerError       string                  `json:"lower_error,omitempty"`
	SemanticDigest   string                  `json:"semantic_digest,omitempty"`
	Nodes            []NodeObservation       `json:"nodes,omitempty"`
	OutputDigest     string                  `json:"output_digest"`
}

type PredicateObservation struct {
	Rule              string `json:"rule"`
	Applicable        bool   `json:"applicable"`
	ViolationObserved bool   `json:"violation_observed"`
	PassObserved      bool   `json:"pass_observed"`
	UnknownObserved   bool   `json:"unknown_observed"`
	Reason            string `json:"reason"`
}

type ShrinkObservation struct {
	CandidateDigest string               `json:"candidate_digest"`
	SourceBytes     int                  `json:"source_bytes"`
	Observation     ExecutionObservation `json:"observation"`
	Predicate       PredicateObservation `json:"predicate"`
}

type Counterexample struct {
	ID                    string               `json:"id"`
	SourceDigest          string               `json:"source_digest"`
	SourceBytes           int                  `json:"source_bytes"`
	Observation           ExecutionObservation `json:"observation"`
	Predicate             PredicateObservation `json:"predicate"`
	ShrinkTrace           []ShrinkObservation  `json:"shrink_trace"`
	MinimalityNumerator   int                  `json:"minimality_numerator"`
	MinimalityDenominator int                  `json:"minimality_denominator"`
	MinimalityProved      bool                 `json:"minimality_proved"`
	Stage                 string               `json:"stage"`
	Step                  string               `json:"step"`
	Reason                string               `json:"reason"`
}

type ResolutionEvidence struct {
	ID               string               `json:"id"`
	CounterexampleID string               `json:"counterexample_id"`
	InputID          string               `json:"input_id"`
	Observation      ExecutionObservation `json:"observation"`
	Predicate        PredicateObservation `json:"predicate"`
	Stage            string               `json:"stage"`
	Step             string               `json:"step"`
	Reason           string               `json:"reason"`
	ProofChoice      string               `json:"proof_choice"`
	MetaOperation    string               `json:"meta_operation"`
	Producer         string               `json:"producer"`
	Consumer         string               `json:"consumer"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type DecisionInput struct {
	CandidateID           string `json:"candidate_id"`
	CandidateDigest       string `json:"candidate_digest"`
	CounterexampleID      string `json:"counterexample_id"`
	CounterexampleDigest  string `json:"counterexample_digest"`
	ResolutionID          string `json:"resolution_id"`
	ResolutionDigest      string `json:"resolution_digest"`
	RequiredBeforeCompile bool   `json:"required_before_compile"`
}

// ClaimTransition is append-only evidence. A counterexample first moves OPEN
// to REFUTED; only an observed rerun can append REFUTED to DISCHARGED.
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
	Schema               string               `json:"schema"`
	ContractID           string               `json:"contract_id"`
	HeadSHA              string               `json:"head_sha"`
	SourcePath           string               `json:"source_path"`
	SourceDigest         string               `json:"source_digest"`
	SemanticDigest       string               `json:"semantic_digest"`
	ScenarioID           string               `json:"scenario_id"`
	Producer             string               `json:"producer"`
	Consumer             string               `json:"consumer"`
	MetaOperation        string               `json:"meta_operation"`
	ProofChoice          string               `json:"proof_choice"`
	Decision             string               `json:"decision"`
	Resolution           string               `json:"resolution"`
	Reason               string               `json:"reason"`
	Coordinate           Coordinate           `json:"coordinate"`
	CandidateObservation ExecutionObservation `json:"candidate_observation"`
	CandidatePredicate   PredicateObservation `json:"candidate_predicate"`
	Counterexample       *Counterexample      `json:"counterexample,omitempty"`
	ResolutionEvidence   *ResolutionEvidence  `json:"resolution_evidence,omitempty"`
	DecisionInput        DecisionInput        `json:"decision_input"`
	ClaimTransitions     []ClaimTransition    `json:"claim_transitions"`
	Effects              Effects              `json:"effects"`
	Digest               string               `json:"digest"`
}

type Summary struct {
	CasesSatisfied                  int  `json:"cases_satisfied"`
	CasesTotal                      int  `json:"cases_total"`
	CorpusExecutions                int  `json:"corpus_executions"`
	DiscoveredCounterexamples       int  `json:"discovered_counterexamples"`
	ShrinkCandidateExecutions       int  `json:"shrink_candidate_executions"`
	MinimalityNumerator             int  `json:"minimality_numerator"`
	MinimalityDenominator           int  `json:"minimality_denominator"`
	ResolutionReruns                int  `json:"resolution_reruns"`
	PromotionsAfterResolution       int  `json:"promotions_after_resolution"`
	UnknownCoordinatesPreserved     int  `json:"unknown_coordinates_preserved"`
	ClaimTransitionsPreserved       int  `json:"claim_transitions_preserved"`
	ReceiptsVerified                int  `json:"receipts_verified"`
	DeterministicReplays            int  `json:"deterministic_replays"`
	SourceReconstructionNumerator   int  `json:"source_reconstruction_numerator"`
	SourceReconstructionDenominator int  `json:"source_reconstruction_denominator"`
	ProducerImportNumerator         int  `json:"producer_import_numerator"`
	ProducerImportDenominator       int  `json:"producer_import_denominator"`
	RepositoryWrites                int  `json:"repository_writes"`
	MutationAuthority               bool `json:"mutation_authority"`
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
	WorkspaceEffects     Effects
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

type InterventionSide struct {
	SourceDigest            string `json:"source_digest"`
	SemanticDigest          string `json:"semantic_digest"`
	Rule                    string `json:"rule"`
	Decision                string `json:"decision"`
	Resolution              string `json:"resolution"`
	FirstCounterexampleID   string `json:"first_counterexample_id"`
	CounterexamplesObserved int    `json:"counterexamples_observed"`
	ClaimTransitionDigest   string `json:"claim_transition_digest"`
}

type InterventionReport struct {
	Schema               string `json:"schema"`
	SemanticIntervention struct {
		Before                 InterventionSide `json:"before"`
		After                  InterventionSide `json:"after"`
		SemanticDigestEqual    bool             `json:"semantic_digest_equal"`
		DecisionChanged        bool             `json:"decision_changed"`
		CounterexampleChanged  bool             `json:"counterexample_changed"`
		ClaimTransitionChanged bool             `json:"claim_transition_changed"`
	} `json:"semantic_intervention"`
	CommentOnlyIntervention struct {
		Before               InterventionSide `json:"before"`
		After                InterventionSide `json:"after"`
		SemanticDigestEqual  bool             `json:"semantic_digest_equal"`
		DecisionEqual        bool             `json:"decision_equal"`
		CounterexampleEqual  bool             `json:"counterexample_equal"`
		ClaimTransitionEqual bool             `json:"claim_transition_equal"`
	} `json:"comment_only_intervention"`
	Digest string `json:"digest"`
}

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, ID: ContractID, Version: 2,
		SourcePath: "examples/counterexample-first-compiler/main.gooo",
		Package:    "counterexamplefirst", Namespace: "counterexamplefirst",
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID,
		Predicate: PredicateSpec{
			ID:          "canonical-entity-id",
			Operation:   "syntax.ParseFile->bidir.Lower",
			Rule:        RuleIdentityV1,
			SourceFact:  "activity CanonicalEntityID computes value program",
			FailureRule: "lowered Entity.ID != canonical identity(namespace, kind, name)",
		},
		Fixed: FixedDenominator{Version: DenominatorVersion, Cases: CaseCount,
			Indicators: IndicatorCount, ClaimTransitions: TransitionCount,
			UnknownCoordinates: 1, CorpusInputs: CaseCount},
		Cases: []CaseSpec{
			{ID: "resolved-minimal-counterexample", InputKind: "mutated-source", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "promote-after-resolution"},
			{ID: "canonical-control", InputKind: "canonical-source", ProofChoice: "COUNTEREXAMPLE_REQUIRED", MetaOperation: "require-counterexample-before-compile"},
			{ID: "unresolved-counterexample", InputKind: "mutated-source", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "block-unresolved-counterexample"},
			{ID: "comment-only-control", InputKind: "comment-only-source", ProofChoice: "COMMENT_ONLY_INVARIANCE", MetaOperation: "preserve-semantic-digest"},
			{ID: "unobserved-input", InputKind: "unobserved-source", ProofChoice: "UNKNOWN_PRESERVATION", MetaOperation: "preserve-unknown-coordinate"},
		},
		NotClaimed: []string{"general compiler correctness", "global minimality", "theorem proving", "unbounded corpus coverage", "repository mutation"},
	}
}

// ValidContract checks only the structural contract. It deliberately does not
// compare any expected decision table because outcomes belong to observations.
func ValidContract(value Contract) bool {
	want := CanonicalContract()
	if value.Schema != want.Schema || value.ID != want.ID || value.Version != want.Version ||
		value.SourcePath != want.SourcePath || value.Package != want.Package || value.Namespace != want.Namespace ||
		value.Producer != want.Producer || value.Consumer != want.Consumer || value.MetaOperation != want.MetaOperation ||
		value.Predicate != want.Predicate || value.Fixed != want.Fixed || len(value.Cases) != len(want.Cases) {
		return false
	}
	for index, got := range value.Cases {
		if got != want.Cases[index] {
			return false
		}
	}
	return len(value.NotClaimed) == len(want.NotClaimed)
}
