package ambiguitybudget

const (
	ContractSchema           = "gooo/ambiguity-budget-contract/v3"
	ReceiptSchema            = "gooo/ambiguity-budget-receipt/v3"
	PolicySchema             = "gooo/ambiguity-budget-policy/v1"
	DenominatorSchema        = "gooo/ambiguity-budget-denominator/v1"
	EffectsSchema            = "gooo/ambiguity-budget-workspace-effects/v1"
	Producer                 = "gooo://meta/ambiguity-budget/producer"
	Consumer                 = "gooo://meta/ambiguity-budget/independent-verifier"
	MetaOperation            = "measure-deterministic-ambiguity-budget"
	FoundationProof          = "FOUNDATION"
	CoherenceProof           = "COHERENCE"
	RegressionProof          = "REGRESSION"
	ExpectedCaseTotal        = 4
	ExpectedInterventions    = 2
	IntegerDimensions        = 3
	ExpectedSourceEntities   = 30
	ExpectedSourceActivities = 30
)

var integerDimensions = [...]string{"interpretation_candidates", "unresolved_branches", "evidence_paths"}

type IntegerSet struct {
	InterpretationCandidates int `json:"interpretation_candidates"`
	UnresolvedBranches       int `json:"unresolved_branches"`
	EvidencePaths            int `json:"evidence_paths"`
}

type BudgetDimension struct {
	ID    string `json:"id"`
	Limit int    `json:"limit"`
}

// BudgetPolicy is authoritative in the versioned contract. The executable
// source only binds its FixedBudget activity to this policy version.
type BudgetPolicy struct {
	Schema     string            `json:"schema"`
	ID         string            `json:"id"`
	Version    string            `json:"version"`
	Authority  string            `json:"authority"`
	Dimensions []BudgetDimension `json:"dimensions"`
}

type Denominator struct {
	Schema                string `json:"schema"`
	Version               string `json:"version"`
	Cases                 int    `json:"cases"`
	IntegerObservations   int    `json:"integer_observations"`
	Claims                int    `json:"claims"`
	Interventions         int    `json:"interventions"`
	AuthorityObservations int    `json:"authority_observations"`
}

type Numerator struct {
	CasesConforming             int `json:"cases_conforming"`
	IntegerObservationsObserved int `json:"integer_observations_observed"`
	ClaimsDischarged            int `json:"claims_discharged"`
	ClaimsRefuted               int `json:"claims_refuted"`
	ClaimsOpen                  int `json:"claims_open"`
	InterventionsSatisfied      int `json:"interventions_satisfied"`
	AuthorityObserved           int `json:"authority_observed"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ObservationGap struct {
	Dimension      string     `json:"dimension"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
}

// AmbiguityElements are the semantic graph elements used to derive counts.
// IDs, not declared integers, are the observation surface.
type AmbiguityElements struct {
	CandidateIDs         []string `json:"candidate_ids"`
	ResolvedBranchIDs    []string `json:"resolved_branch_ids"`
	UnresolvedBranchIDs  []string `json:"unresolved_branch_ids"`
	EvidencePathIDs      []string `json:"evidence_path_ids"`
	BranchObservationIDs []string `json:"branch_observation_ids"`
}

// ClaimTransition contains lifecycle and proposition evidence only. Raw
// source provenance is kept beside it on the case/intervention receipt so a
// comment-only change cannot alter semantic claim evidence.
type ClaimTransition struct {
	CaseID            string `json:"case_id"`
	Proposition       string `json:"proposition"`
	PropositionDigest string `json:"proposition_digest"`
	From              string `json:"from"`
	To                string `json:"to"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
	EvidenceDigest    string `json:"evidence_digest"`
}

// CaseContract constrains identity only. Counts and outcomes are derived from
// lowered graph elements.
type CaseContract struct {
	ID       string `json:"id"`
	Activity string `json:"activity"`
}

type InterventionContract struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	TargetActivity string `json:"target_activity"`
}

type Contract struct {
	Schema          string                 `json:"schema"`
	ID              string                 `json:"id"`
	SourcePath      string                 `json:"source_path"`
	SourcePackage   string                 `json:"source_package"`
	SourceNamespace string                 `json:"source_namespace"`
	BudgetActivity  string                 `json:"budget_activity"`
	BudgetPolicy    BudgetPolicy           `json:"budget_policy"`
	Denominator     Denominator            `json:"denominator"`
	Cases           []CaseContract         `json:"cases"`
	Interventions   []InterventionContract `json:"interventions"`
	NotClaimed      []string               `json:"not_claimed"`
}

type Input struct {
	SubjectSHA      string
	Contract        Contract
	Source          []byte
	EffectsArtifact []byte
}

type ProgramObservation struct {
	Activity               string            `json:"activity"`
	Program                string            `json:"program"`
	ProgramKind            string            `json:"program_kind"`
	ID                     string            `json:"id"`
	Class                  string            `json:"class,omitempty"`
	InputState             string            `json:"input_state,omitempty"`
	Counts                 IntegerSet        `json:"counts"`
	UnobservedDimensions   []string          `json:"unobserved_dimensions,omitempty"`
	ObservationGaps        []ObservationGap  `json:"observation_gaps,omitempty"`
	Elements               AmbiguityElements `json:"elements"`
	ElementDigest          string            `json:"element_digest"`
	ActivitySemanticDigest string            `json:"activity_semantic_digest"`
	SemanticDigest         string            `json:"semantic_digest"`
	Digest                 string            `json:"digest"`
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
	ID                     string            `json:"id"`
	Activity               string            `json:"activity"`
	RawSourceDigest        string            `json:"raw_source_digest"`
	Class                  string            `json:"class"`
	InputState             string            `json:"input_state"`
	Program                string            `json:"program"`
	ProgramDigest          string            `json:"program_digest"`
	ProgramSemanticDigest  string            `json:"program_semantic_digest"`
	ActivitySemanticDigest string            `json:"activity_semantic_digest"`
	Elements               AmbiguityElements `json:"elements"`
	ElementDigest          string            `json:"element_digest"`
	Counts                 IntegerSet        `json:"counts"`
	UnobservedDimensions   []string          `json:"unobserved_dimensions,omitempty"`
	ObservationGaps        []ObservationGap  `json:"observation_gaps,omitempty"`
	Decision               string            `json:"decision"`
	Resolution             string            `json:"resolution"`
	Reason                 string            `json:"reason"`
	Coordinate             Coordinate        `json:"coordinate"`
	Proposition            string            `json:"proposition"`
	PropositionDigest      string            `json:"proposition_digest"`
	Claim                  ClaimTransition   `json:"claim"`
	EvidenceDigest         string            `json:"evidence_digest"`
	Conformance            string            `json:"conformance"`
}

type Indicator struct {
	MetricID           string `json:"metric_id"`
	CaseID             string `json:"case_id"`
	Dimension          string `json:"dimension"`
	ProofChoice        string `json:"proof_choice"`
	Producer           string `json:"producer"`
	Consumer           string `json:"consumer"`
	MetaOperation      string `json:"meta_operation"`
	Observed           int    `json:"observed"`
	CoordinateObserved bool   `json:"coordinate_observed"`
	Budget             int    `json:"budget"`
	Relation           string `json:"relation"`
	Evaluation         string `json:"evaluation"`
	EvidenceDigest     string `json:"evidence_digest"`
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
	ID                   string            `json:"id"`
	Kind                 string            `json:"kind"`
	TargetActivity       string            `json:"target_activity"`
	SourceDigestBefore   string            `json:"source_digest_before"`
	SourceDigestAfter    string            `json:"source_digest_after"`
	SemanticDigestBefore string            `json:"semantic_digest_before"`
	SemanticDigestAfter  string            `json:"semantic_digest_after"`
	ElementsBefore       AmbiguityElements `json:"elements_before"`
	ElementsAfter        AmbiguityElements `json:"elements_after"`
	CountsBefore         IntegerSet        `json:"counts_before"`
	CountsAfter          IntegerSet        `json:"counts_after"`
	UnobservedBefore     []string          `json:"unobserved_before,omitempty"`
	UnobservedAfter      []string          `json:"unobserved_after,omitempty"`
	ClassBefore          string            `json:"class_before"`
	ClassAfter           string            `json:"class_after"`
	InputStateBefore     string            `json:"input_state_before"`
	InputStateAfter      string            `json:"input_state_after"`
	ClaimBefore          ClaimTransition   `json:"claim_before"`
	ClaimAfter           ClaimTransition   `json:"claim_after"`
	DecisionBefore       string            `json:"decision_before"`
	ResolutionBefore     string            `json:"resolution_before"`
	ReasonBefore         string            `json:"reason_before"`
	DecisionAfter        string            `json:"decision_after"`
	ResolutionAfter      string            `json:"resolution_after"`
	ReasonAfter          string            `json:"reason_after"`
	Satisfied            bool              `json:"satisfied"`
	EvidenceDigest       string            `json:"evidence_digest"`
}

// WorkspaceEffects is supplied by CI after comparing tracked and untracked
// workspace snapshots. Authority is deliberately unknown without a separate
// capability observation.
type Effects struct {
	Schema                      string `json:"schema"`
	Version                     string `json:"version"`
	ArtifactDigest              string `json:"artifact_digest"`
	TrackedAndUntracked         bool   `json:"tracked_and_untracked"`
	SnapshotBeforeDigest        string `json:"snapshot_before"`
	SnapshotAfterDigest         string `json:"snapshot_after"`
	RepositoryWrites            int    `json:"repository_writes"`
	WriteSetEqual               bool   `json:"write_set_equal"`
	MutationAuthority           string `json:"mutation_authority"`
	MutationAuthorityResolution string `json:"mutation_authority_resolution"`
}

type Summary struct {
	CasesTotal           int         `json:"cases_total"`
	KnownCases           int         `json:"known_cases"`
	ZeroAmbiguityCases   int         `json:"zero_ambiguity_cases"`
	BoundaryCases        int         `json:"boundary_cases"`
	OverBudgetCases      int         `json:"over_budget_cases"`
	UnknownCases         int         `json:"unknown_cases"`
	LowerResolutionCases int         `json:"lower_resolution_cases"`
	OpenClaims           int         `json:"open_claims"`
	IntegerDimensions    int         `json:"integer_dimensions"`
	Denominator          Denominator `json:"denominator"`
	Numerator            Numerator   `json:"numerator"`
}

type Receipt struct {
	Schema                string                `json:"schema"`
	SubjectSHA            string                `json:"subject_sha"`
	ContractID            string                `json:"contract_id"`
	Source                SourceObservation     `json:"source"`
	BudgetPolicy          BudgetPolicy          `json:"budget_policy"`
	BudgetBinding         string                `json:"budget_binding"`
	BudgetAuthority       string                `json:"budget_authority"`
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
