package ambiguitybudget

const (
	ContractSchema = "gooo/ambiguity-budget-contract/v2"
	ReceiptSchema  = "gooo/ambiguity-budget-receipt/v2"

	Producer              = "gooo://meta/ambiguity-budget/producer"
	Consumer              = "gooo://meta/ambiguity-budget/independent-verifier"
	MetaOperation         = "measure-deterministic-ambiguity-budget"
	FoundationProof       = "FOUNDATION"
	CoherenceProof        = "COHERENCE"
	RegressionProof       = "REGRESSION"
	ExpectedCaseTotal     = 4
	ExpectedInterventions = 2
	IntegerDimensions     = 3
	// FixedDenominator counts the two declared interventions. It is deliberately
	// independent of the twelve integer coordinates and is not a score.
	FixedDenominator = ExpectedInterventions
)

type IntegerSet struct {
	InterpretationCandidates int `json:"interpretation_candidates"`
	UnresolvedBranches       int `json:"unresolved_branches"`
	EvidencePaths            int `json:"evidence_paths"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	CaseID         string `json:"case_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

// CaseContract constrains only stable identity. Counts, classes, decisions,
// and reasons are observed from the executable computes declarations.
type CaseContract struct {
	ID       string `json:"id"`
	Activity string `json:"activity"`
}

// InterventionContract constrains only intervention identity and target. The
// before/after integer sets are never supplied by policy.
type InterventionContract struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	TargetActivity string `json:"target_activity"`
}

type Contract struct {
	Schema           string                 `json:"schema"`
	ID               string                 `json:"id"`
	SourcePath       string                 `json:"source_path"`
	SourcePackage    string                 `json:"source_package"`
	SourceNamespace  string                 `json:"source_namespace"`
	BudgetActivity   string                 `json:"budget_activity"`
	FixedDenominator int                    `json:"fixed_denominator"`
	Cases            []CaseContract         `json:"cases"`
	Interventions    []InterventionContract `json:"interventions"`
	NotClaimed       []string               `json:"not_claimed"`
}

type Input struct {
	SubjectSHA string
	Contract   Contract
	Source     []byte
}

type ProgramObservation struct {
	Activity    string     `json:"activity"`
	Program     string     `json:"program"`
	ProgramKind string     `json:"program_kind"`
	ID          string     `json:"id"`
	Class       string     `json:"class,omitempty"`
	InputState  string     `json:"input_state,omitempty"`
	Counts      IntegerSet `json:"counts"`
	Digest      string     `json:"digest"`
}

type SourceObservation struct {
	Path           string               `json:"path"`
	Digest         string               `json:"digest"`
	SemanticDigest string               `json:"semantic_digest"`
	Lowering       string               `json:"lowering"`
	Package        string               `json:"package"`
	Namespace      string               `json:"namespace"`
	Entities       int                  `json:"entities"`
	Activities     int                  `json:"activities"`
	Programs       []ProgramObservation `json:"programs"`
}

type CaseReceipt struct {
	ID             string          `json:"id"`
	Activity       string          `json:"activity"`
	Class          string          `json:"class"`
	InputState     string          `json:"input_state"`
	Program        string          `json:"program"`
	ProgramDigest  string          `json:"program_digest"`
	Counts         IntegerSet      `json:"counts"`
	Decision       string          `json:"decision"`
	Resolution     string          `json:"resolution"`
	Reason         string          `json:"reason"`
	Coordinate     Coordinate      `json:"coordinate"`
	Claim          ClaimTransition `json:"claim"`
	EvidenceDigest string          `json:"evidence_digest"`
	Conformance    string          `json:"conformance"`
}

type Indicator struct {
	MetricID       string `json:"metric_id"`
	CaseID         string `json:"case_id"`
	Dimension      string `json:"dimension"`
	ProofChoice    string `json:"proof_choice"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	Observed       int    `json:"observed"`
	Budget         int    `json:"budget"`
	Relation       string `json:"relation"`
	Evaluation     string `json:"evaluation"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type InterventionReceipt struct {
	ID                   string     `json:"id"`
	Kind                 string     `json:"kind"`
	TargetActivity       string     `json:"target_activity"`
	SourceDigestBefore   string     `json:"source_digest_before"`
	SourceDigestAfter    string     `json:"source_digest_after"`
	SemanticDigestBefore string     `json:"semantic_digest_before"`
	SemanticDigestAfter  string     `json:"semantic_digest_after"`
	CountsBefore         IntegerSet `json:"counts_before"`
	CountsAfter          IntegerSet `json:"counts_after"`
	DecisionBefore       string     `json:"decision_before"`
	ResolutionBefore     string     `json:"resolution_before"`
	ReasonBefore         string     `json:"reason_before"`
	DecisionAfter        string     `json:"decision_after"`
	ResolutionAfter      string     `json:"resolution_after"`
	ReasonAfter          string     `json:"reason_after"`
	Satisfied            bool       `json:"satisfied"`
	EvidenceDigest       string     `json:"evidence_digest"`
}

// Summary contains cardinalities and states only. It intentionally contains
// no percentage, basis-point, or aggregate score.
type Summary struct {
	CasesTotal           int `json:"cases_total"`
	KnownCases           int `json:"known_cases"`
	ZeroAmbiguityCases   int `json:"zero_ambiguity_cases"`
	BoundaryCases        int `json:"boundary_cases"`
	OverBudgetCases      int `json:"over_budget_cases"`
	UnknownCases         int `json:"unknown_cases"`
	LowerResolutionCases int `json:"lower_resolution_cases"`
	OpenClaims           int `json:"open_claims"`
	IntegerDimensions    int `json:"integer_dimensions"`
	InterventionsTotal   int `json:"interventions_total"`
	FixedDenominator     int `json:"fixed_denominator"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema                string                `json:"schema"`
	SubjectSHA            string                `json:"subject_sha"`
	ContractID            string                `json:"contract_id"`
	Source                SourceObservation     `json:"source"`
	Budget                IntegerSet            `json:"budget"`
	Producer              string                `json:"producer"`
	Consumer              string                `json:"consumer"`
	MetaOperation         string                `json:"meta_operation"`
	ProofChoice           string                `json:"proof_choice"`
	ConformanceDecision   string                `json:"conformance_decision"`
	ConformanceResolution string                `json:"conformance_resolution"`
	ConformanceReason     string                `json:"conformance_reason"`
	SubjectDecision       string                `json:"subject_decision"`
	SubjectResolution     string                `json:"subject_resolution"`
	SubjectReason         string                `json:"subject_reason"`
	SubjectCoordinate     Coordinate            `json:"subject_coordinate"`
	Cases                 []CaseReceipt         `json:"cases"`
	Claims                []ClaimTransition     `json:"claims"`
	Indicators            []Indicator           `json:"indicators"`
	Interventions         []InterventionReceipt `json:"interventions"`
	Proofs                []Proof               `json:"proofs"`
	Summary               Summary               `json:"summary"`
	NotClaimed            []string              `json:"not_claimed"`
	Effects               Effects               `json:"effects"`
	FactsDigest           string                `json:"facts_digest"`
	Digest                string                `json:"digest"`
}
