package counterexamplefirst

const (
	ContractSchema     = "gooo/counterexample-first-compiler-contract/v3"
	ReceiptSchema      = "gooo/counterexample-first-compiler-receipt/v3"
	ReportSchema       = "gooo/counterexample-first-judge-report/v3"
	CorpusSchema       = "gooo/counterexample-first-scenarios/v3"
	DenominatorVersion = "counterexample-first-denominator/v3"
	ContractID         = "counterexample-first-meta-compilation-v3"
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
	UniqueClaims       int    `json:"unique_claims"`
	UniquePredicates   int    `json:"unique_predicates"`
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
	Predicates    []PredicateSpec  `json:"predicates"`
	Fixed         FixedDenominator `json:"fixed_denominator"`
	Cases         []CaseSpec       `json:"cases"`
	NotClaimed    []string         `json:"not_claimed"`
}

// PredicateSpec is a distinct observation predicate. A predicate is not an
// expected result: its pass/fail/unknown value is computed from execution.
type PredicateSpec struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Operation   string `json:"operation"`
	Rule        string `json:"rule"`
	SourceFact  string `json:"source_fact"`
	FailureRule string `json:"failure_rule"`
}

type CaseSpec struct {
	ID            string `json:"id"`
	ClaimID       string `json:"claim_id"`
	Proposition   string `json:"proposition"`
	PredicateID   string `json:"predicate_id"`
	InputKind     string `json:"input_kind"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

// ScenarioCorpus contains only observation inputs. Conclusions are generated
// from parser/lowerer observations and never supplied by this wire format.
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
	ID          string  `json:"id"`
	ClaimID     string  `json:"claim_id"`
	PredicateID string  `json:"predicate_id"`
	Claim       string  `json:"claim"`
	Source      *string `json:"source"`
}

// ResolutionInput names a repair operation. It cannot smuggle an unrelated
// passing source into the proof; the compiler derives repair source from the
// observed minimal counterexample.
type ResolutionInput struct {
	ID        string `json:"id"`
	Operation string `json:"operation"`
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

type FactObservation struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Status    string `json:"status"`
}

type GraphEdgeObservation struct {
	From    string `json:"from"`
	Through string `json:"through"`
	To      string `json:"to"`
}

// MetaOperationObservation is reconstructed from lowered nodes and facts.
// The four activities and their data edges are execution authority for the
// counterexample, resolution, and promotion stages.
type MetaOperationObservation struct {
	RequiredActivities []string               `json:"required_activities"`
	ActivityOrder      []string               `json:"activity_order"`
	Edges              []GraphEdgeObservation `json:"edges"`
	ActivitiesPresent  bool                   `json:"activities_present"`
	Connected          bool                   `json:"connected"`
	Authorized         bool                   `json:"authorized"`
	Reason             string                 `json:"reason"`
	Digest             string                 `json:"digest"`
}

// ExecutionObservation is a projection of actual ParseFile -> Lower output.
// It is evidence, not an input assertion.
type ExecutionObservation struct {
	InputID          string                   `json:"input_id"`
	SourceDigest     string                   `json:"source_digest"`
	SourceBytes      int                      `json:"source_bytes"`
	ParseDiagnostics []DiagnosticObservation  `json:"parse_diagnostics"`
	ParseOK          bool                     `json:"parse_ok"`
	LowerOK          bool                     `json:"lower_ok"`
	LowerError       string                   `json:"lower_error,omitempty"`
	SemanticDigest   string                   `json:"semantic_digest,omitempty"`
	Nodes            []NodeObservation        `json:"nodes,omitempty"`
	Facts            []FactObservation        `json:"facts,omitempty"`
	MetaOperation    MetaOperationObservation `json:"meta_operation"`
	OutputDigest     string                   `json:"output_digest"`
}

type PredicateObservation struct {
	PredicateID       string `json:"predicate_id"`
	Rule              string `json:"rule"`
	Applicable        bool   `json:"applicable"`
	ViolationObserved bool   `json:"violation_observed"`
	PassObserved      bool   `json:"pass_observed"`
	UnknownObserved   bool   `json:"unknown_observed"`
	Reason            string `json:"reason"`
	EvidenceDigest    string `json:"evidence_digest"`
}

type ShrinkObservation struct {
	CandidateDigest string               `json:"candidate_digest"`
	SourceBytes     int                  `json:"source_bytes"`
	Observation     ExecutionObservation `json:"observation"`
	Predicate       PredicateObservation `json:"predicate"`
}

type Counterexample struct {
	ID                            string               `json:"id"`
	Source                        string               `json:"source"`
	SourceDigest                  string               `json:"source_digest"`
	SourceBytes                   int                  `json:"source_bytes"`
	Observation                   ExecutionObservation `json:"observation"`
	Predicate                     PredicateObservation `json:"predicate"`
	ShrinkTrace                   []ShrinkObservation  `json:"shrink_trace"`
	FiniteNeighborhoodNumerator   int                  `json:"finite_neighborhood_numerator"`
	FiniteNeighborhoodDenominator int                  `json:"finite_neighborhood_denominator"`
	FiniteNeighborhoodIrreducible bool                 `json:"finite_neighborhood_irreducible"`
	Stage                         string               `json:"stage"`
	Step                          string               `json:"step"`
	Reason                        string               `json:"reason"`
}

type ResolutionEvidence struct {
	ID                   string               `json:"id"`
	CounterexampleID     string               `json:"counterexample_id"`
	InputID              string               `json:"input_id"`
	OriginalSourceDigest string               `json:"original_source_digest"`
	RepairSourceDigest   string               `json:"repair_source_digest"`
	RepairDeltaDigest    string               `json:"repair_delta_digest"`
	RepairOperation      string               `json:"repair_operation"`
	SameClaim            bool                 `json:"same_claim"`
	SamePredicate        bool                 `json:"same_predicate"`
	ClaimID              string               `json:"claim_id"`
	PropositionDigest    string               `json:"proposition_digest"`
	PredicateID          string               `json:"predicate_id"`
	Observation          ExecutionObservation `json:"observation"`
	Predicate            PredicateObservation `json:"predicate"`
	Stage                string               `json:"stage"`
	Step                 string               `json:"step"`
	Reason               string               `json:"reason"`
	ProofChoice          string               `json:"proof_choice"`
	MetaOperation        string               `json:"meta_operation"`
	Producer             string               `json:"producer"`
	Consumer             string               `json:"consumer"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type DecisionInput struct {
	ClaimID               string `json:"claim_id"`
	PropositionDigest     string `json:"proposition_digest"`
	PredicateID           string `json:"predicate_id"`
	CandidateID           string `json:"candidate_id"`
	CandidateDigest       string `json:"candidate_digest"`
	CounterexampleID      string `json:"counterexample_id"`
	CounterexampleDigest  string `json:"counterexample_digest"`
	ResolutionID          string `json:"resolution_id"`
	ResolutionDigest      string `json:"resolution_digest"`
	RepairDeltaDigest     string `json:"repair_delta_digest"`
	RequiredBeforeCompile bool   `json:"required_before_compile"`
}

// ClaimTransition is append-only evidence. A counterexample first moves OPEN
// to REFUTED; only an observed repair rerun can append REFUTED to DISCHARGED.
type ClaimTransition struct {
	Sequence          int    `json:"sequence"`
	ClaimID           string `json:"claim_id"`
	PropositionDigest string `json:"proposition_digest"`
	PredicateID       string `json:"predicate_id"`
	From              string `json:"from"`
	To                string `json:"to"`
	Status            string `json:"status"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	EvidenceDigest    string `json:"evidence_digest"`
	PredicateDigest   string `json:"predicate_digest"`
}

type Effects struct {
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthority  string `json:"mutation_authority"`
	CapabilityEvidence string `json:"capability_evidence"`
}

type DecisionReceipt struct {
	Schema               string                   `json:"schema"`
	ContractID           string                   `json:"contract_id"`
	HeadSHA              string                   `json:"head_sha"`
	SourcePath           string                   `json:"source_path"`
	SourceDigest         string                   `json:"source_digest"`
	SemanticDigest       string                   `json:"semantic_digest"`
	ProgramMetaOperation MetaOperationObservation `json:"program_meta_operation"`
	ScenarioID           string                   `json:"scenario_id"`
	ClaimID              string                   `json:"claim_id"`
	PropositionDigest    string                   `json:"proposition_digest"`
	PredicateID          string                   `json:"predicate_id"`
	Producer             string                   `json:"producer"`
	Consumer             string                   `json:"consumer"`
	MetaOperation        string                   `json:"meta_operation"`
	ProofChoice          string                   `json:"proof_choice"`
	Decision             string                   `json:"decision"`
	Resolution           string                   `json:"resolution"`
	Reason               string                   `json:"reason"`
	Coordinate           Coordinate               `json:"coordinate"`
	CandidateObservation ExecutionObservation     `json:"candidate_observation"`
	CandidatePredicate   PredicateObservation     `json:"candidate_predicate"`
	Counterexample       *Counterexample          `json:"counterexample,omitempty"`
	ResolutionEvidence   *ResolutionEvidence      `json:"resolution_evidence,omitempty"`
	DecisionInput        DecisionInput            `json:"decision_input"`
	ClaimTransitions     []ClaimTransition        `json:"claim_transitions"`
	Effects              Effects                  `json:"effects"`
	Digest               string                   `json:"digest"`
}

type Summary struct {
	ReceiptsReconstructed           int    `json:"receipts_reconstructed"`
	CasesTotal                      int    `json:"cases_total"`
	CorpusExecutions                int    `json:"corpus_executions"`
	DiscoveredCounterexamples       int    `json:"discovered_counterexamples"`
	ShrinkCandidateExecutions       int    `json:"shrink_candidate_executions"`
	FiniteNeighborhoodNumerator     int    `json:"finite_neighborhood_numerator"`
	FiniteNeighborhoodDenominator   int    `json:"finite_neighborhood_denominator"`
	ResolutionReruns                int    `json:"resolution_reruns"`
	PromotionsAfterResolution       int    `json:"promotions_after_resolution"`
	UnknownCoordinatesPreserved     int    `json:"unknown_coordinates_preserved"`
	ClaimTransitionsPreserved       int    `json:"claim_transitions_preserved"`
	ReceiptsVerified                int    `json:"receipts_verified"`
	DeterministicReplays            int    `json:"deterministic_replays"`
	UniqueClaimsObserved            int    `json:"unique_claims_observed"`
	UniquePredicatesObserved        int    `json:"unique_predicates_observed"`
	SourceReconstructionNumerator   int    `json:"source_reconstruction_numerator"`
	SourceReconstructionDenominator int    `json:"source_reconstruction_denominator"`
	ProducerImportNumerator         int    `json:"producer_import_numerator"`
	ProducerImportDenominator       int    `json:"producer_import_denominator"`
	RepositoryWrites                int    `json:"repository_writes"`
	MutationAuthority               string `json:"mutation_authority"`
	CapabilityEvidence              string `json:"capability_evidence"`
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

type CaseIntervention struct {
	CaseID                  string `json:"case_id"`
	ClaimID                 string `json:"claim_id"`
	PropositionDigest       string `json:"proposition_digest"`
	PredicateID             string `json:"predicate_id"`
	TargetOperation         string `json:"target_operation"`
	SourceDigest            string `json:"source_digest"`
	SemanticDigest          string `json:"semantic_digest"`
	Decision                string `json:"decision"`
	Resolution              string `json:"resolution"`
	TransitionDigest        string `json:"transition_digest"`
	CounterexampleID        string `json:"counterexample_id"`
	MetaOperationAuthorized bool   `json:"meta_operation_authorized"`
}

type InterventionSide struct {
	SourceDigest            string             `json:"source_digest"`
	SemanticDigest          string             `json:"semantic_digest"`
	Rule                    string             `json:"rule"`
	MetaOperationAuthorized bool               `json:"meta_operation_authorized"`
	Cases                   []CaseIntervention `json:"cases"`
	CaseTransitionDigest    string             `json:"case_transition_digest"`
}

type InterventionComparison struct {
	Before                 InterventionSide `json:"before"`
	After                  InterventionSide `json:"after"`
	SemanticDigestEqual    bool             `json:"semantic_digest_equal"`
	DecisionChanged        bool             `json:"decision_changed"`
	CounterexampleChanged  bool             `json:"counterexample_changed"`
	ClaimTransitionChanged bool             `json:"claim_transition_changed"`
	AllCasesAddressed      bool             `json:"all_cases_addressed"`
}

type InterventionReport struct {
	Schema                    string                 `json:"schema"`
	SemanticIntervention      InterventionComparison `json:"semantic_intervention"`
	CommentOnlyIntervention   InterventionComparison `json:"comment_only_intervention"`
	MetaOperationIntervention InterventionComparison `json:"meta_operation_intervention"`
	Digest                    string                 `json:"digest"`
}

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, ID: ContractID, Version: 3,
		SourcePath: "examples/counterexample-first-compiler/main.gooo",
		Package:    "counterexamplefirst", Namespace: "counterexamplefirst",
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID,
		Predicates: []PredicateSpec{
			{ID: "identity-drift-detected", Kind: "ENTITY_ID_DRIFT", Operation: "syntax.ParseFile->bidir.Lower", Rule: RuleIdentityV1, SourceFact: "lowered Entity.ID", FailureRule: "entity ID mismatch emits a counterexample"},
			{ID: "canonical-source-admissible", Kind: "CANONICAL_SOURCE", Operation: "syntax.ParseFile->bidir.Lower", Rule: RuleIdentityV1, SourceFact: "lowered Entity.ID", FailureRule: "entity ID mismatch prevents canonical-source admissibility"},
			{ID: "resolution-required", Kind: "RESOLUTION_REQUIRED", Operation: "syntax.ParseFile->bidir.Lower", Rule: RuleIdentityV1, SourceFact: "counterexample predicate", FailureRule: "observed identity drift requires a repair rerun"},
			{ID: "semantic-digest-invariant", Kind: "SEMANTIC_DIGEST", Operation: "syntax.ParseFile->bidir.Lower", Rule: RuleIdentityV1, SourceFact: "semantic IR stable hash", FailureRule: "candidate semantic digest differs from the baseline"},
			{ID: "source-acquisition-present", Kind: "SOURCE_ACQUISITION", Operation: "syntax.ParseFile->bidir.Lower", Rule: RuleIdentityV1, SourceFact: "raw corpus source", FailureRule: "source absence is not an observable execution"},
		},
		Fixed: FixedDenominator{Version: DenominatorVersion, Cases: CaseCount, UniqueClaims: CaseCount, UniquePredicates: CaseCount, Indicators: IndicatorCount, ClaimTransitions: TransitionCount, UnknownCoordinates: 1, CorpusInputs: CaseCount},
		Cases: []CaseSpec{
			{ID: "resolved-minimal-counterexample", ClaimID: "claim-resolved-repair", Proposition: "the candidate identity drift can be repaired by canonicalizing the same minimal source", PredicateID: "identity-drift-detected", InputKind: "mutated-source", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "promote-after-resolution"},
			{ID: "canonical-control", ClaimID: "claim-canonical-control", Proposition: "a canonical candidate has no observed identity violation but is insufficient evidence for promotion", PredicateID: "canonical-source-admissible", InputKind: "canonical-source", ProofChoice: "COUNTEREXAMPLE_REQUIRED", MetaOperation: "require-counterexample-before-compile"},
			{ID: "unresolved-counterexample", ClaimID: "claim-unresolved-boundary", Proposition: "an identity violation without repair evidence remains refuted", PredicateID: "resolution-required", InputKind: "mutated-source", ProofChoice: "COUNTEREXAMPLE_RESOLUTION", MetaOperation: "block-unresolved-counterexample"},
			{ID: "comment-only-control", ClaimID: "claim-comment-invariance", Proposition: "comment-only source changes preserve semantic lowering and predicate evidence", PredicateID: "semantic-digest-invariant", InputKind: "comment-only-source", ProofChoice: "COMMENT_ONLY_INVARIANCE", MetaOperation: "preserve-semantic-digest"},
			{ID: "unobserved-input", ClaimID: "claim-source-acquisition", Proposition: "absent source acquisition retains UNKNOWN rather than inferring a decision", PredicateID: "source-acquisition-present", InputKind: "unobserved-source", ProofChoice: "UNKNOWN_PRESERVATION", MetaOperation: "preserve-unknown-coordinate"},
		},
		NotClaimed: []string{"general compiler correctness", "global or cost minimality", "theorem proving", "unbounded corpus coverage", "repository mutation"},
	}
}

// ValidContract checks only structural inputs and the fixed experiment shape;
// it does not compare any expected decision table because outcomes belong to
// observations.
func ValidContract(value Contract) bool {
	want := CanonicalContract()
	if value.Schema != want.Schema || value.ID != want.ID || value.Version != want.Version || value.SourcePath != want.SourcePath || value.Package != want.Package || value.Namespace != want.Namespace || value.Producer != want.Producer || value.Consumer != want.Consumer || value.MetaOperation != want.MetaOperation || value.Fixed != want.Fixed || len(value.Predicates) != len(want.Predicates) || len(value.Cases) != len(want.Cases) {
		return false
	}
	for index, got := range value.Predicates {
		if got != want.Predicates[index] {
			return false
		}
	}
	for index, got := range value.Cases {
		if got != want.Cases[index] {
			return false
		}
	}
	return len(value.NotClaimed) == len(want.NotClaimed)
}
