package generation

// SemanticOperationEnvelopeSchema identifies the small, replayable operation
// boundary owned by the semantic layer.
const SemanticOperationEnvelopeSchema = "gooo/semantic-operation-envelope/v1"

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
	Schema           string   `json:"schema"`
	ScenarioID       string   `json:"scenario_id"`
	Changed          bool     `json:"changed"`
	Operations       []string `json:"operations"`
	RepositoryWrites int      `json:"repository_writes"`
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
	Schema          string            `json:"schema"`
	Intent          OperationIntent   `json:"intent"`
	Source          SourceRevision    `json:"source"`
	Grant           EffectGrant       `json:"grant"`
	Request         *EffectRequest    `json:"request"`
	Result          *EffectResult     `json:"result"`
	Replay          ReplayIdentity    `json:"replay"`
	Decision        OperationDecision `json:"decision"`
	Activities      []string          `json:"activities"`
	AuthorityDigest string            `json:"authority_digest"`
	ToolchainDigest string            `json:"toolchain_digest"`
}

// EnvelopeMetrics are exact integer observations. The library does not run
// tests or write the input repository, so both corresponding fields remain 0.
type EnvelopeMetrics struct {
	OperationRequests        int `json:"operation_requests"`
	OperationResults         int `json:"operation_results"`
	EffectsGranted           int `json:"effects_granted"`
	EffectsUsed              int `json:"effects_used"`
	ReplayComparisons        int `json:"replay_comparisons"`
	ReplayMismatches         int `json:"replay_mismatches"`
	StaleRejections          int `json:"stale_rejections"`
	EffectEscalationsRefuted int `json:"effect_escalations_refuted"`
	InputDescendantDirs      int `json:"input_descendant_dirs"`
	InputRegularFiles        int `json:"input_regular_files"`
	InputGoPhysicalLines     int `json:"input_go_physical_lines"`
	InputGoooPhysicalLines   int `json:"input_gooo_physical_lines"`
	OutputArtifactFiles      int `json:"output_artifact_files"`
	PeakRSSKib               int `json:"peak_rss_kib"`
	WallMS                   int `json:"wall_ms"`
	RepositoryWrites         int `json:"repository_writes"`
	LocalTestExecutions      int `json:"local_test_executions"`
}

// SemanticOperationReceipt is the compiler-owned receipt written as one of
// the six generated artifacts.
type SemanticOperationReceipt struct {
	Schema              string            `json:"schema"`
	ScenarioID          string            `json:"scenario_id"`
	AuthorityDigest     string            `json:"authority_digest"`
	SourceRevision      SourceRevision    `json:"source_revision"`
	Decision            OperationDecision `json:"decision"`
	Replay              ReplayIdentity    `json:"replay"`
	Activities          []string          `json:"activities"`
	ManifestDigest      string            `json:"manifest_digest"`
	RequestDigest       string            `json:"request_digest"`
	ResultDigest        string            `json:"result_digest"`
	SemanticPatchDigest string            `json:"semantic_patch_digest"`
	Metrics             EnvelopeMetrics   `json:"metrics"`
	ExternalUserUtility string            `json:"external_user_utility"`
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
	ScenarioID    string
	Decision      string
	Reason        string
	ReceiptDigest string
	Metrics       EnvelopeMetrics
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
