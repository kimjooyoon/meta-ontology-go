package selfimprovementexecutiongrant

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                         = "gooo/self-improvement-execution-grant/v1"
	RequestSchema                  = "gooo/self-improvement-execution-grant-request/v1"
	ResolutionSchema               = "gooo/self-improvement-execution-grant-resolution/v1"
	ReceiptSchema                  = "gooo/self-improvement-execution-grant-receipt/v1"
	PolicySchema                   = "gooo/self-improvement-execution-grant-policy/v1"
	CanonicalCasesSchema           = "gooo/self-improvement-execution-grant-cases/v1"
	VerificationSchema             = "gooo/self-improvement-execution-grant-verification/v1"
	ContractID                     = "gooo://self-improvement/execution-grant/v1"
	PolicyName                     = "SeparateExecutionGrant"
	PolicyPath                     = "examples/self-improvement-execution-grant/grant.gooo"
	GrantRequestReason             = "EXECUTION_GRANT_REQUESTED"
	GrantTarget                    = "self-improvement-execution"
	GrantMode                      = "separate-one-use-grant"
	MaxExecutions                  = 1
	PerformanceUnknown             = "UNKNOWN"
	GrantDecisionSchema            = "gooo/self-improvement-execution-grant-decision/v1"
	DecisionAllow                  = "ALLOW"
	DecisionDeny                   = "DENY"
	DecisionSourceWorkflowDispatch = "workflow_dispatch"
	DecisionSourceCanonical        = "canonical-fixture"
	ActorEvidenceLabel             = "GITHUB_EVENT_ACTOR_EVIDENCE"
	CanonicalEvidenceLabel         = "CANONICAL_FIXTURE_METADATA"
	ConsumptionObligation          = "NEXT_EXECUTOR_MUST_VERIFY_AND_CONSUME_ONCE"
	ConsumptionPending             = "UNCONSUMED_NOT_EXECUTED"
)

type Decision string

const (
	DecisionClosed  Decision = "CLOSED"
	DecisionUnknown Decision = "UNKNOWN"
	DecisionRefuted Decision = "REFUTED"
)

type Resolution string

const (
	ResolutionExact             Resolution = "EXACT"
	ResolutionLower             Resolution = "LOWER_RESOLUTION"
	ResolutionDenied            Resolution = "DENIED"
	ResolutionGrantedUnconsumed Resolution = "GRANTED_UNCONSUMED"
)

type V24Binding struct {
	RequestSchema           string `json:"request_schema"`
	RequestDigest           string `json:"request_digest"`
	ResolutionSchema        string `json:"resolution_schema"`
	ResolutionDigest        string `json:"resolution_digest"`
	CandidateStableID       string `json:"candidate_stable_id"`
	CandidateDigest         string `json:"candidate_digest"`
	SubjectSHA              string `json:"subject_sha"`
	ObservationDigest       string `json:"observation_digest"`
	ContractDigest          string `json:"contract_digest"`
	AuthorizationDecision   string `json:"authorization_decision"`
	AuthorizationResolution string `json:"authorization_resolution"`
	AuthorizationOutcome    string `json:"authorization_outcome"`
	RequestValid            bool   `json:"request_valid"`
	ResolutionValid         bool   `json:"resolution_valid"`
}

type V25Binding struct {
	Schema                        string `json:"schema"`
	ContractID                    string `json:"contract_id"`
	ContractDigest                string `json:"contract_digest"`
	Decision                      string `json:"decision"`
	Resolution                    string `json:"resolution"`
	CandidateStableID             string `json:"candidate_stable_id"`
	CandidateDigest               string `json:"candidate_digest"`
	SubjectSHA                    string `json:"subject_sha"`
	ObservationDigest             string `json:"observation_digest"`
	CandidateInputDigest          string `json:"candidate_input_digest"`
	OperationID                   string `json:"operation_id"`
	BoundedTarget                 string `json:"bounded_target"`
	EvaluatorRegistryDigest       string `json:"evaluator_registry_digest"`
	ToolchainTestContractIdentity string `json:"toolchain_test_contract_identity"`
	MaxExecutions                 int    `json:"max_executions"`
	RepositoryWritesAllowed       bool   `json:"repository_writes_allowed"`
	ExecutionAuthorized           bool   `json:"execution_authorized"`
	ExecutionGrantRequired        bool   `json:"execution_grant_required"`
	Valid                         bool   `json:"valid"`
}

type SourceArtifact struct {
	Repository             string `json:"repository"`
	WorkflowRunID          int64  `json:"workflow_run_id"`
	WorkflowRunAttempt     int    `json:"workflow_run_attempt"`
	ArtifactID             int64  `json:"artifact_id"`
	ArtifactDigest         string `json:"artifact_digest"`
	ObservedArtifactDigest string `json:"observed_artifact_digest,omitempty"`
	ArtifactExpired        bool   `json:"artifact_expired"`
	ArtifactExpiryKnown    bool   `json:"artifact_expiry_known"`
}

type ActorEvidence struct {
	Repository         string `json:"repository"`
	Actor              string `json:"actor"`
	WorkflowRunID      int64  `json:"workflow_run_id"`
	WorkflowRunAttempt int    `json:"workflow_run_attempt"`
	Event              string `json:"event"`
	EvidenceLabel      string `json:"evidence_label"`
}

type GrantRequest struct {
	Schema     string         `json:"schema"`
	Lifecycle  string         `json:"lifecycle"`
	ContractID string         `json:"contract_id"`
	V24        V24Binding     `json:"v24_authorization"`
	V25        V25Binding     `json:"v25_pre_execution_contract"`
	Source     SourceArtifact `json:"source_artifact"`
	Target     string         `json:"target"`
	Mode       string         `json:"mode"`
	Decision   string         `json:"decision"`
	Resolution string         `json:"resolution"`
	Reason     string         `json:"reason"`
	Digest     string         `json:"digest"`
}

type GrantDecisionInput struct {
	Schema         string         `json:"schema"`
	Decision       string         `json:"decision"`
	RequestDigest  string         `json:"request_digest"`
	V24            V24Binding     `json:"v24_authorization"`
	V25            V25Binding     `json:"v25_pre_execution_contract"`
	Source         SourceArtifact `json:"source_artifact"`
	ActorEvidence  ActorEvidence  `json:"actor_evidence"`
	DecisionSource string         `json:"decision_source"`
	DecisionDigest string         `json:"decision_digest,omitempty"`
}

type GrantInput struct {
	Request        GrantRequest         `json:"request"`
	DecisionInputs []GrantDecisionInput `json:"decision_inputs,omitempty"`
	Live           bool                 `json:"-"`
}

type UnknownState struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

type GrantReceipt struct {
	Schema                 string        `json:"schema"`
	GrantID                string        `json:"grant_id"`
	RequestDigest          string        `json:"request_digest"`
	Decision               string        `json:"decision"`
	DecisionSource         string        `json:"decision_source"`
	ActorEvidence          ActorEvidence `json:"actor_evidence"`
	GrantAllowsExecution   bool          `json:"grant_allows_execution"`
	RemainingUses          int           `json:"remaining_uses"`
	ConsumedUses           int           `json:"consumed_uses"`
	ExecutionCount         int           `json:"execution_count"`
	ConsumptionStatus      string        `json:"consumption_status"`
	ConsumptionObligation  string        `json:"consumption_obligation"`
	OneUseEnforcementState string        `json:"one_use_enforcement_state"`
	Digest                 string        `json:"digest"`
}

type GrantMetrics struct {
	StructuralSeparateGrantEdgesBefore int    `json:"structural_separate_grant_edges_before"`
	StructuralSeparateGrantEdgesAfter  int    `json:"structural_separate_grant_edges_after"`
	LiveGrantRequests                  int    `json:"live_grant_requests"`
	LiveGrants                         int    `json:"live_grants"`
	LiveExecutionCount                 int    `json:"live_execution_count"`
	CanonicalGrantedCases              int    `json:"canonical_granted_cases"`
	CanonicalExecutionCount            int    `json:"canonical_execution_count"`
	SixFieldUnknowns                   int    `json:"six_field_unknowns"`
	RefutedContradictions              int    `json:"refuted_contradictions"`
	GrantRemainingUses                 int    `json:"grant_remaining_uses"`
	GrantConsumedUses                  int    `json:"grant_consumed_uses"`
	RepositoryWrites                   int    `json:"repository_writes"`
	LocalTestExecutions                int    `json:"local_test_executions"`
	FallbackAccepted                   int    `json:"fallback_accepted"`
	IndependentReplayComparisons       int    `json:"independent_replay_comparisons"`
	ArtifactFiles                      int    `json:"artifact_files"`
	ArtifactTypes                      int    `json:"artifact_types"`
	GoPhysicalLines                    int    `json:"go_physical_lines"`
	GoooPhysicalLines                  int    `json:"gooo_physical_lines"`
	PerformanceImprovement             string `json:"performance_improvement"`
}

type PolicyEvidence struct {
	Schema           string `json:"schema"`
	PolicyID         string `json:"policy_id"`
	SourceDigest     string `json:"source_digest"`
	CanonicalDigest  string `json:"canonical_digest"`
	SemanticIRDigest string `json:"semantic_ir_digest"`
	StateCount       int    `json:"state_count"`
	TransitionCount  int    `json:"transition_count"`
	CaseCount        int    `json:"case_count"`
	ClosedCases      int    `json:"closed_cases"`
	UnknownCases     int    `json:"unknown_cases"`
	RefutedCases     int    `json:"refuted_cases"`
}

type GrantResolution struct {
	Schema                string               `json:"schema"`
	RequestDigest         string               `json:"request_digest"`
	Decision              Decision             `json:"decision"`
	Resolution            Resolution           `json:"resolution"`
	Reason                string               `json:"reason"`
	Unknown               *UnknownState        `json:"unknown,omitempty"`
	MissingFields         []string             `json:"missing_fields,omitempty"`
	ContradictoryFields   []string             `json:"contradictory_fields,omitempty"`
	DecisionInputs        []GrantDecisionInput `json:"decision_inputs,omitempty"`
	GrantAllowsExecution  bool                 `json:"grant_allows_execution"`
	RemainingUses         int                  `json:"remaining_uses"`
	ConsumedUses          int                  `json:"consumed_uses"`
	ExecutionCount        int                  `json:"execution_count"`
	RepositoryWrites      int                  `json:"repository_writes"`
	LocalTestExecutions   int                  `json:"local_test_executions"`
	FallbackAccepted      int                  `json:"fallback_accepted"`
	ConsumptionObligation string               `json:"consumption_obligation"`
	OneUseEnforced        bool                 `json:"one_use_enforced"`
	Receipt               *GrantReceipt        `json:"receipt,omitempty"`
	Metrics               GrantMetrics         `json:"metrics"`
	Digest                string               `json:"digest"`
}

type Verification struct {
	Schema                       string     `json:"schema"`
	RequestDigest                string     `json:"request_digest"`
	ResolutionDigest             string     `json:"resolution_digest"`
	IndependentDecision          Decision   `json:"independent_decision"`
	IndependentResolution        Resolution `json:"independent_resolution"`
	IndependentReason            string     `json:"independent_reason"`
	Verified                     bool       `json:"verified"`
	IndependentReplayComparisons int        `json:"independent_replay_comparisons"`
	LiveGrants                   int        `json:"live_grants"`
	ExecutionCount               int        `json:"execution_count"`
	GrantConsumedUses            int        `json:"grant_consumed_uses"`
	RepositoryWrites             int        `json:"repository_writes"`
	LocalTestExecutions          int        `json:"local_test_executions"`
	Digest                       string     `json:"digest"`
}

type LiveReport struct {
	Schema         string          `json:"schema"`
	Policy         PolicyEvidence  `json:"policy"`
	Request        GrantRequest    `json:"request"`
	GrantDecision  string          `json:"grant_decision"`
	DecisionSource string          `json:"decision_source"`
	Resolution     GrantResolution `json:"resolution"`
	Verification   Verification    `json:"verification"`
	Metrics        GrantMetrics    `json:"metrics"`
	Digest         string          `json:"digest"`
}

type CanonicalCase struct {
	ID                   string        `json:"id"`
	ExpectedDecision     Decision      `json:"expected_decision"`
	ExpectedResolution   Resolution    `json:"expected_resolution"`
	ExpectedReason       string        `json:"expected_reason"`
	ActualDecision       Decision      `json:"actual_decision"`
	ActualResolution     Resolution    `json:"actual_resolution"`
	ActualReason         string        `json:"actual_reason"`
	Unknown              *UnknownState `json:"unknown,omitempty"`
	GrantAllowsExecution bool          `json:"grant_allows_execution"`
	RemainingUses        int           `json:"remaining_uses"`
	ConsumedUses         int           `json:"consumed_uses"`
	ExecutionCount       int           `json:"execution_count"`
	Pass                 bool          `json:"pass"`
}

type CanonicalCaseReport struct {
	Schema                             string          `json:"schema"`
	Policy                             PolicyEvidence  `json:"policy"`
	RequiredFields                     []string        `json:"required_fields"`
	RequestDigest                      string          `json:"request_digest"`
	CaseDenominator                    int             `json:"case_denominator"`
	StructuralSeparateGrantEdgesBefore int             `json:"structural_separate_grant_edges_before"`
	StructuralSeparateGrantEdgesAfter  int             `json:"structural_separate_grant_edges_after"`
	ClosedCases                        int             `json:"closed_cases"`
	UnknownCases                       int             `json:"unknown_cases"`
	RefutedCases                       int             `json:"refuted_cases"`
	Counts                             map[string]int  `json:"counts"`
	Cases                              []CanonicalCase `json:"cases"`
	ReplayEqual                        bool            `json:"replay_equal"`
	LiveGrantRequests                  int             `json:"live_grant_requests"`
	LiveGrants                         int             `json:"live_grants"`
	LiveExecutionCount                 int             `json:"live_execution_count"`
	CanonicalGrantedCases              int             `json:"canonical_granted_cases"`
	CanonicalExecutionCount            int             `json:"canonical_execution_count"`
	SixFieldUnknowns                   int             `json:"six_field_unknowns"`
	RefutedContradictions              int             `json:"refuted_contradictions"`
	GrantRemainingUses                 int             `json:"grant_remaining_uses"`
	GrantConsumedUses                  int             `json:"grant_consumed_uses"`
	RepositoryWrites                   int             `json:"repository_writes"`
	LocalTestExecutions                int             `json:"local_test_executions"`
	FallbackAccepted                   int             `json:"fallback_accepted"`
	IndependentReplayComparisons       int             `json:"independent_replay_comparisons"`
	ArtifactFiles                      int             `json:"artifact_files"`
	ArtifactTypes                      int             `json:"artifact_types"`
	GoPhysicalLines                    int             `json:"go_physical_lines"`
	GoooPhysicalLines                  int             `json:"gooo_physical_lines"`
	PerformanceImprovement             string          `json:"performance_improvement"`
	Decision                           Decision        `json:"decision"`
	Resolution                         Resolution      `json:"resolution"`
	Reason                             string          `json:"reason"`
	Digest                             string          `json:"digest"`
}

type PolicyProgram struct {
	Evidence  PolicyEvidence  `json:"evidence"`
	Policy    semantic.Policy `json:"-"`
	Inventory SourceInventory `json:"inventory"`
}

type SourceInventory struct {
	GoFiles           int  `json:"go_files"`
	GoooFiles         int  `json:"gooo_files"`
	GoPhysicalLines   int  `json:"go_physical_lines"`
	GoooPhysicalLines int  `json:"gooo_physical_lines"`
	Observed          bool `json:"observed"`
}

func digestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func clearRequestDigest(value GrantRequest) GrantRequest {
	value.Digest = ""
	return value
}

func clearDecisionDigest(value GrantDecisionInput) GrantDecisionInput {
	value.DecisionDigest = ""
	return value
}

func clearResolutionDigest(value GrantResolution) GrantResolution {
	value.Digest = ""
	return value
}

func clearReceiptDigest(value GrantReceipt) GrantReceipt {
	value.Digest = ""
	return value
}

func clearCanonicalDigest(value CanonicalCaseReport) CanonicalCaseReport {
	value.Digest = ""
	return value
}

func requestDigest(value GrantRequest) string        { return digestJSON(clearRequestDigest(value)) }
func decisionDigest(value GrantDecisionInput) string { return digestJSON(clearDecisionDigest(value)) }
func resolutionDigest(value GrantResolution) string  { return digestJSON(clearResolutionDigest(value)) }
func receiptDigest(value GrantReceipt) string        { return digestJSON(clearReceiptDigest(value)) }
func canonicalDigest(value CanonicalCaseReport) string {
	return digestJSON(clearCanonicalDigest(value))
}
func verificationDigest(value Verification) string { value.Digest = ""; return digestJSON(value) }

func ProjectV24(request candidate.AuthorizationRequest, resolution candidate.AuthorizationResolution) V24Binding {
	return V24Binding{
		RequestSchema: request.Schema, RequestDigest: request.Digest,
		ResolutionSchema: resolution.Schema, ResolutionDigest: resolution.Digest,
		CandidateStableID: request.Candidate.CandidateID, CandidateDigest: request.Candidate.CandidateDigest,
		SubjectSHA: request.Candidate.SubjectSHA, ObservationDigest: request.Candidate.SourceObservationDigest,
		ContractDigest:        request.Candidate.ContractCanonicalDigest,
		AuthorizationDecision: resolution.Decision, AuthorizationResolution: resolution.Resolution,
		AuthorizationOutcome: resolution.Outcome,
		RequestValid:         candidate.ValidateAuthorizationRequest(request) == nil,
		ResolutionValid:      candidate.ValidateAuthorizationResolution(resolution) == nil,
	}
}

func ProjectV25(resolution v25.ContractResolution) V25Binding {
	return V25Binding{
		Schema: resolution.Schema, ContractID: resolution.ContractID, ContractDigest: resolution.Digest,
		Decision: string(resolution.Decision), Resolution: string(resolution.Resolution),
		CandidateStableID: resolution.CandidateStableID, CandidateDigest: resolution.CandidateDigest,
		SubjectSHA: resolution.SubjectSHA, ObservationDigest: resolution.ObservationDigest,
		CandidateInputDigest: resolution.CandidateInputDigest, OperationID: string(resolution.OperationID),
		BoundedTarget:                 string(resolution.BoundedTarget),
		EvaluatorRegistryDigest:       resolution.EvaluatorRegistryDigest,
		ToolchainTestContractIdentity: resolution.ToolchainTestContractID,
		MaxExecutions:                 resolution.MaxExecutions, RepositoryWritesAllowed: resolution.RepositoryWritesAllowed,
		ExecutionAuthorized: resolution.ExecutionAuthorized, ExecutionGrantRequired: resolution.ExecutionGrantRequired,
		Valid: v25.ValidateResolution(resolution) == nil,
	}
}

func BuildRequest(program PolicyProgram, v24Binding V24Binding, v25Binding V25Binding, source SourceArtifact) GrantRequest {
	request := GrantRequest{Schema: RequestSchema, Lifecycle: "REQUESTED", ContractID: ContractID,
		V24: v24Binding, V25: v25Binding, Source: source, Target: GrantTarget, Mode: GrantMode,
		Decision: "REQUESTED", Resolution: "UNRESOLVED", Reason: GrantRequestReason}
	request.Digest = requestDigest(request)
	return request
}

func BuildDecisionInput(request GrantRequest, decision string, source string, actor ActorEvidence) GrantDecisionInput {
	input := GrantDecisionInput{Schema: GrantDecisionSchema, Decision: decision, RequestDigest: request.Digest,
		V24: request.V24, V25: request.V25, Source: request.Source, DecisionSource: source, ActorEvidence: actor}
	input.DecisionDigest = decisionDigest(input)
	return input
}

func RequiredBindingNames() []string {
	return []string{"v24_request_digest", "v24_resolution_digest", "candidate_stable_id", "candidate_digest", "subject_sha", "observation_digest", "v24_contract_digest", "v25_contract_digest", "operation_id", "evaluator_registry_digest", "toolchain_test_contract_identity", "max_executions", "repository_writes_allowed", "source_workflow_run_id", "source_workflow_run_attempt", "source_artifact_id", "source_artifact_digest", "source_artifact_expiry"}
}

func CompilePolicy(repository fs.FS, path string) (PolicyProgram, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("read execution grant policy: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return PolicyProgram{}, errors.New("execution grant policy has syntax errors")
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("format execution grant policy: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("lower execution grant policy: %w", err)
	}
	if len(ir.Policies) != 1 {
		return PolicyProgram{}, errors.New("execution grant policy must lower to one semantic policy")
	}
	policy := ir.Policies[0]
	if err := validatePolicy(policy); err != nil {
		return PolicyProgram{}, err
	}
	inventory, err := ObserveInventory(repository)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("observe source inventory: %w", err)
	}
	return PolicyProgram{Policy: policy, Inventory: inventory, Evidence: PolicyEvidence{
		Schema: PolicySchema, PolicyID: policy.ID.String(), SourceDigest: digestBytes(raw),
		CanonicalDigest: digestBytes([]byte(canonical)), SemanticIRDigest: digestBytes([]byte(policy.SemanticCanonical())),
		StateCount: len(policy.States), TransitionCount: len(policy.Transitions), CaseCount: len(policy.Cases),
		ClosedCases: 3, UnknownCases: 3, RefutedCases: 3,
	}}, nil
}

func ObserveInventory(repository fs.FS) (SourceInventory, error) {
	inventory := SourceInventory{Observed: true}
	err := fs.WalkDir(repository, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			return err
		}
		lines := bytes.Count(raw, []byte{'\n'})
		if len(raw) > 0 && raw[len(raw)-1] != '\n' {
			lines++
		}
		switch {
		case strings.HasSuffix(path, ".go"):
			inventory.GoFiles++
			inventory.GoPhysicalLines += lines
		case strings.HasSuffix(path, ".gooo"):
			inventory.GoooFiles++
			inventory.GoooPhysicalLines += lines
		}
		return nil
	})
	return inventory, err
}

func validatePolicy(policy semantic.Policy) error {
	if policy.Name != PolicyName || policy.ID.String() != ContractID || len(policy.States) != 12 || len(policy.Transitions) != 9 || len(policy.Cases) != 9 {
		return errors.New("execution grant policy denominator or identity drifted")
	}
	counts := map[string]int{}
	for _, current := range policy.Cases {
		counts[current.Resolution.Decision]++
		if current.Resolution.Decision == string(DecisionUnknown) && (current.Resolution.Stage == "" || current.Resolution.Step == 0 || current.Resolution.Reason == "" || current.Resolution.UnknownClass == "" || current.Resolution.NextOperation == "" || len(current.Resolution.BlockedBy) == 0) {
			return fmt.Errorf("unknown policy case %q does not declare all six fields", current.Name)
		}
	}
	if counts[string(DecisionClosed)] != 3 || counts[string(DecisionUnknown)] != 3 || counts[string(DecisionRefuted)] != 3 {
		return errors.New("execution grant policy case partition drifted")
	}
	return nil
}

func validArtifact(source SourceArtifact) bool {
	return source.Repository != "" && source.WorkflowRunID > 0 && source.WorkflowRunAttempt > 0 && source.ArtifactID > 0 && validDigest(source.ArtifactDigest) && (source.ObservedArtifactDigest == "" || source.ArtifactDigest == source.ObservedArtifactDigest) && source.ArtifactExpiryKnown && !source.ArtifactExpired
}
