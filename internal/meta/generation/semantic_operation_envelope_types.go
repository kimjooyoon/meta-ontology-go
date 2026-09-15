package generation

// SemanticOperationEnvelopeSchema identifies the small, replayable operation
// boundary owned by the semantic layer.
const SemanticOperationEnvelopeSchema = "gooo/semantic-operation-envelope/v1"

// SemanticObservationSchema identifies the narrow compiler observation carried
// by the operation envelope. It records repeated semantic work; it does not
// assert that a reuse rule is safe or beneficial.
const SemanticObservationSchema = "gooo/semantic-self-observation/v1"

const (
	SemanticObservationActivity      = "semantic-activity"
	SemanticObservationPhase         = "semantic-lowering"
	SemanticObservationOperationID   = "bidir.lower.entity-fields-v1"
	SemanticObservationInputIdentity = "normalized-semantic-ir/sha256"
	SemanticObservationCandidateRule = "same-pure-operation-and-input-digest-count>1"
	SemanticObservationUnknownReason = "MISSING_EXACT_PAIR_EVIDENCE"
	SemanticObservationUnknownNext   = "CAPTURE_EXACT_BEHAVIOR_AND_DETERMINISM_PAIR"
	SemanticObservationUnknownClass  = "INCOMPLETE_EVIDENCE"
	SemanticObservationUnknownStage  = "COHERENCE"
	SemanticObservationUnknownStep   = "COMPARE_EXACT_PAIR"
	SemanticObservationContradiction = "OBSERVATION_CONTRADICTION"
)

const semanticOperationToolchainDigest = "semantic-operation-envelope-toolchain-v1"

var semanticOperationActivities = [...]string{
	"DeclareOperationIntent",
	"BindSourceRevision",
	"DeclareEffectGrant",
	"EmitEffectRequest",
	"RecordEffectResult",
	"VerifyReplayIdentity",
	"ClassifySemanticOutcome",
	"PublishOperationReceipt",
}

// OperationIntent is the user-owned intent carried into the semantic IR.
type OperationIntent struct {
	OperationID   string `json:"operation_id"`
	ScenarioID    string `json:"scenario_id"`
	RequestedMode string `json:"requested_mode"`
}

// SourceRevision binds every request and result to one exact source revision.
type SourceRevision struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

// EffectGrant is the closed set of effects that an operation may use.
type EffectGrant struct {
	GrantID  string   `json:"grant_id"`
	ParentID string   `json:"parent_id"`
	Effects  []string `json:"effects"`
}

// EffectRequest is the deterministic request handed to an external executor.
type EffectRequest struct {
	RequestID      string   `json:"request_id"`
	OperationID    string   `json:"operation_id"`
	SourceRevision string   `json:"source_revision"`
	Effects        []string `json:"effects"`
	ReplayIdentity string   `json:"replay_identity"`
	PayloadDigest  string   `json:"payload_digest"`
}

// EffectResult is evidence returned by an executor. It is never treated as a
// semantic approval until it is checked against the request and grant.
type EffectResult struct {
	ResultID       string   `json:"result_id"`
	RequestID      string   `json:"request_id"`
	SourceRevision string   `json:"source_revision"`
	Effects        []string `json:"effects"`
	PayloadDigest  string   `json:"payload_digest"`
	ArtifactDigest string   `json:"artifact_digest"`
}

// SemanticPatch describes the proposal without applying it to the input
// repository.
type SemanticPatch struct {
	Schema           string               `json:"schema"`
	ScenarioID       string               `json:"scenario_id"`
	Changed          bool                 `json:"changed"`
	Operations       []string             `json:"operations"`
	RepositoryWrites int                  `json:"repository_writes"`
	Observation      *SemanticObservation `json:"observation,omitempty"`
}

// ReplayIdentity binds a request to its replay comparison.
type ReplayIdentity struct {
	Identity              string `json:"identity"`
	CurrentRequestDigest  string `json:"current_request_digest"`
	PreviousRequestDigest string `json:"previous_request_digest"`
	Compared              bool   `json:"compared"`
}

// OperationDecision is the semantic outcome. REFUTED always outranks UNKNOWN,
// which always outranks CLOSED.
type OperationDecision struct {
	Decision string                `json:"decision"`
	Reason   string                `json:"reason"`
	Unknown  *EnvelopeUnknownState `json:"unknown"`
}

// EnvelopeUnknownState is intentionally limited to the six required fields.
type EnvelopeUnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

// SemanticOperationIR is the semantic intermediate representation produced
// directly from the .gooo authority and consumed by the artifact generator.
type SemanticOperationIR struct {
	Schema          string               `json:"schema"`
	Intent          OperationIntent      `json:"intent"`
	Source          SourceRevision       `json:"source"`
	Grant           EffectGrant          `json:"grant"`
	Request         *EffectRequest       `json:"request"`
	Result          *EffectResult        `json:"result"`
	Replay          ReplayIdentity       `json:"replay"`
	Decision        OperationDecision    `json:"decision"`
	Activities      []string             `json:"activities"`
	AuthorityDigest string               `json:"authority_digest"`
	ToolchainDigest string               `json:"toolchain_digest"`
	Observation     *SemanticObservation `json:"observation,omitempty"`
}

// EnvelopeMetrics are exact integer observations. The envelope library does
// not run tests or write the input repository; callers may bind measured
// caller-owned observations when those operations actually execute.
type EnvelopeMetrics struct {
	OperationRequests        int   `json:"operation_requests"`
	OperationResults         int   `json:"operation_results"`
	EffectsGranted           int   `json:"effects_granted"`
	EffectsUsed              int   `json:"effects_used"`
	ReplayComparisons        int   `json:"replay_comparisons"`
	ReplayMismatches         int   `json:"replay_mismatches"`
	StaleRejections          int   `json:"stale_rejections"`
	EffectEscalationsRefuted int   `json:"effect_escalations_refuted"`
	InputDescendantDirs      int   `json:"input_descendant_dirs"`
	InputRegularFiles        int   `json:"input_regular_files"`
	InputGoPhysicalLines     int   `json:"input_go_physical_lines"`
	InputGoooPhysicalLines   int   `json:"input_gooo_physical_lines"`
	OutputArtifactFiles      int   `json:"output_artifact_files"`
	PeakRSSKib               int   `json:"peak_rss_kib"`
	WallMS                   int   `json:"wall_ms"`
	RepositoryWrites         int   `json:"repository_writes"`
	LocalTestExecutions      int   `json:"local_test_executions"`
	ObservedOperations       int   `json:"observed_operations"`
	DistinctInputDigests     int   `json:"distinct_input_digests"`
	DuplicateEvaluations     int   `json:"duplicate_evaluations"`
	CandidatesEmitted        int   `json:"candidates_emitted"`
	BeforeOperationCount     int   `json:"before_operation_count"`
	AfterOperationCount      int   `json:"after_operation_count"`
	AllocationCount          int64 `json:"allocation_count"`
	AllocationBytes          int64 `json:"allocation_bytes"`
	BuildMS                  int64 `json:"build_ms"`
	TestMS                   int64 `json:"test_ms"`
	ExecutedTests            int64 `json:"executed_tests"`
	ReusedTests              int64 `json:"reused_tests"`
}

// SemanticOperationReceipt is the compiler-owned receipt written as one of
// the six generated artifacts.
type SemanticOperationReceipt struct {
	Schema              string               `json:"schema"`
	ScenarioID          string               `json:"scenario_id"`
	AuthorityDigest     string               `json:"authority_digest"`
	SourceRevision      SourceRevision       `json:"source_revision"`
	Decision            OperationDecision    `json:"decision"`
	Replay              ReplayIdentity       `json:"replay"`
	Activities          []string             `json:"activities"`
	ManifestDigest      string               `json:"manifest_digest"`
	RequestDigest       string               `json:"request_digest"`
	ResultDigest        string               `json:"result_digest"`
	SemanticPatchDigest string               `json:"semantic_patch_digest"`
	Metrics             EnvelopeMetrics      `json:"metrics"`
	ExternalUserUtility string               `json:"external_user_utility"`
	Observation         *SemanticObservation `json:"observation,omitempty"`
}

// SemanticObservationContract is the exact declaration extracted from the
// .gooo authority. Its fields describe the observation boundary, not a
// proposed optimization.
type SemanticObservationContract struct {
	Activity               string   `json:"activity"`
	Phase                  string   `json:"phase"`
	OperationID            string   `json:"operation_id"`
	CanonicalInputIdentity string   `json:"canonical_input_identity"`
	AllowedEffects         []string `json:"allowed_effects"`
	Pure                   bool     `json:"pure"`
	CandidateRule          string   `json:"candidate_rule"`
}

// SemanticObservationSpan is a stable source span attached to an observed
// compiler operation.
type SemanticObservationSpan struct {
	File        string `json:"file"`
	StartOffset int    `json:"start_offset"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	EndOffset   int    `json:"end_offset"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
}

// SemanticObservationEvent is emitted by a real compiler phase after it has
// evaluated one semantic operation.
type SemanticObservationEvent struct {
	Sequence    int                       `json:"sequence"`
	Phase       string                    `json:"phase"`
	OperationID string                    `json:"operation_id"`
	InputDigest string                    `json:"input_digest"`
	Pure        bool                      `json:"pure"`
	Effects     []string                  `json:"effects"`
	SourceSpans []SemanticObservationSpan `json:"source_spans"`
}

// SemanticObservationCandidate is a stable, evidence-only proposal. The
// reducible count is arithmetic (observed minus one), not a safety or speed
// claim.
type SemanticObservationCandidate struct {
	StableID               string                    `json:"stable_id"`
	Phase                  string                    `json:"phase"`
	OperationID            string                    `json:"operation_id"`
	InputDigest            string                    `json:"input_digest"`
	SourceSpans            []SemanticObservationSpan `json:"source_spans"`
	ObservedCount          int                       `json:"observed_count"`
	ExpectedReducibleCount int                       `json:"expected_reducible_count"`
	SafetyAssessment       string                    `json:"safety_assessment"`
	BenefitAssessment      string                    `json:"benefit_assessment"`
}

// SemanticObservationPairEvidence is the gate for any later exact change.
// This iteration records a replay pair but adopts no change.
type SemanticObservationPairEvidence struct {
	EvidenceAvailable    bool   `json:"evidence_available"`
	ChangeAdopted        bool   `json:"change_adopted"`
	BehaviorEqual        bool   `json:"behavior_equal"`
	DeterminismEqual     bool   `json:"determinism_equal"`
	BeforeOperationCount int    `json:"before_operation_count"`
	AfterOperationCount  int    `json:"after_operation_count"`
	Contradiction        string `json:"contradiction"`
}

// SemanticObservationMetrics keeps exact integer measurements alongside the
// semantic evidence. Build/test values are supplied by the caller-owned CI
// harness when those operations are actually executed.
type SemanticObservationMetrics struct {
	ObservedOperations     int   `json:"observed_operations"`
	DistinctInputDigests   int   `json:"distinct_input_digests"`
	DuplicateEvaluations   int   `json:"duplicate_evaluations"`
	CandidatesEmitted      int   `json:"candidates_emitted"`
	BeforeOperationCount   int   `json:"before_operation_count"`
	AfterOperationCount    int   `json:"after_operation_count"`
	AllocationCount        int64 `json:"allocation_count"`
	AllocationBytes        int64 `json:"allocation_bytes"`
	WallMS                 int64 `json:"wall_ms"`
	PeakRSSKib             int64 `json:"peak_rss_kib"`
	BuildMS                int64 `json:"build_ms"`
	TestMS                 int64 `json:"test_ms"`
	ExecutedTests          int64 `json:"executed_tests"`
	ReusedTests            int64 `json:"reused_tests"`
	InputGoPhysicalLines   int   `json:"input_go_physical_lines"`
	InputGoooPhysicalLines int   `json:"input_gooo_physical_lines"`
	OutputArtifactFiles    int   `json:"output_artifact_files"`
	RepositoryWrites       int   `json:"repository_writes"`
	LocalTestExecutions    int   `json:"local_test_executions"`
}

// SemanticObservation is the compiler-produced self-observation report bound
// into an operation envelope.
type SemanticObservation struct {
	Schema               string                          `json:"schema"`
	Contract             SemanticObservationContract     `json:"contract"`
	ContractDigest       string                          `json:"contract_digest"`
	InputSourceDigest    string                          `json:"input_source_digest"`
	Events               []SemanticObservationEvent      `json:"events"`
	ObservedOperations   int                             `json:"observed_operations"`
	DistinctInputDigests int                             `json:"distinct_input_digests"`
	DuplicateEvaluations int                             `json:"duplicate_evaluations"`
	CandidatesEmitted    int                             `json:"candidates_emitted"`
	Candidates           []SemanticObservationCandidate  `json:"candidates"`
	Decision             string                          `json:"decision"`
	Reason               string                          `json:"reason"`
	Unknown              *EnvelopeUnknownState           `json:"unknown"`
	PairEvidence         SemanticObservationPairEvidence `json:"pair_evidence"`
	Adoption             *SemanticAdoptionEvidence       `json:"adoption,omitempty"`
	Metrics              SemanticObservationMetrics      `json:"metrics"`
}

// SemanticOperationArtifact is an in-memory generated file.
type SemanticOperationArtifact struct {
	Name     string
	Contents []byte
}

// SemanticOperationRun exposes the IR, receipt, and generated artifact bytes.
type SemanticOperationRun struct {
	IR            SemanticOperationIR
	Receipt       SemanticOperationReceipt
	Artifacts     []SemanticOperationArtifact
	ReceiptDigest string
}

// SemanticOperationVerification is returned by the independent verifier.
type SemanticOperationVerification struct {
	ScenarioID            string
	Decision              string
	Reason                string
	ReceiptDigest         string
	Metrics               EnvelopeMetrics
	ObservationDecision   string
	ObservationReason     string
	ObservationCandidates int
}

// SemanticOperationScenarioIDs returns the fixed denominator without allowing
// callers to shrink or reorder the contract.
func SemanticOperationScenarioIDs() []string {
	return []string{"C1", "C2", "U1", "U2", "R1", "R2"}
}

// SemanticOperationActivityNames returns the released eight-activity graph.
func SemanticOperationActivityNames() []string {
	return append([]string(nil), semanticOperationActivities[:]...)
}

// SemanticObservationScenarioIDs returns the fixed observation denominator.
func SemanticObservationScenarioIDs() []string {
	return []string{"NORMAL", "UNKNOWN", "REFUTED", "REPLAY"}
}
