package selfimprovementcontinuation

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
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                 = "gooo/self-improvement-ci-continuation/v1"
	RequestSchema          = "gooo/self-improvement-ci-continuation-request/v1"
	ResolutionSchema       = "gooo/self-improvement-ci-continuation-resolution/v1"
	VerificationSchema     = "gooo/self-improvement-ci-continuation-verification/v1"
	CanonicalCasesSchema   = "gooo/self-improvement-ci-continuation-cases/v1"
	PolicySchema           = "gooo/self-improvement-ci-continuation-policy/v1"
	ContractID             = "gooo://self-improvement/ci-continuation/v1"
	PolicyName             = "CIContinuationContract"
	PolicyPath             = "examples/self-improvement-ci-continuation/continuation.gooo"
	SourceWorkflowName     = "Self-improvement candidate authorization bridge"
	SourceWorkflowPath     = ".github/workflows/self-improvement-candidate-authorization.yml"
	SourceRepository       = "kimjooyoon/meta-ontology-go"
	SourceEvent            = "workflow_run"
	SourceRef              = "refs/heads/dev"
	DispatchRef            = "dev"
	V25WorkflowName        = "Self-improvement candidate execution contract"
	V25WorkflowPath        = ".github/workflows/self-improvement-execution-contract.yml"
	V26WorkflowName        = "Self-improvement separate execution grant"
	V26WorkflowPath        = ".github/workflows/self-improvement-execution-grant.yml"
	DispatchMode           = "workflow_dispatch"
	PerformanceUnknown     = "UNKNOWN"
	ArtifactFiles          = 5
	ArtifactTypes          = 5
	CounterexampleRunID    = int64(33926584593)
	DepthEdgesBefore       = 2
	DepthEdgesAfter        = 0
	IdentityBindingsBefore = 0
	IdentityBindingsAfter  = 2
)

type Decision string

const (
	DecisionClosed  Decision = "CLOSED"
	DecisionUnknown Decision = "UNKNOWN"
	DecisionRefuted Decision = "REFUTED"
)

type Resolution string

const (
	ResolutionExact Resolution = "EXACT"
	ResolutionLower Resolution = "LOWER_RESOLUTION"
)

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

// ContinuationInput is the typed CI_CONTINUATION_REQUEST payload. It is a
// scheduling proof only: no field in this type grants execution or mutation.
type ContinuationInput struct {
	SourceWorkflowName           string `json:"source_workflow_name"`
	SourceWorkflowPath           string `json:"source_workflow_path"`
	SourceRepository             string `json:"source_repository"`
	SourceEvent                  string `json:"source_event"`
	SourceRef                    string `json:"source_ref"`
	SourceHeadSHA                string `json:"source_head_sha"`
	SourceRunID                  int64  `json:"source_run_id"`
	SourceRunAttempt             int    `json:"source_run_attempt"`
	SourceArtifactName           string `json:"source_artifact_name"`
	SourceArtifactID             int64  `json:"source_artifact_id"`
	SourceArtifactArchiveDigest  string `json:"source_artifact_archive_digest"`
	SourceArtifactObservedDigest string `json:"source_artifact_observed_digest"`
	SourceReceiptDigest          string `json:"source_receipt_digest"`
	TargetWorkflowName           string `json:"target_workflow_name"`
	TargetWorkflowPath           string `json:"target_workflow_path"`
	DispatchRef                  string `json:"dispatch_ref"`
	DispatchMode                 string `json:"dispatch_mode"`
	Replay                       bool   `json:"replay"`
	DuplicateDispatch            bool   `json:"duplicate_dispatch"`
	DuplicateConflict            bool   `json:"duplicate_conflict"`
	ManualDispatches             int    `json:"manual_dispatches"`
	UnauthorizedDispatches       int    `json:"unauthorized_dispatches"`
	ExecutionAuthorized          bool   `json:"execution_authorized"`
	ExecutionGrants              int    `json:"execution_grants"`
	LiveGrantDecision            int    `json:"live_grant_decision"`
	LiveExecutionCount           int    `json:"live_execution_count"`
	GrantConsumedUses            int    `json:"grant_consumed_uses"`
	RepositoryWrites             int    `json:"repository_writes"`
	LocalTestExecutions          int    `json:"local_test_executions"`
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

type Metrics struct {
	WorkflowRunContinuationEdgesBefore int    `json:"workflow_run_continuation_edges_before"`
	WorkflowRunContinuationEdgesAfter  int    `json:"workflow_run_continuation_edges_after"`
	ExactIdentityBindingsBefore        int    `json:"exact_identity_bindings_before"`
	ExactIdentityBindingsAfter         int    `json:"exact_identity_bindings_after"`
	ManualDispatchesBefore             int    `json:"manual_dispatches_before"`
	ManualDispatchesAfter              int    `json:"manual_dispatches_after"`
	LiveGrantDecision                  int    `json:"live_grant_decision"`
	LiveGrants                         int    `json:"live_grants"`
	LiveExecutionCount                 int    `json:"live_execution_count"`
	GrantConsumedUses                  int    `json:"grant_consumed_uses"`
	RepositoryWrites                   int    `json:"repository_writes"`
	LocalTestExecutions                int    `json:"local_test_executions"`
	CanonicalExecutionCount            int    `json:"canonical_execution_count"`
	CanonicalCases                     int    `json:"canonical_cases"`
	ClosedCases                        int    `json:"closed_cases"`
	UnknownCases                       int    `json:"unknown_cases"`
	RefutedCases                       int    `json:"refuted_cases"`
	SixFieldUnknowns                   int    `json:"six_field_unknowns"`
	UnauthorizedDispatches             int    `json:"unauthorized_dispatches"`
	FallbackAccepted                   int    `json:"fallback_accepted"`
	IndependentReplayComparisons       int    `json:"independent_replay_comparisons"`
	ArtifactFiles                      int    `json:"artifact_files"`
	ArtifactTypes                      int    `json:"artifact_types"`
	GoPhysicalLines                    int    `json:"go_physical_lines"`
	GoooPhysicalLines                  int    `json:"gooo_physical_lines"`
	PerformanceImprovement             string `json:"performance_improvement"`
	CounterexampleRunID                int64  `json:"counterexample_run_id"`
}

type ContinuationRequest struct {
	Schema              string            `json:"schema"`
	Lifecycle           string            `json:"lifecycle"`
	ContractID          string            `json:"contract_id"`
	Input               ContinuationInput `json:"continuation_input"`
	Decision            Decision          `json:"decision"`
	Resolution          Resolution        `json:"resolution"`
	Reason              string            `json:"reason"`
	ExecutionAuthorized bool              `json:"execution_authorized"`
	RepositoryWrites    int               `json:"repository_writes"`
	LocalTestExecutions int               `json:"local_test_executions"`
	Digest              string            `json:"digest"`
}

type ContinuationResolution struct {
	Schema              string            `json:"schema"`
	RequestDigest       string            `json:"request_digest"`
	Input               ContinuationInput `json:"continuation_input"`
	Decision            Decision          `json:"decision"`
	Resolution          Resolution        `json:"resolution"`
	Reason              string            `json:"reason"`
	Unknown             *UnknownState     `json:"unknown,omitempty"`
	MissingFields       []string          `json:"missing_fields,omitempty"`
	ContradictoryFields []string          `json:"contradictory_fields,omitempty"`
	ExecutionAuthorized bool              `json:"execution_authorized"`
	ExecutionGrants     int               `json:"execution_grants"`
	LiveGrantDecision   int               `json:"live_grant_decision"`
	LiveExecutionCount  int               `json:"live_execution_count"`
	GrantConsumedUses   int               `json:"grant_consumed_uses"`
	RepositoryWrites    int               `json:"repository_writes"`
	LocalTestExecutions int               `json:"local_test_executions"`
	Digest              string            `json:"digest"`
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
	ExecutionAuthorized          bool       `json:"execution_authorized"`
	LiveGrantDecision            int        `json:"live_grant_decision"`
	LiveExecutionCount           int        `json:"live_execution_count"`
	GrantConsumedUses            int        `json:"grant_consumed_uses"`
	RepositoryWrites             int        `json:"repository_writes"`
	LocalTestExecutions          int        `json:"local_test_executions"`
	Digest                       string     `json:"digest"`
}

type Report struct {
	Schema       string                 `json:"schema"`
	Policy       PolicyEvidence         `json:"policy"`
	Request      ContinuationRequest    `json:"request"`
	Resolution   ContinuationResolution `json:"resolution"`
	Verification Verification           `json:"verification"`
	Metrics      Metrics                `json:"metrics"`
	Digest       string                 `json:"digest"`
}

type CanonicalCase struct {
	ID               string        `json:"id"`
	ExpectedDecision Decision      `json:"expected_decision"`
	ExpectedReason   string        `json:"expected_reason"`
	ActualDecision   Decision      `json:"actual_decision"`
	ActualReason     string        `json:"actual_reason"`
	Unknown          *UnknownState `json:"unknown,omitempty"`
	Pass             bool          `json:"pass"`
}

type CanonicalCaseReport struct {
	Schema          string          `json:"schema"`
	Policy          PolicyEvidence  `json:"policy"`
	CaseDenominator int             `json:"case_denominator"`
	ClosedCases     int             `json:"closed_cases"`
	UnknownCases    int             `json:"unknown_cases"`
	RefutedCases    int             `json:"refuted_cases"`
	Counts          map[string]int  `json:"counts"`
	Cases           []CanonicalCase `json:"cases"`
	ReplayEqual     bool            `json:"replay_equal"`
	Metrics         Metrics         `json:"metrics"`
	Decision        Decision        `json:"decision"`
	Resolution      Resolution      `json:"resolution"`
	Reason          string          `json:"reason"`
	Digest          string          `json:"digest"`
}

type PolicyProgram struct {
	Evidence          PolicyEvidence  `json:"evidence"`
	Policy            semantic.Policy `json:"-"`
	GoPhysicalLines   int             `json:"go_physical_lines"`
	GoooPhysicalLines int             `json:"gooo_physical_lines"`
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

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func clearRequestDigest(value ContinuationRequest) ContinuationRequest {
	value.Digest = ""
	return value
}
func requestDigest(value ContinuationRequest) string { return digestJSON(clearRequestDigest(value)) }
func clearResolutionDigest(value ContinuationResolution) ContinuationResolution {
	value.Digest = ""
	return value
}
func resolutionDigest(value ContinuationResolution) string {
	return digestJSON(clearResolutionDigest(value))
}
func clearVerificationDigest(value Verification) Verification { value.Digest = ""; return value }
func verificationDigest(value Verification) string            { return digestJSON(clearVerificationDigest(value)) }
func clearReportDigest(value Report) Report                   { value.Digest = ""; return value }
func reportDigest(value Report) string                        { return digestJSON(clearReportDigest(value)) }
func clearCanonicalDigest(value CanonicalCaseReport) CanonicalCaseReport {
	value.Digest = ""
	return value
}
func canonicalDigest(value CanonicalCaseReport) string {
	return digestJSON(clearCanonicalDigest(value))
}

func ObserveInventory(repository fs.FS) (int, int, error) {
	goLines, goooLines := 0, 0
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
		if strings.HasSuffix(path, ".go") {
			goLines += lines
		}
		if strings.HasSuffix(path, ".gooo") {
			goooLines += lines
		}
		return nil
	})
	return goLines, goooLines, err
}

func CompilePolicy(repository fs.FS, path string) (PolicyProgram, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("read continuation contract: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return PolicyProgram{}, errors.New("continuation contract has syntax errors")
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("format continuation contract: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("lower continuation contract: %w", err)
	}
	if len(ir.Policies) != 1 {
		return PolicyProgram{}, errors.New("continuation contract must lower to one policy")
	}
	policy := ir.Policies[0]
	if err := validatePolicy(policy); err != nil {
		return PolicyProgram{}, err
	}
	goLines, goooLines, err := ObserveInventory(repository)
	if err != nil {
		return PolicyProgram{}, err
	}
	return PolicyProgram{Policy: policy, GoPhysicalLines: goLines, GoooPhysicalLines: goooLines, Evidence: PolicyEvidence{
		Schema: PolicySchema, PolicyID: policy.ID.String(), SourceDigest: digestBytes(raw), CanonicalDigest: digestBytes([]byte(canonical)), SemanticIRDigest: digestBytes([]byte(policy.SemanticCanonical())),
		StateCount: len(policy.States), TransitionCount: len(policy.Transitions), CaseCount: len(policy.Cases), ClosedCases: 3, UnknownCases: 3, RefutedCases: 3,
	}}, nil
}

func validatePolicy(policy semantic.Policy) error {
	if policy.Name != PolicyName || policy.ID.String() != ContractID || len(policy.States) != 12 || len(policy.Transitions) != 9 || len(policy.Cases) != 9 {
		return errors.New("continuation policy denominator or identity drifted")
	}
	counts := map[string]int{}
	for _, current := range policy.Cases {
		counts[current.Resolution.Decision]++
		if current.Resolution.Decision == string(DecisionUnknown) && (current.Resolution.Stage == "" || current.Resolution.Step == "" || current.Resolution.Reason == "" || current.Resolution.UnknownClass == "" || current.Resolution.NextOperation == "" || len(current.Resolution.BlockedBy) == 0) {
			return fmt.Errorf("unknown continuation case %q lacks six fields", current.Name)
		}
	}
	if counts[string(DecisionClosed)] != 3 || counts[string(DecisionUnknown)] != 3 || counts[string(DecisionRefuted)] != 3 {
		return errors.New("continuation policy partition drifted")
	}
	return nil
}
