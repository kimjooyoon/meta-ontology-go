package observereffect

const (
	LedgerSchema     = "gooo/observer-effect-ledger/v1"
	ReceiptSchema    = "gooo/observer-effect-receipt/v1"
	JudgmentSchema   = "gooo/observer-effect-judgment/v1"
	ExperimentName   = "meta-observer-self-effect"
	FixedDenominator = 12
)

type Source struct {
	Path              string                 `json:"path"`
	Digest            string                 `json:"digest"`
	GoooSource        bool                   `json:"gooo_source"`
	CanonicalParse    bool                   `json:"canonical_parse"`
	CanonicalLowering bool                   `json:"canonical_lowering"`
	SemanticDigest    string                 `json:"semantic_digest"`
	Interventions     []SemanticIntervention `json:"interventions"`
}

type SnapshotDelta struct {
	BeforeDigest   string `json:"before_digest"`
	AfterDigest    string `json:"after_digest"`
	Changed        bool   `json:"changed"`
	BeforeObserved bool   `json:"before_observed"`
	AfterObserved  bool   `json:"after_observed"`
	Status         string `json:"status"`
	Resolution     string `json:"resolution"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type Observation struct {
	RepositoryStorage SnapshotDelta `json:"repository_storage"`
	Environment       SnapshotDelta `json:"environment"`
	LogicalTime       SnapshotDelta `json:"logical_time"`
}

type Effect struct {
	Domain            string `json:"domain"`
	ObservedChanged   bool   `json:"observed_changed"`
	MutationAttempted bool   `json:"mutation_attempted"`
	Planned           bool   `json:"planned"`
	BeforeDigest      string `json:"before_digest"`
	AfterDigest       string `json:"after_digest"`
	WriteCount        int    `json:"write_count"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	Status            string `json:"status"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
}

type SemanticIntervention struct {
	Name              string `json:"name"`
	Mutation          string `json:"mutation"`
	ParseValid        bool   `json:"parse_valid"`
	LoweringValid     bool   `json:"lowering_valid"`
	BaselineDigest    string `json:"baseline_digest"`
	MutatedDigest     string `json:"mutated_digest"`
	SemanticInvariant bool   `json:"semantic_invariant"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
}

type CoordinateAdjudication struct {
	Coordinate     string `json:"coordinate"`
	Status         string `json:"status"`
	Resolution     string `json:"resolution"`
	BeforeObserved bool   `json:"before_observed"`
	AfterObserved  bool   `json:"after_observed"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
}

type Unknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	ClaimID                string `json:"claim_id"`
	Persistent             bool   `json:"persistent"`
	Sequence               int    `json:"sequence"`
	PreviousState          string `json:"previous_state"`
	CurrentState           string `json:"current_state"`
	Transition             string `json:"transition"`
	PreviousEvidenceDigest string `json:"previous_evidence_digest"`
	EvidenceDigest         string `json:"evidence_digest"`
}

type CausalEdge struct {
	ID            string `json:"id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Relation      string `json:"relation"`
	Before        int    `json:"before"`
	After         int    `json:"after"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type TopologySubscriber struct {
	Workflow      string `json:"workflow"`
	Upstream      string `json:"upstream"`
	Expected      string `json:"expected"`
	Actual        string `json:"actual"`
	Concurrency   string `json:"concurrency"`
	Status        string `json:"status"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type TopologyEvidence struct {
	Scope                                           string               `json:"scope"`
	WorkflowRunSubscribersAudited                   int                  `json:"workflow_run_subscribers_audited"`
	WorkflowRunSubscribersExpected                  int                  `json:"workflow_run_subscribers_expected"`
	BranchFilteredSubscribersBefore                 int                  `json:"branch_filtered_subscribers_before"`
	BranchFilteredSubscribers                       int                  `json:"branch_filtered_subscribers"`
	BranchFilteredSubscribersExpected               int                  `json:"branch_filtered_subscribers_expected"`
	DuplicatePROObservationPathsBefore              int                  `json:"duplicate_pr_observation_paths_before"`
	DuplicatePROObservationPathsAfter               int                  `json:"duplicate_pr_observation_paths_after"`
	ExpectedSkippedCIChildRunsPerPRCompletionBefore int                  `json:"expected_skipped_ci_child_run_objects_per_pr_completion_before"`
	ExpectedSkippedCIChildRunsPerPRCompletionAfter  int                  `json:"expected_skipped_ci_child_run_objects_per_pr_completion_after"`
	Subscribers                                     []TopologySubscriber `json:"subscribers"`
	CausalEdges                                     []CausalEdge         `json:"causal_edges"`
	Exact                                           bool                 `json:"exact"`
	Producer                                        string               `json:"producer"`
	Consumer                                        string               `json:"consumer"`
	MetaOperation                                   string               `json:"meta_operation"`
	ProofChoice                                     string               `json:"proof_choice"`
	Stage                                           string               `json:"stage"`
	Step                                            string               `json:"step"`
	Reason                                          string               `json:"reason"`
}

type RunnerScopedEvidence struct {
	Scope                      string `json:"scope"`
	Classification             string `json:"classification"`
	Status                     string `json:"status"`
	Source                     string `json:"source"`
	ObservationRef             string `json:"observation_ref"`
	ObservedAt                 string `json:"observed_at"`
	Query                      string `json:"query"`
	SubjectSHA                 string `json:"subject_sha"`
	EvidenceDigest             string `json:"evidence_digest"`
	SkippedWorkflowRuns        int    `json:"skipped_workflow_runs"`
	QueuedWorkflowRuns         int    `json:"queued_workflow_runs"`
	TimeDependent              bool   `json:"time_dependent"`
	CurrentEvidence            bool   `json:"current_evidence"`
	IncludedInFixedDenominator bool   `json:"included_in_fixed_denominator"`
	Producer                   string `json:"producer"`
	Consumer                   string `json:"consumer"`
	MetaOperation              string `json:"meta_operation"`
	ProofChoice                string `json:"proof_choice"`
	Stage                      string `json:"stage"`
	Step                       string `json:"step"`
	Reason                     string `json:"reason"`
}

type GuardianExpectation struct {
	Scope                      string `json:"scope"`
	Code                       string `json:"code"`
	ExpectedDecision           string `json:"expected_decision"`
	ExpectedRoute              string `json:"expected_route"`
	RequiredContext            bool   `json:"required_context"`
	IncludedInFixedDenominator bool   `json:"included_in_fixed_denominator"`
	Producer                   string `json:"producer"`
	Consumer                   string `json:"consumer"`
	MetaOperation              string `json:"meta_operation"`
	ProofChoice                string `json:"proof_choice"`
	Stage                      string `json:"stage"`
	Step                       string `json:"step"`
	Reason                     string `json:"reason"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Expected      string `json:"expected"`
	Actual        string `json:"actual"`
	Status        string `json:"status"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type Metrics struct {
	FixedDenominator     int `json:"fixed_denominator"`
	Satisfied            int `json:"satisfied"`
	CoverageBasisPoints  int `json:"coverage_basis_points"`
	ObservationSatisfied int `json:"observation_satisfied"`
	ObservationTotal     int `json:"observation_total"`
	EffectSatisfied      int `json:"effect_satisfied"`
	EffectTotal          int `json:"effect_total"`
	GuardrailSatisfied   int `json:"guardrail_satisfied"`
	GuardrailTotal       int `json:"guardrail_total"`
}

type Authority struct {
	RepositoryWrites    int  `json:"repository_writes"`
	OutputWrites        int  `json:"output_writes"`
	MutationAuthority   bool `json:"mutation_authority"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type Receipt struct {
	Schema            string          `json:"schema"`
	Kind              string          `json:"kind"`
	Producer          string          `json:"producer"`
	Consumer          string          `json:"consumer"`
	MetaOperation     string          `json:"meta_operation"`
	ProofChoice       string          `json:"proof_choice"`
	Subject           string          `json:"subject"`
	SubjectDigest     string          `json:"subject_digest"`
	Decision          string          `json:"decision"`
	Resolution        string          `json:"resolution"`
	RepositoryWrites  int             `json:"repository_writes"`
	MutationAuthority bool            `json:"mutation_authority"`
	Unknown           Unknown         `json:"unknown"`
	Coordinate        Unknown         `json:"coordinate"`
	Reason            string          `json:"reason"`
	ClaimTransition   ClaimTransition `json:"claim_transition"`
	EvidenceDigest    string          `json:"evidence_digest"`
	Digest            string          `json:"digest"`
}

type Report struct {
	Schema              string                   `json:"schema"`
	Experiment          string                   `json:"experiment"`
	Source              Source                   `json:"source"`
	Observation         Observation              `json:"observation"`
	Effects             []Effect                 `json:"effects"`
	ReceiptDigests      []string                 `json:"receipt_digests"`
	Unknown             Unknown                  `json:"unknown"`
	Coordinate          Unknown                  `json:"coordinate"`
	Reason              string                   `json:"reason"`
	ClaimTransition     ClaimTransition          `json:"claim_transition"`
	Coordinates         []CoordinateAdjudication `json:"coordinates"`
	Topology            TopologyEvidence         `json:"topology"`
	RunnerScoped        RunnerScopedEvidence     `json:"runner_scoped"`
	Guardian            GuardianExpectation      `json:"guardian_expectation"`
	Metrics             Metrics                  `json:"metrics"`
	Authority           Authority                `json:"authority"`
	RepositoryWrites    int                      `json:"repository_writes"`
	MutationAuthority   bool                     `json:"mutation_authority"`
	PromotionAuthorized bool                     `json:"promotion_authorized"`
	Decision            string                   `json:"decision"`
	Resolution          string                   `json:"resolution"`
	EvidenceDigest      string                   `json:"evidence_digest"`
	Digest              string                   `json:"digest"`
	Indicators          []Indicator              `json:"indicators"`
}
