package selfimprovementexecutioncontract

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
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                             = "gooo/self-improvement-execution-contract/v1"
	PolicySchema                       = "gooo/self-improvement-execution-contract-policy/v1"
	CanonicalCasesSchema               = "gooo/self-improvement-execution-contract-cases/v1"
	VerificationSchema                 = "gooo/self-improvement-execution-contract-verification/v1"
	ContractID                         = "gooo://self-improvement/candidate-execution-contract/v1"
	PolicyName                         = "CandidateExecutionContract"
	PolicyPath                         = "examples/self-improvement-execution-contract/contract.gooo"
	CandidateExperimentKind            = "VALUE_WITNESS_EXPERIMENT"
	CandidateMetaOperation             = "propose-value-level-witness-experiment"
	KnownPhase                         = "VALUE_WITNESS_DECLARATION"
	KnownOperationID                   = "self-improvement.value-witness-experiment.v1"
	KnownBoundedTarget                 = BoundedTarget("VALUE_WITNESS_EXPERIMENT")
	ExecutionGrantBlockedBy            = "separate_execution_grant"
	PreExecutionRequiredField          = 12
	PerformanceUnknown                 = "UNKNOWN"
	InputAuthorityName                 = CallerOwnedInput
	OutputAuthorityName                = CallerOwnedOutput
	FixedEvaluatorRegistryDigest       = "sha256:7149e030bf738191961becd518a03b42f2d5714e4815c4835a3ba10d85b29b0a"
	FixedToolchainTestContractIdentity = "sha256:0be1f592782114f96b3422c3622cc57593c9db4e13dea376bd6e7065ccdb8acd"
)

type Decision string

const (
	DecisionClosed  Decision = "CLOSED"
	DecisionUnknown Decision = "UNKNOWN"
	DecisionRefuted Decision = "REFUTED"
)

type Resolution string

const (
	ResolutionDeclared Resolution = "DECLARED"
	ResolutionLower    Resolution = "LOWER_RESOLUTION"
	ResolutionExact    Resolution = "EXACT"
)

type Phase string
type OperationID string
type BoundedTarget string

type RootAuthority string

const (
	CallerOwnedInput  RootAuthority = "CALLER_OWNED_INPUT"
	CallerOwnedOutput RootAuthority = "CALLER_OWNED_OUTPUT"
)

type SourceSpan struct {
	SourceID  string `json:"source_id"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type CandidateEvidence struct {
	StableID          string `json:"candidate_stable_id"`
	Digest            string `json:"candidate_digest"`
	SubjectSHA        string `json:"subject_sha"`
	ObservationDigest string `json:"observation_digest"`
	InputDigest       string `json:"candidate_input_digest"`
	ExperimentKind    string `json:"experiment_kind"`
	MetaOperation     string `json:"meta_operation"`
}

type AuthorizationBinding struct {
	RequestSchema       string `json:"request_schema"`
	RequestDigest       string `json:"request_digest"`
	ContractID          string `json:"authorization_contract_id"`
	ContractDigest      string `json:"authorization_contract_digest"`
	Scope               string `json:"authorization_scope"`
	Decision            string `json:"authorization_decision"`
	ExecutionAllowed    bool   `json:"execution_allowed"`
	RepositoryWrites    int    `json:"repository_writes"`
	LocalTestExecutions int    `json:"local_test_executions"`
	LiveAuthorized      int    `json:"live_authorized"`
	LiveState           string `json:"live_state"`
}

type ObservationEvidence struct {
	Schema               string        `json:"schema"`
	ObservationDigest    string        `json:"observation_digest"`
	CandidateInputDigest string        `json:"candidate_input_digest"`
	CandidateStableID    string        `json:"candidate_stable_id"`
	SubjectSHA           string        `json:"subject_sha"`
	Phase                Phase         `json:"phase"`
	OperationID          OperationID   `json:"operation_id"`
	BoundedTarget        BoundedTarget `json:"bounded_target"`
	SourceSpans          []SourceSpan  `json:"source_spans"`
	ObservedCount        int           `json:"observed_count"`
	ObservedCountKnown   bool          `json:"observed_count_known"`
}

type RegistryEvidence struct {
	Schema                        string        `json:"schema"`
	EvaluatorRegistryDigest       string        `json:"evaluator_registry_digest"`
	ToolchainTestContractIdentity string        `json:"toolchain_test_contract_identity"`
	Phase                         Phase         `json:"phase"`
	OperationID                   OperationID   `json:"operation_id"`
	BoundedTarget                 BoundedTarget `json:"bounded_target"`
	InputAuthority                RootAuthority `json:"input_authority"`
	OutputAuthority               RootAuthority `json:"output_authority"`
	MaxExecutions                 int           `json:"max_executions"`
	RepositoryWritesAllowed       bool          `json:"repository_writes_allowed"`
	SafetyDeclared                bool          `json:"safety_declared"`
}

type ContractInput struct {
	Candidate     CandidateEvidence    `json:"candidate"`
	Authorization AuthorizationBinding `json:"authorization"`
	Observation   ObservationEvidence  `json:"observation"`
	Registry      RegistryEvidence     `json:"registry"`
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

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Metrics struct {
	PreExecutionRequiredFields           int    `json:"pre_execution_required_fields"`
	BoundFields                          int    `json:"bound_fields"`
	MissingFields                        int    `json:"missing_fields"`
	ContradictoryFields                  int    `json:"contradictory_fields"`
	StructuralCandidateToOperationBefore int    `json:"structural_candidate_to_operation_edges_before"`
	StructuralCandidateToOperationAfter  int    `json:"structural_candidate_to_operation_edges_after"`
	LiveExecutionCount                   int    `json:"live_execution_count"`
	CanonicalExecutionCount              int    `json:"canonical_execution_count"`
	ExecutionGrants                      int    `json:"execution_grants"`
	RepositoryWrites                     int    `json:"repository_writes"`
	LocalTestExecutions                  int    `json:"local_test_executions"`
	FallbackAccepted                     int    `json:"fallback_accepted"`
	IndependentReplayComparisons         int    `json:"independent_replay_comparisons"`
	ArtifactFiles                        int    `json:"artifact_files"`
	ArtifactTypes                        int    `json:"artifact_types"`
	GoPhysicalLines                      int    `json:"go_physical_lines"`
	GoooPhysicalLines                    int    `json:"gooo_physical_lines"`
	PerformanceImprovement               string `json:"performance_improvement"`
}

type ContractResolution struct {
	Schema                  string         `json:"schema"`
	ContractID              string         `json:"contract_id"`
	Policy                  PolicyEvidence `json:"policy"`
	RequiredFields          []string       `json:"required_fields"`
	CandidateStableID       string         `json:"candidate_stable_id"`
	CandidateDigest         string         `json:"candidate_digest"`
	SubjectSHA              string         `json:"subject_sha"`
	ObservationDigest       string         `json:"observation_digest"`
	CandidateInputDigest    string         `json:"candidate_input_digest"`
	Phase                   Phase          `json:"phase"`
	OperationID             OperationID    `json:"operation_id"`
	BoundedTarget           BoundedTarget  `json:"bounded_target"`
	EvaluatorRegistryDigest string         `json:"evaluator_registry_digest"`
	ToolchainTestContractID string         `json:"toolchain_test_contract_identity"`
	Decision                Decision       `json:"decision"`
	Resolution              Resolution     `json:"resolution"`
	Reason                  string         `json:"reason"`
	Unknown                 *UnknownState  `json:"unknown,omitempty"`
	MissingFields           []string       `json:"missing_fields,omitempty"`
	ContradictoryFields     []string       `json:"contradictory_fields,omitempty"`
	InputAuthority          RootAuthority  `json:"input_authority"`
	OutputAuthority         RootAuthority  `json:"output_authority"`
	MaxExecutions           int            `json:"max_executions"`
	RepositoryWritesAllowed bool           `json:"repository_writes_allowed"`
	ExecutionAuthorized     bool           `json:"execution_authorized"`
	ExecutionGrantRequired  bool           `json:"execution_grant_required"`
	ExecutionGrantBlockedBy string         `json:"execution_grant_blocked_by"`
	OutputEvidenceDeferred  bool           `json:"output_evidence_deferred"`
	RuntimeResultDeferred   bool           `json:"runtime_result_deferred"`
	Metrics                 Metrics        `json:"metrics"`
	Digest                  string         `json:"digest"`
}

type Verification struct {
	Schema                       string     `json:"schema"`
	ContractDigest               string     `json:"contract_digest"`
	IndependentDecision          Decision   `json:"independent_decision"`
	IndependentResolution        Resolution `json:"independent_resolution"`
	IndependentReason            string     `json:"independent_reason"`
	Verified                     bool       `json:"verified"`
	IndependentReplayComparisons int        `json:"independent_replay_comparisons"`
	RepositoryWrites             int        `json:"repository_writes"`
	LocalTestExecutions          int        `json:"local_test_executions"`
	ExecutionGrants              int        `json:"execution_grants"`
	Digest                       string     `json:"digest"`
}

type LiveReport struct {
	ContractResolution
	Verification Verification `json:"verification"`
}

type CanonicalCase struct {
	ID                  string        `json:"id"`
	ExpectedDecision    Decision      `json:"expected_decision"`
	ExpectedReason      string        `json:"expected_reason"`
	ActualDecision      Decision      `json:"actual_decision"`
	ActualReason        string        `json:"actual_reason"`
	Unknown             *UnknownState `json:"unknown,omitempty"`
	MissingFields       []string      `json:"missing_fields,omitempty"`
	ContradictoryFields []string      `json:"contradictory_fields,omitempty"`
	Pass                bool          `json:"pass"`
}

type CanonicalCaseReport struct {
	Schema                               string          `json:"schema"`
	Policy                               PolicyEvidence  `json:"policy"`
	RequiredFields                       []string        `json:"required_fields"`
	CaseDenominator                      int             `json:"case_denominator"`
	BoundFields                          int             `json:"bound_fields"`
	MissingFields                        int             `json:"missing_fields"`
	ContradictoryFields                  int             `json:"contradictory_fields"`
	StructuralCandidateToOperationBefore int             `json:"structural_candidate_to_operation_edges_before"`
	StructuralCandidateToOperationAfter  int             `json:"structural_candidate_to_operation_edges_after"`
	ClosedCases                          int             `json:"closed_cases"`
	UnknownCases                         int             `json:"unknown_cases"`
	RefutedCases                         int             `json:"refuted_cases"`
	Counts                               map[string]int  `json:"counts"`
	Cases                                []CanonicalCase `json:"cases"`
	ReplayEqual                          bool            `json:"replay_equal"`
	LiveExecutionCount                   int             `json:"live_execution_count"`
	CanonicalExecutionCount              int             `json:"canonical_execution_count"`
	ExecutionGrants                      int             `json:"execution_grants"`
	RepositoryWrites                     int             `json:"repository_writes"`
	LocalTestExecutions                  int             `json:"local_test_executions"`
	FallbackAccepted                     int             `json:"fallback_accepted"`
	IndependentReplayComparisons         int             `json:"independent_replay_comparisons"`
	ArtifactFiles                        int             `json:"artifact_files"`
	ArtifactTypes                        int             `json:"artifact_types"`
	GoPhysicalLines                      int             `json:"go_physical_lines"`
	GoooPhysicalLines                    int             `json:"gooo_physical_lines"`
	PerformanceImprovement               string          `json:"performance_improvement"`
	Decision                             Decision        `json:"decision"`
	Resolution                           Resolution      `json:"resolution"`
	Reason                               string          `json:"reason"`
	Digest                               string          `json:"digest"`
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

func cloneStrings(values []string) []string { return append([]string(nil), values...) }

func sortedStrings(values []string) []string {
	result := cloneStrings(values)
	sort.Strings(result)
	return result
}

func clearResolutionDigest(value ContractResolution) ContractResolution {
	value.Digest = ""
	return value
}

func resolutionDigest(value ContractResolution) string {
	return digestJSON(clearResolutionDigest(value))
}

func clearVerificationDigest(value Verification) Verification {
	value.Digest = ""
	return value
}

func verificationDigest(value Verification) string { return digestJSON(clearVerificationDigest(value)) }

func clearCanonicalDigest(value CanonicalCaseReport) CanonicalCaseReport {
	value.Digest = ""
	return value
}

func canonicalDigest(value CanonicalCaseReport) string {
	return digestJSON(clearCanonicalDigest(value))
}

func KnownRegistry() RegistryEvidence {
	return RegistryEvidence{
		Schema:                        "gooo/self-improvement-execution-operation-registry/v1",
		EvaluatorRegistryDigest:       FixedEvaluatorRegistryDigest,
		ToolchainTestContractIdentity: FixedToolchainTestContractIdentity,
		Phase:                         Phase(KnownPhase), OperationID: OperationID(KnownOperationID), BoundedTarget: KnownBoundedTarget,
		InputAuthority: CallerOwnedInput, OutputAuthority: CallerOwnedOutput,
		MaxExecutions: 1, RepositoryWritesAllowed: false, SafetyDeclared: true,
	}
}

func ProjectAuthorizationRequest(request selfimprovementcandidate.AuthorizationRequest, registry RegistryEvidence) ContractInput {
	return ContractInput{
		Candidate: CandidateEvidence{
			StableID: request.Candidate.CandidateID, Digest: request.Candidate.CandidateDigest,
			SubjectSHA: request.Candidate.SubjectSHA, ObservationDigest: request.Candidate.SourceObservationDigest,
			InputDigest: request.Candidate.InputSourceDigest,
		},
		Authorization: AuthorizationBinding{
			RequestSchema: request.Schema, RequestDigest: request.Digest,
			ContractID: request.Candidate.ContractID, ContractDigest: request.Candidate.ContractCanonicalDigest,
			Scope: request.Candidate.Scope, ExecutionAllowed: request.ExecutionAllowed,
			RepositoryWrites: request.RepositoryWrites, LocalTestExecutions: request.LocalTestExecutions,
			LiveAuthorized: request.LiveAuthorized, LiveState: request.LiveState,
		},
		Registry: registry,
	}
}

func RequiredFieldNames() []string {
	return []string{
		"candidate_stable_id", "candidate_digest", "subject_sha", "observation_digest",
		"candidate_input_digest", "phase", "operation_id", "bounded_target", "source_spans",
		"observed_count", "evaluator_registry_digest", "toolchain_test_contract_identity",
	}
}

func CompilePolicy(repository fs.FS, path string) (PolicyProgram, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("read execution contract policy: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return PolicyProgram{}, errors.New("execution contract policy has syntax errors")
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("format execution contract policy: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyProgram{}, fmt.Errorf("lower execution contract policy: %w", err)
	}
	if len(ir.Policies) != 1 {
		return PolicyProgram{}, errors.New("execution contract policy must lower to one semantic policy")
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
		return errors.New("execution contract policy denominator or identity drifted")
	}
	counts := map[string]int{}
	for _, current := range policy.Cases {
		counts[current.Resolution.Decision]++
		if current.Resolution.Decision == string(DecisionUnknown) &&
			(current.Resolution.Stage == "" || current.Resolution.Reason == "" || current.Resolution.UnknownClass == "" || current.Resolution.NextOperation == "" || len(current.Resolution.BlockedBy) == 0) {
			return fmt.Errorf("unknown policy case %q does not declare all six fields", current.Name)
		}
	}
	if counts[string(DecisionClosed)] != 3 || counts[string(DecisionUnknown)] != 3 || counts[string(DecisionRefuted)] != 3 {
		return errors.New("execution contract policy case partition drifted")
	}
	return nil
}
