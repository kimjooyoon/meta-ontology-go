package selfimprovementcandidate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	AuthorizationContractSchema   = "gooo/self-improvement-candidate-authorization-contract/v2"
	AuthorizationRequestSchema    = "gooo/self-improvement-candidate-authorization-request/v2"
	AuthorizationDecisionSchema   = "gooo/self-improvement-candidate-authorization-decision/v2"
	AuthorizationResolutionSchema = "gooo/self-improvement-candidate-authorization-resolution/v2"
	AuthorizationCasesSchema      = "gooo/self-improvement-candidate-authorization-cases/v2"
	AuthorizationContractID       = "gooo://self-improvement/candidate-authorization/v2"
	AuthorizationScope            = "candidate-eligibility-only"
	AuthorizationRequestState     = "REQUESTED"
	AuthorizationUnresolved       = "UNRESOLVED"
	AuthorizationClosed           = "CLOSED"
	AuthorizationUnknown          = "UNKNOWN"
	AuthorizationRefuted          = "REFUTED"
	AuthorizationAllow            = "ALLOW"
	AuthorizationAuthorized       = "AUTHORIZED"
	AuthorizationUnknownReason    = "SYSTEM_AUTHORIZATION_EVIDENCE_INCOMPLETE"
	AuthorizationUnknownStage     = "ELIGIBILITY"
	AuthorizationUnknownStep      = "DERIVE_FROM_EVIDENCE"
	AuthorizationUnknownClass     = "INCOMPLETE_EVIDENCE"
	AuthorizationUnknownNext      = "COMPLETE_SYSTEM_AUTHORIZATION_EVIDENCE"
	AuthorizationRefutedReason    = "CANDIDATE_EVIDENCE_CONTRADICTION"
	AuthorizationClosedAllow      = "EXACT_SYSTEM_DERIVED_CANDIDATE_ELIGIBILITY"
	AuthorizationArtifactReason   = "CANDIDATE_ARTIFACT_UNAVAILABLE"
	AuthorizationFieldReason      = "MISSING_CANDIDATE_ADAPTER_FIELD"
	AuthorizationRequestReason    = "ELIGIBILITY_REQUEST_INTEGRITY_MISMATCH"
	AuthorizationContractReason   = "AUTHORIZATION_CONTRACT_INTEGRITY_MISMATCH"
	AuthorizationEvidenceRule     = "EXACT_PROPOSED_CANDIDATE_V1"
	AuthorizationEvidenceSource   = "system-derived-exact-evidence"
)

// AuthorizationContract is the compiled semantic IR for authorization.gooo.
// It is deliberately small: the contract declares the request, decision, and
// resolution transitions, while the typed evaluator below owns their exact
// binding and resolution semantics.
type AuthorizationContract struct {
	Schema             string `json:"schema"`
	ContractID         string `json:"contract_id"`
	Path               string `json:"path"`
	Package            string `json:"package"`
	Namespace          string `json:"namespace"`
	EntityCount        int    `json:"entity_count"`
	ActivityCount      int    `json:"activity_count"`
	SourceDigest       string `json:"source_digest"`
	CanonicalDigest    string `json:"canonical_digest"`
	RequestActivity    string `json:"request_activity"`
	DecisionActivity   string `json:"decision_activity"`
	ResolutionActivity string `json:"resolution_activity"`
}

type CandidateBinding struct {
	ReportSchema            string                            `json:"report_schema"`
	CandidateSchema         string                            `json:"candidate_schema"`
	CandidateID             string                            `json:"candidate_id"`
	CandidateDigest         string                            `json:"candidate_digest"`
	CandidateReportDigest   string                            `json:"candidate_report_digest"`
	SourceObservationDigest string                            `json:"source_observation_digest"`
	InputSourceDigest       string                            `json:"input_source_digest"`
	ExecutionInputDigest    string                            `json:"execution_input_digest"`
	ExecutionInput          *valuewitnessinput.ExecutionInput `json:"execution_input,omitempty"`
	ExperimentKind          string                            `json:"experiment_kind"`
	MetaOperation           string                            `json:"meta_operation"`
	SubjectSHA              string                            `json:"subject_sha"`
	SourceWorkflowRunID     int64                             `json:"source_workflow_run_id"`
	PolicyVersion           string                            `json:"policy_version"`
	PolicyDigest            string                            `json:"policy_digest"`
	ContractID              string                            `json:"contract_id"`
	ContractSourceDigest    string                            `json:"contract_source_digest"`
	ContractCanonicalDigest string                            `json:"contract_canonical_digest"`
	Scope                   string                            `json:"scope"`
	ScopeDigest             string                            `json:"scope_digest"`
	Authority               Authority                         `json:"authority"`
}

type ArtifactMetadata struct {
	Repository string `json:"repository"`
	SubjectSHA string `json:"subject_sha"`
	// RunID identifies the workflow run that uploaded this artifact. It is
	// distinct from CandidateBinding.SourceWorkflowRunID, which identifies the
	// source observation recorded in the candidate report.
	RunID         int64  `json:"run_id"`
	RunAttempt    int    `json:"run_attempt"`
	ArtifactID    int64  `json:"artifact_id"`
	ArtifactName  string `json:"artifact_name"`
	ArchiveDigest string `json:"archive_digest"`
	SizeBytes     int64  `json:"size_bytes"`
	Expired       bool   `json:"expired"`
}

type AuthorizationMetrics struct {
	StructuralUnboundEdgesBefore int `json:"structural_unbound_edges_before"`
	StructuralUnboundEdgesAfter  int `json:"structural_unbound_edges_after"`
	AdapterEdges                 int `json:"adapter_edges"`
	Requests                     int `json:"requests"`
	Decisions                    int `json:"decisions"`
	SystemDerivedDecisions       int `json:"system_derived_decisions"`
	UnknownSixField              int `json:"unknown_six_field"`
	RefutedContradictions        int `json:"refuted_contradictions"`
	FallbackAccepted             int `json:"fallback_accepted"`
	IndependentReplayComparisons int `json:"independent_replay_comparisons"`
	ArtifactFiles                int `json:"artifact_files"`
	ArtifactTypes                int `json:"artifact_types"`
	RepositoryWrites             int `json:"repository_writes"`
	LocalTestExecutions          int `json:"local_test_executions"`
}

// AuthorizationRequest is the request-side semantic IR. It is not a human
// decision and never grants execution or repository mutation.
type AuthorizationRequest struct {
	Schema              string                `json:"schema"`
	Lifecycle           string                `json:"lifecycle"`
	Contract            AuthorizationContract `json:"contract"`
	Candidate           CandidateBinding      `json:"candidate"`
	Artifact            ArtifactMetadata      `json:"artifact"`
	Target              string                `json:"target"`
	Mode                string                `json:"mode"`
	ExecutionAllowed    bool                  `json:"execution_allowed"`
	RepositoryWrites    int                   `json:"repository_writes"`
	LocalTestExecutions int                   `json:"local_test_executions"`
	Decision            string                `json:"decision"`
	Resolution          string                `json:"resolution"`
	Reason              string                `json:"reason"`
	LiveAuthorized      int                   `json:"live_authorized"`
	LiveState           string                `json:"live_state"`
	Metrics             AuthorizationMetrics  `json:"metrics"`
	Digest              string                `json:"digest"`
}

// AuthorizationDecisionInput is a deterministic system evidence receipt. It
// is derived from the exact request and cannot be supplied by a person or
// override the evaluator's decision.
type AuthorizationDecisionInput struct {
	Schema                  string `json:"schema"`
	Decision                string `json:"decision"`
	RequestDigest           string `json:"request_digest"`
	CandidateID             string `json:"candidate_id"`
	CandidateDigest         string `json:"candidate_digest"`
	CandidateReportDigest   string `json:"candidate_report_digest"`
	SourceObservationDigest string `json:"source_observation_digest"`
	InputSourceDigest       string `json:"input_source_digest"`
	ExecutionInputDigest    string `json:"execution_input_digest"`
	PolicyDigest            string `json:"policy_digest"`
	ContractDigest          string `json:"contract_digest"`
	SubjectSHA              string `json:"subject_sha"`
	ScopeDigest             string `json:"scope_digest"`
	Repository              string `json:"repository"`
	CandidateArtifactID     int64  `json:"candidate_artifact_id"`
	CandidateArtifactDigest string `json:"candidate_artifact_digest"`
	DecisionRule            string `json:"decision_rule"`
	DecisionSource          string `json:"decision_source"`
	EvidenceDigest          string `json:"evidence_digest"`
	DecisionDigest          string `json:"decision_digest,omitempty"`
}

// AuthorizationReceipt embeds the released semantic-self-adoption
// authorization contract and adds candidate-specific system evidence. The
// receipt records eligibility only; execution remains explicitly disabled.
type AuthorizationReceipt struct {
	generation.SemanticAdoptionAuthorization
	Decision                string `json:"decision"`
	CandidateID             string `json:"candidate_id"`
	CandidateDigest         string `json:"candidate_digest"`
	CandidateReportDigest   string `json:"candidate_report_digest"`
	SourceObservationDigest string `json:"source_observation_digest"`
	ExecutionInputDigest    string `json:"execution_input_digest"`
	SubjectSHA              string `json:"subject_sha"`
	PolicyVersion           string `json:"policy_version"`
	PolicyDigest            string `json:"policy_digest"`
	ContractID              string `json:"contract_id"`
	ContractSourceDigest    string `json:"contract_source_digest"`
	ContractCanonicalDigest string `json:"contract_canonical_digest"`
	Scope                   string `json:"scope"`
	ScopeDigest             string `json:"scope_digest"`
	ExecutionAuthorized     bool   `json:"execution_authorized"`
	CandidateArtifactID     int64  `json:"candidate_artifact_id"`
	CandidateArtifactDigest string `json:"candidate_artifact_digest"`
	Repository              string `json:"repository"`
	DecisionSource          string `json:"decision_source"`
	DecisionRule            string `json:"decision_rule"`
	EvidenceDigest          string `json:"evidence_digest"`
	DecisionInputDigest     string `json:"decision_input_digest"`
}

// AuthorizationResolution is the decision-side semantic IR and independent
// consumer receipt. It emits a system-derived eligibility receipt only for an
// exact candidate binding; execution and repository mutation remain disabled.
type AuthorizationResolution struct {
	Schema              string                           `json:"schema"`
	Lifecycle           string                           `json:"lifecycle"`
	RequestDigest       string                           `json:"request_digest"`
	Candidate           CandidateBinding                 `json:"candidate"`
	Artifact            ArtifactMetadata                 `json:"artifact"`
	Decision            string                           `json:"decision"`
	Outcome             string                           `json:"outcome"`
	Resolution          string                           `json:"resolution"`
	Reason              string                           `json:"reason"`
	Unknown             *generation.EnvelopeUnknownState `json:"unknown,omitempty"`
	DecisionInputs      []AuthorizationDecisionInput     `json:"decision_evidence,omitempty"`
	Authorization       *AuthorizationReceipt            `json:"authorization,omitempty"`
	LiveAuthorized      int                              `json:"live_authorized"`
	LiveState           string                           `json:"live_state"`
	ExecutionAuthorized bool                             `json:"execution_authorized"`
	Metrics             AuthorizationMetrics             `json:"metrics"`
	RepositoryWrites    int                              `json:"repository_writes"`
	LocalTestExecutions int                              `json:"local_test_executions"`
	Digest              string                           `json:"digest"`
}

type AuthorizationIR struct {
	Contract   AuthorizationContract       `json:"contract"`
	Request    AuthorizationRequest        `json:"request"`
	Decision   *AuthorizationDecisionInput `json:"decision,omitempty"`
	Resolution AuthorizationResolution     `json:"resolution"`
}

type AuthorizationVerification struct {
	Schema              string `json:"schema"`
	RequestDigest       string `json:"request_digest"`
	ResolutionDigest    string `json:"resolution_digest"`
	Decision            string `json:"decision"`
	Resolution          string `json:"resolution"`
	DecisionVerified    bool   `json:"decision_verified"`
	ExecutionAuthorized bool   `json:"execution_authorized"`
	RepositoryWrites    int    `json:"repository_writes"`
	LocalTestExecutions int    `json:"local_test_executions"`
	LiveAuthorized      int    `json:"live_authorized"`
	LiveState           string `json:"live_state"`
	Digest              string `json:"digest"`
}

type CanonicalCase struct {
	ID               string                           `json:"id"`
	ExpectedDecision string                           `json:"expected_decision"`
	ExpectedReason   string                           `json:"expected_reason"`
	ActualDecision   string                           `json:"actual_decision"`
	ActualReason     string                           `json:"actual_reason"`
	Unknown          *generation.EnvelopeUnknownState `json:"unknown,omitempty"`
	Pass             bool                             `json:"pass"`
}

type CanonicalCaseReport struct {
	Schema          string                         `json:"schema"`
	RequestDigest   string                         `json:"request_digest"`
	CaseDenominator int                            `json:"case_denominator"`
	ClosedCases     int                            `json:"closed_cases"`
	UnknownCases    int                            `json:"unknown_cases"`
	RefutedCases    int                            `json:"refuted_cases"`
	Counts          map[string]int                 `json:"counts"`
	Cases           []CanonicalCase                `json:"cases"`
	ReplayEqual     bool                           `json:"replay_equal"`
	LiveAuthorized  int                            `json:"live_authorized"`
	LiveState       string                         `json:"live_state"`
	Metrics         AuthorizationMetrics           `json:"metrics"`
	Decision        string                         `json:"decision"`
	Resolution      string                         `json:"resolution"`
	Reason          string                         `json:"reason"`
	Roundtrip       AuthorizationRoundtripEvidence `json:"roundtrip"`
	Digest          string                         `json:"digest"`
}

// AuthorizationRoundtripEvidence records the v27A.1 regression boundary. The
// first post-merge authorization run is retained as the counterexample for
// pointer-address comparison; the repaired verifier closes the same value
// after JSON serialization without adding execution authority.
type AuthorizationRoundtripEvidence struct {
	AuthorizationRoundtripExactBefore int   `json:"authorization_roundtrip_exact_before"`
	AuthorizationRoundtripExactAfter  int   `json:"authorization_roundtrip_exact_after"`
	PointerIdentityDependencyBefore   int   `json:"pointer_identity_dependency_before"`
	PointerIdentityDependencyAfter    int   `json:"pointer_identity_dependency_after"`
	CounterexampleRunID               int64 `json:"counterexample_run_id"`
}

func compileAuthorizationContract(repository fs.FS, path string) (AuthorizationContract, error) {
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		return AuthorizationContract{}, fmt.Errorf("read authorization contract: %w", err)
	}
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return AuthorizationContract{}, errors.New("authorization contract has syntax errors")
	}
	canonical, err := syntax.Format(file)
	if err != nil || !authorizationDeclarationsKnown(file) {
		return AuthorizationContract{}, errors.New("authorization contract declarations are invalid")
	}
	return AuthorizationContract{
		Schema: AuthorizationContractSchema, ContractID: AuthorizationContractID, Path: path,
		Package: file.Package.Name, Namespace: file.Namespace.Name,
		EntityCount: 4, ActivityCount: 3, SourceDigest: digestBytes(raw),
		CanonicalDigest:    digestBytes([]byte(canonical)),
		RequestActivity:    "RequestCandidateAuthorization",
		DecisionActivity:   "DeriveCandidateAuthorization",
		ResolutionActivity: "ResolveCandidateAuthorization",
	}, nil
}

func authorizationDeclarationsKnown(file *syntax.File) bool {
	if file == nil || file.Package == nil || file.Namespace == nil ||
		file.Package.Name != "selfimprovement" || file.Namespace.Name != "selfimprovement" {
		return false
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	if len(declarations) != 7 {
		return false
	}
	entities := map[string]string{}
	activities := map[string][2]string{}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value.FieldsPresent || len(value.Fields) != 0 || value.Name == "" || value.ID == "" {
				return false
			}
			if _, duplicate := entities[value.Name]; duplicate {
				return false
			}
			entities[value.Name] = value.ID
		case *syntax.ActivityDecl:
			inputs := value.Inputs
			if inputs == nil {
				inputs = value.Parameters
			}
			if len(inputs) != 1 || value.Output == "" {
				return false
			}
			if _, duplicate := activities[value.Name]; duplicate {
				return false
			}
			activities[value.Name] = [2]string{inputs[0].Name, value.Output}
		default:
			return false
		}
	}
	expectedEntities := map[string]string{
		"NonExecutingImprovementCandidate": "gooo://self-improvement/entity/non-executing-improvement-candidate",
		"CandidateAuthorizationRequest":    "gooo://self-improvement/entity/candidate-authorization-request",
		"CandidateAuthorizationDecision":   "gooo://self-improvement/entity/system-derived-candidate-authorization-decision",
		"CandidateAuthorizationResolution": "gooo://self-improvement/entity/candidate-authorization-resolution",
	}
	expectedActivities := map[string][2]string{
		"RequestCandidateAuthorization": {"NonExecutingImprovementCandidate", "CandidateAuthorizationRequest"},
		"DeriveCandidateAuthorization":  {"CandidateAuthorizationRequest", "CandidateAuthorizationDecision"},
		"ResolveCandidateAuthorization": {"CandidateAuthorizationDecision", "CandidateAuthorizationResolution"},
	}
	return mapsEqual(entities, expectedEntities) && mapsEqual(activities, expectedActivities)
}

func decodeCandidate(raw []byte) (Report, error) {
	var report Report
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return Report{}, fmt.Errorf("decode candidate: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Report{}, errors.New("decode candidate: trailing content")
	}
	return report, nil
}

func BuildAuthorizationRequest(repository fs.FS, contractPath string, candidateRaw []byte, metadata ArtifactMetadata) (AuthorizationRequest, error) {
	contract, err := compileAuthorizationContract(repository, contractPath)
	if err != nil {
		return AuthorizationRequest{}, err
	}
	report, err := decodeCandidate(candidateRaw)
	if err != nil {
		return AuthorizationRequest{}, err
	}
	if err := Validate(report, report.SubjectSHA, report.SourceWorkflowRunID); err != nil || report.Decision != DecisionProposed {
		return AuthorizationRequest{}, errors.New("candidate is not an exact proposed receipt")
	}
	if !validArtifactMetadata(metadata) {
		return AuthorizationRequest{}, errors.New("candidate artifact metadata is unavailable")
	}
	if report.SubjectSHA != metadata.SubjectSHA {
		return AuthorizationRequest{}, errors.New("candidate report subject does not match its artifact")
	}
	candidate := report.Candidates[0]
	binding := CandidateBinding{
		ReportSchema: ReportSchema, CandidateSchema: candidate.Schema, CandidateID: candidate.ID,
		CandidateDigest: candidate.Digest, CandidateReportDigest: digestBytes(candidateRaw),
		SourceObservationDigest: report.SourceObservationDigest, InputSourceDigest: report.SourceFileDigest,
		ExecutionInputDigest: candidate.ExecutionInputDigest, ExecutionInput: report.ExecutionInput,
		ExperimentKind: candidate.ExperimentKind, MetaOperation: candidate.MetaOperation,
		SubjectSHA: report.SubjectSHA, SourceWorkflowRunID: report.SourceWorkflowRunID,
		PolicyVersion: report.PolicyVersion, PolicyDigest: digestBytes([]byte(report.PolicyVersion)),
		ContractID: report.Contract.ContractID, ContractSourceDigest: report.Contract.SourceDigest,
		ContractCanonicalDigest: report.Contract.CanonicalDigest, Scope: AuthorizationScope,
		ScopeDigest: digestBytes([]byte(AuthorizationScope)), Authority: report.Authority,
	}
	request := AuthorizationRequest{
		Schema: AuthorizationRequestSchema, Lifecycle: AuthorizationRequestState, Contract: contract,
		Candidate: binding, Artifact: metadata, Target: generation.SemanticAdoptionTarget,
		Mode: generation.SemanticAdoptionMode, ExecutionAllowed: false, RepositoryWrites: 0,
		LocalTestExecutions: 0, Decision: AuthorizationRequestState, Resolution: AuthorizationUnresolved,
		Reason: "CANDIDATE_AUTHORIZATION_REQUESTED", LiveAuthorized: 0, LiveState: "UNKNOWN",
		Metrics: AuthorizationMetrics{StructuralUnboundEdgesBefore: 1, StructuralUnboundEdgesAfter: 0,
			AdapterEdges: 1, Requests: 1, ArtifactFiles: 2, ArtifactTypes: 2},
	}
	request.Digest = requestDigest(request)
	return request, nil
}

// BuildUnavailableAuthorizationResolution preserves a six-field UNKNOWN when
// the candidate artifact cannot be read. It intentionally has no request or
// authorization payload because the candidate binding was never observed.
func BuildUnavailableAuthorizationResolution(metadata ArtifactMetadata) AuthorizationResolution {
	requestDigest := metadata.ArchiveDigest
	if !validDigest(requestDigest) {
		requestDigest = zeroDigest
	}
	resolution := AuthorizationResolution{
		Schema: AuthorizationResolutionSchema, Lifecycle: "AUTHORIZATION_RESOLUTION",
		RequestDigest: requestDigest, Artifact: metadata, Decision: AuthorizationUnknown,
		Resolution: ResolutionLower, Reason: AuthorizationArtifactReason,
		Unknown:        unknownState(AuthorizationArtifactReason, "RESTORE_CANDIDATE_ARTIFACT", []string{"candidate_artifact"}),
		LiveAuthorized: 0, LiveState: "UNKNOWN", RepositoryWrites: 0, LocalTestExecutions: 0,
		Metrics: AuthorizationMetrics{StructuralUnboundEdgesBefore: 1, StructuralUnboundEdgesAfter: 0,
			AdapterEdges: 1, Requests: 1, UnknownSixField: 1, ArtifactFiles: 0, ArtifactTypes: 0},
	}
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func requestDigest(request AuthorizationRequest) string {
	request.Digest = ""
	return digestJSON(request)
}

func resolutionDigest(resolution AuthorizationResolution) string {
	resolution.Digest = ""
	return digestJSON(resolution)
}

func verificationDigest(verification AuthorizationVerification) string {
	verification.Digest = ""
	return digestJSON(verification)
}

func decisionDigest(input AuthorizationDecisionInput) string {
	input.DecisionDigest = ""
	return digestJSON(input)
}

func validArtifactMetadata(metadata ArtifactMetadata) bool {
	return metadata.Repository != "" && validSHA(metadata.SubjectSHA) && metadata.RunID > 0 && metadata.RunAttempt > 0 &&
		metadata.ArtifactID > 0 && metadata.ArtifactName != "" && validDigest(metadata.ArchiveDigest) &&
		metadata.SizeBytes > 0 && !metadata.Expired
}

func incompleteAuthorizationContract(contract AuthorizationContract) bool {
	return contract.Schema == "" || contract.ContractID == "" || contract.Path == "" || contract.Package == "" ||
		contract.Namespace == "" || contract.SourceDigest == "" || contract.CanonicalDigest == ""
}

func validAuthorizationContract(contract AuthorizationContract) bool {
	return contract.Schema == AuthorizationContractSchema && contract.ContractID == AuthorizationContractID &&
		contract.Path == "examples/self-improvement/authorization.gooo" && contract.Package == "selfimprovement" &&
		contract.Namespace == "selfimprovement" && contract.EntityCount == 4 && contract.ActivityCount == 3 &&
		validDigest(contract.SourceDigest) && validDigest(contract.CanonicalDigest) &&
		contract.RequestActivity == "RequestCandidateAuthorization" &&
		contract.DecisionActivity == "DeriveCandidateAuthorization" &&
		contract.ResolutionActivity == "ResolveCandidateAuthorization"
}

func validCandidateBinding(binding CandidateBinding) bool {
	return binding.ReportSchema == ReportSchema && binding.CandidateSchema == CandidateSchema &&
		validDigest(binding.CandidateID) && validDigest(binding.CandidateDigest) &&
		validDigest(binding.CandidateReportDigest) && validDigest(binding.SourceObservationDigest) &&
		validDigest(binding.InputSourceDigest) && validDigest(binding.ExecutionInputDigest) &&
		binding.ExperimentKind == "VALUE_WITNESS_EXPERIMENT" && binding.MetaOperation == "propose-value-level-witness-experiment" &&
		binding.ExecutionInput != nil && valuewitnessinput.Validate(*binding.ExecutionInput) == nil &&
		binding.ExecutionInput.Digest == binding.ExecutionInputDigest && binding.ExecutionInput.CandidateStableID == binding.CandidateID &&
		binding.ExecutionInput.CandidateDigest == binding.CandidateDigest && binding.ExecutionInput.SubjectSHA == binding.SubjectSHA &&
		binding.ExecutionInput.ObservationDigest == binding.SourceObservationDigest && validSHA(binding.SubjectSHA) && binding.SourceWorkflowRunID > 0 &&
		binding.PolicyVersion == PolicyVersion && validDigest(binding.PolicyDigest) &&
		binding.ContractID == ContractID && validDigest(binding.ContractSourceDigest) &&
		validDigest(binding.ContractCanonicalDigest) && binding.Scope == AuthorizationScope &&
		validDigest(binding.ScopeDigest) && authorityClosed(binding.Authority)
}

func incompleteCandidateBinding(binding CandidateBinding) bool {
	return binding.ReportSchema == "" || binding.CandidateSchema == "" || binding.CandidateID == "" ||
		binding.CandidateDigest == "" || binding.CandidateReportDigest == "" || binding.SourceObservationDigest == "" ||
		binding.InputSourceDigest == "" || binding.ExecutionInputDigest == "" || binding.ExecutionInput == nil ||
		binding.ExperimentKind == "" || binding.MetaOperation == "" || binding.SubjectSHA == "" || binding.SourceWorkflowRunID == 0 ||
		binding.PolicyVersion == "" || binding.PolicyDigest == "" || binding.ContractID == "" ||
		binding.ContractSourceDigest == "" || binding.ContractCanonicalDigest == "" || binding.Scope == "" ||
		binding.ScopeDigest == ""
}

func requestCommonValid(request AuthorizationRequest) bool {
	return request.Schema == AuthorizationRequestSchema && request.Lifecycle == AuthorizationRequestState &&
		request.Decision == AuthorizationRequestState && request.Resolution == AuthorizationUnresolved &&
		request.Reason == "CANDIDATE_AUTHORIZATION_REQUESTED" && request.Target == generation.SemanticAdoptionTarget &&
		request.Mode == generation.SemanticAdoptionMode && !request.ExecutionAllowed && request.RepositoryWrites == 0 &&
		request.LocalTestExecutions == 0 && request.LiveAuthorized == 0 && request.LiveState == "UNKNOWN" &&
		request.Digest == requestDigest(request)
}

func unknownState(reason, next string, blockedBy []string) *generation.EnvelopeUnknownState {
	return &generation.EnvelopeUnknownState{Stage: AuthorizationUnknownStage, Step: AuthorizationUnknownStep,
		Reason: reason, UnknownClass: AuthorizationUnknownClass, NextOperation: next, BlockedBy: blockedBy}
}

func resolutionBase(request AuthorizationRequest) AuthorizationResolution {
	return AuthorizationResolution{Schema: AuthorizationResolutionSchema, Lifecycle: "AUTHORIZATION_RESOLUTION",
		RequestDigest: request.Digest, Candidate: request.Candidate, Artifact: request.Artifact,
		LiveAuthorized: 0, LiveState: "UNKNOWN", Metrics: AuthorizationMetrics{
			StructuralUnboundEdgesBefore: 1, StructuralUnboundEdgesAfter: 0, AdapterEdges: 1,
			Requests: 1, ArtifactFiles: 2, ArtifactTypes: 2}, RepositoryWrites: 0, LocalTestExecutions: 0}
}

func unknownResolution(request AuthorizationRequest, reason, next string, blockedBy []string) AuthorizationResolution {
	resolution := resolutionBase(request)
	resolution.Decision = AuthorizationUnknown
	resolution.Resolution = ResolutionLower
	resolution.Reason = reason
	resolution.Unknown = unknownState(reason, next, blockedBy)
	resolution.Metrics.UnknownSixField = 1
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func refutedResolution(request AuthorizationRequest, reason string, inputs []AuthorizationDecisionInput) AuthorizationResolution {
	resolution := resolutionBase(request)
	resolution.Decision = AuthorizationRefuted
	resolution.Resolution = ResolutionExact
	resolution.Reason = reason
	resolution.DecisionInputs = append([]AuthorizationDecisionInput(nil), inputs...)
	resolution.Metrics.RefutedContradictions = 1
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func systemEvidenceDigest(request AuthorizationRequest) string {
	return digestJSON(struct {
		RequestDigest string                `json:"request_digest"`
		Candidate     CandidateBinding      `json:"candidate"`
		Artifact      ArtifactMetadata      `json:"artifact"`
		Contract      AuthorizationContract `json:"contract"`
	}{RequestDigest: request.Digest, Candidate: request.Candidate, Artifact: request.Artifact, Contract: request.Contract})
}

func deriveSystemDecision(request AuthorizationRequest) AuthorizationDecisionInput {
	input := AuthorizationDecisionInput{
		Schema: AuthorizationDecisionSchema, Decision: AuthorizationAllow, RequestDigest: request.Digest,
		CandidateID: request.Candidate.CandidateID, CandidateDigest: request.Candidate.CandidateDigest,
		CandidateReportDigest: request.Candidate.CandidateReportDigest, SourceObservationDigest: request.Candidate.SourceObservationDigest,
		InputSourceDigest: request.Candidate.InputSourceDigest, ExecutionInputDigest: request.Candidate.ExecutionInputDigest,
		PolicyDigest: request.Candidate.PolicyDigest, ContractDigest: request.Candidate.ContractCanonicalDigest,
		SubjectSHA: request.Candidate.SubjectSHA, ScopeDigest: request.Candidate.ScopeDigest,
		Repository: request.Artifact.Repository, CandidateArtifactID: request.Artifact.ArtifactID,
		CandidateArtifactDigest: request.Artifact.ArchiveDigest, DecisionRule: AuthorizationEvidenceRule,
		DecisionSource: AuthorizationEvidenceSource, EvidenceDigest: systemEvidenceDigest(request),
	}
	input.DecisionDigest = decisionDigest(input)
	return input
}

func decisionInputValid(request AuthorizationRequest, input AuthorizationDecisionInput) bool {
	return reflect.DeepEqual(input, deriveSystemDecision(request))
}

func candidateBindingEqual(left, right CandidateBinding) bool {
	// CandidateBinding contains the source-backed ExecutionInput pointer. A
	// JSON round-trip necessarily allocates a distinct pointer, so compare the
	// complete value graph instead of pointer addresses. DeepEqual also keeps
	// nil and present inputs distinct, and preserves slice ordering for the
	// canonical source/corpus binding.
	return reflect.DeepEqual(left, right)
}

func buildAuthorization(request AuthorizationRequest, input AuthorizationDecisionInput) *AuthorizationReceipt {
	authorized := input.Decision == AuthorizationAllow
	base := generation.SemanticAdoptionAuthorization{
		Schema:            generation.SemanticAdoptionSystemDecisionSchema,
		AuthorizationID:   "candidate-system-decision/" + strings.TrimPrefix(decisionDigest(input), "sha256:"),
		AuthorizationMode: generation.SemanticAdoptionSystemDecisionMode,
		ProposalDigest:    adoptionDigest(request.Digest), CandidateStableID: adoptionDigest(request.Candidate.CandidateID),
		CandidateInputDigest: adoptionDigest(request.Candidate.InputSourceDigest),
		ContractDigest:       adoptionDigest(request.Candidate.ContractCanonicalDigest),
		InputSourceDigest:    adoptionDigest(request.Candidate.SourceObservationDigest),
		Authorized:           authorized, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	return &AuthorizationReceipt{
		SemanticAdoptionAuthorization: base, Decision: input.Decision,
		CandidateID: request.Candidate.CandidateID, CandidateDigest: request.Candidate.CandidateDigest,
		CandidateReportDigest:   request.Candidate.CandidateReportDigest,
		SourceObservationDigest: request.Candidate.SourceObservationDigest, SubjectSHA: request.Candidate.SubjectSHA,
		ExecutionInputDigest: request.Candidate.ExecutionInputDigest,
		PolicyVersion:        request.Candidate.PolicyVersion, PolicyDigest: request.Candidate.PolicyDigest,
		ContractID: request.Candidate.ContractID, ContractSourceDigest: request.Candidate.ContractSourceDigest,
		ContractCanonicalDigest: request.Candidate.ContractCanonicalDigest, Scope: request.Candidate.Scope,
		ScopeDigest: request.Candidate.ScopeDigest, ExecutionAuthorized: false,
		CandidateArtifactID: request.Artifact.ArtifactID, CandidateArtifactDigest: request.Artifact.ArchiveDigest,
		Repository: input.Repository, DecisionSource: input.DecisionSource, DecisionRule: input.DecisionRule,
		EvidenceDigest: input.EvidenceDigest, DecisionInputDigest: decisionDigest(input),
	}
}

// adoptionDigest converts the candidate bridge's content-addressed form to
// the released generation authorization form, which uses bare SHA-256 hex.
func adoptionDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

// ResolveAuthorization derives candidate eligibility from complete, exact
// system evidence. REFUTED binding contradictions dominate UNKNOWN.
func ResolveAuthorization(request AuthorizationRequest) AuthorizationResolution {
	if !requestCommonValid(request) {
		return refutedResolution(request, AuthorizationRequestReason, nil)
	}
	if incompleteAuthorizationContract(request.Contract) {
		return unknownResolution(request, AuthorizationFieldReason, "RESTORE_AUTHORIZATION_CONTRACT", []string{"authorization_contract"})
	}
	if !validAuthorizationContract(request.Contract) {
		return refutedResolution(request, AuthorizationContractReason, nil)
	}
	if incompleteCandidateBinding(request.Candidate) {
		return unknownResolution(request, AuthorizationFieldReason, "PROVIDE_COMPLETE_CANDIDATE_BINDING", []string{"candidate_adapter_field"})
	}
	if request.Candidate.Authority != (Authority{}) {
		return refutedResolution(request, "CANDIDATE_AUTHORITY_CONTRADICTION", nil)
	}
	if !validCandidateBinding(request.Candidate) {
		return refutedResolution(request, AuthorizationRefutedReason, nil)
	}
	if request.Artifact.Expired || request.Artifact.ArtifactID == 0 || request.Artifact.ArtifactName == "" ||
		request.Artifact.ArchiveDigest == "" || request.Artifact.SizeBytes == 0 {
		return unknownResolution(request, AuthorizationArtifactReason, "RESTORE_CANDIDATE_ARTIFACT", []string{"candidate_artifact"})
	}
	if !validArtifactMetadata(request.Artifact) {
		return refutedResolution(request, AuthorizationRefutedReason, nil)
	}
	if request.Artifact.SubjectSHA != request.Candidate.SubjectSHA {
		return refutedResolution(request, AuthorizationRefutedReason, nil)
	}

	input := deriveSystemDecision(request)
	resolution := resolutionBase(request)
	resolution.Decision = AuthorizationClosed
	resolution.Resolution = ResolutionExact
	resolution.DecisionInputs = []AuthorizationDecisionInput{input}
	resolution.Authorization = buildAuthorization(request, input)
	resolution.Outcome = AuthorizationAuthorized
	resolution.Reason = AuthorizationClosedAllow
	resolution.Metrics.Decisions = 1
	resolution.Metrics.SystemDerivedDecisions = 1
	resolution.LiveAuthorized = 1
	resolution.LiveState = "CLOSED/AUTHORIZED"
	resolution.ExecutionAuthorized = false
	resolution.Digest = resolutionDigest(resolution)
	return resolution
}

func ValidateAuthorizationRequest(request AuthorizationRequest) error {
	if !requestCommonValid(request) {
		return errors.New("authorization request identity mismatch")
	}
	if !validAuthorizationContract(request.Contract) || !validCandidateBinding(request.Candidate) || !validArtifactMetadata(request.Artifact) ||
		request.Artifact.SubjectSHA != request.Candidate.SubjectSHA {
		return errors.New("authorization request binding mismatch")
	}
	if request.Metrics.StructuralUnboundEdgesBefore != 1 || request.Metrics.StructuralUnboundEdgesAfter != 0 ||
		request.Metrics.AdapterEdges != 1 || request.Metrics.Requests != 1 || request.Metrics.Decisions != 0 ||
		request.Metrics.UnknownSixField != 0 || request.Metrics.RefutedContradictions != 0 || request.Metrics.FallbackAccepted != 0 ||
		request.Metrics.RepositoryWrites != 0 || request.Metrics.LocalTestExecutions != 0 {
		return errors.New("authorization request metrics are not exact")
	}
	return nil
}

func ValidateAuthorizationResolution(resolution AuthorizationResolution) error {
	if resolution.Schema != AuthorizationResolutionSchema || resolution.RequestDigest == "" || resolution.Digest != resolutionDigest(resolution) ||
		resolution.RepositoryWrites != 0 || resolution.LocalTestExecutions != 0 || resolution.ExecutionAuthorized || resolution.Metrics.StructuralUnboundEdgesBefore != 1 ||
		resolution.Metrics.StructuralUnboundEdgesAfter != 0 || resolution.Metrics.AdapterEdges != 1 || resolution.Metrics.Requests != 1 ||
		resolution.Metrics.FallbackAccepted != 0 || resolution.Metrics.RepositoryWrites != 0 || resolution.Metrics.LocalTestExecutions != 0 {
		return errors.New("authorization resolution identity or safety mismatch")
	}
	switch resolution.Decision {
	case AuthorizationUnknown:
		if resolution.Resolution != ResolutionLower || resolution.Unknown == nil || resolution.Authorization != nil ||
			len(resolution.DecisionInputs) != 0 || resolution.Outcome != "" || resolution.Metrics.UnknownSixField != 1 ||
			resolution.Metrics.RefutedContradictions != 0 || resolution.Metrics.SystemDerivedDecisions != 0 || resolution.Unknown.Stage != AuthorizationUnknownStage ||
			resolution.Unknown.Step != AuthorizationUnknownStep || resolution.Unknown.UnknownClass != AuthorizationUnknownClass ||
			len(resolution.Unknown.BlockedBy) == 0 {
			return errors.New("authorization UNKNOWN resolution is not causal")
		}
	case AuthorizationRefuted:
		if resolution.Resolution != ResolutionExact || resolution.Unknown != nil || resolution.Authorization != nil ||
			len(resolution.DecisionInputs) != 0 || resolution.Outcome != "" || resolution.Metrics.UnknownSixField != 0 ||
			resolution.Metrics.RefutedContradictions != 1 || resolution.Metrics.SystemDerivedDecisions != 0 {
			return errors.New("authorization REFUTED resolution is not causal")
		}
	case AuthorizationClosed:
		if resolution.Resolution != ResolutionExact || resolution.Unknown != nil || resolution.Authorization == nil ||
			len(resolution.DecisionInputs) != 1 || resolution.Outcome != AuthorizationAuthorized ||
			resolution.LiveAuthorized != 1 || resolution.LiveState != "CLOSED/AUTHORIZED" ||
			resolution.Metrics.Decisions != 1 || resolution.Metrics.SystemDerivedDecisions != 1 ||
			resolution.Metrics.UnknownSixField != 0 || resolution.Metrics.RefutedContradictions != 0 {
			return errors.New("authorization CLOSED resolution is not exact")
		}
	default:
		return errors.New("authorization resolution decision is unknown")
	}
	return nil
}

// VerifyAuthorizationResolution is an independent consumer: it checks the
// receipt's bindings and safety invariants without trusting the evaluator's
// decision alone.
func VerifyAuthorizationResolution(request AuthorizationRequest, resolution AuthorizationResolution) error {
	if err := ValidateAuthorizationRequest(request); err != nil {
		if !requestCommonValid(request) || request.Digest != requestDigest(request) {
			return err
		}
	}
	if err := ValidateAuthorizationResolution(resolution); err != nil {
		return err
	}
	if resolution.RequestDigest != request.Digest || !candidateBindingEqual(resolution.Candidate, request.Candidate) || resolution.Artifact != request.Artifact {
		return errors.New("authorization resolution is not bound to the request")
	}
	if expected := ResolveAuthorization(request); !reflect.DeepEqual(expected, resolution) {
		return errors.New("authorization resolution differs from independent system replay")
	}
	if resolution.Decision == AuthorizationUnknown {
		return nil
	}
	if resolution.Decision == AuthorizationRefuted {
		return nil
	}
	input := deriveSystemDecision(request)
	if !decisionInputValid(request, resolution.DecisionInputs[0]) ||
		!reflect.DeepEqual(resolution.DecisionInputs[0], input) {
		return errors.New("system authorization evidence is not exact")
	}
	auth := resolution.Authorization
	if err := generation.ValidateSemanticAdoptionAuthorization(auth.SemanticAdoptionAuthorization); err != nil {
		return fmt.Errorf("embedded adoption authorization: %w", err)
	}
	if auth.ProposalDigest != adoptionDigest(request.Digest) || auth.CandidateStableID != adoptionDigest(request.Candidate.CandidateID) ||
		auth.CandidateInputDigest != adoptionDigest(request.Candidate.InputSourceDigest) || auth.ContractDigest != adoptionDigest(request.Candidate.ContractCanonicalDigest) ||
		auth.InputSourceDigest != adoptionDigest(request.Candidate.SourceObservationDigest) || auth.CandidateID != request.Candidate.CandidateID ||
		auth.CandidateDigest != request.Candidate.CandidateDigest || auth.CandidateReportDigest != request.Candidate.CandidateReportDigest ||
		auth.ExecutionInputDigest != request.Candidate.ExecutionInputDigest ||
		auth.SourceObservationDigest != request.Candidate.SourceObservationDigest || auth.SubjectSHA != request.Candidate.SubjectSHA ||
		auth.PolicyDigest != request.Candidate.PolicyDigest || auth.ContractID != request.Candidate.ContractID ||
		auth.ContractSourceDigest != request.Candidate.ContractSourceDigest ||
		auth.ContractCanonicalDigest != request.Candidate.ContractCanonicalDigest || auth.Scope != request.Candidate.Scope ||
		auth.ScopeDigest != request.Candidate.ScopeDigest || auth.CandidateArtifactID != request.Artifact.ArtifactID ||
		auth.CandidateArtifactDigest != request.Artifact.ArchiveDigest || auth.ExecutionAuthorized || auth.RepositoryWrites != 0 ||
		auth.LocalTestExecutions != 0 || !auth.Authorized ||
		auth.Schema != generation.SemanticAdoptionSystemDecisionSchema || auth.AuthorizationMode != generation.SemanticAdoptionSystemDecisionMode ||
		auth.Decision != AuthorizationAllow || auth.Repository != request.Artifact.Repository ||
		auth.DecisionSource != input.DecisionSource || auth.DecisionRule != input.DecisionRule || auth.EvidenceDigest != input.EvidenceDigest ||
		resolution.LiveAuthorized != 1 || resolution.LiveState != "CLOSED/AUTHORIZED" || resolution.ExecutionAuthorized ||
		resolution.Decision != AuthorizationClosed ||
		resolution.Outcome != AuthorizationAuthorized || resolution.Reason != AuthorizationClosedAllow {
		return errors.New("authorization receipt binding or safety invariant mismatch")
	}
	if auth.DecisionInputDigest != decisionDigest(input) {
		return errors.New("authorization decision digest mismatch")
	}
	return nil
}

func BuildAuthorizationVerification(request AuthorizationRequest, resolution AuthorizationResolution) (AuthorizationVerification, error) {
	if err := VerifyAuthorizationResolution(request, resolution); err != nil {
		return AuthorizationVerification{}, err
	}
	verification := AuthorizationVerification{Schema: "gooo/self-improvement-candidate-authorization-verification/v2",
		RequestDigest: request.Digest, ResolutionDigest: resolution.Digest, Decision: resolution.Decision,
		Resolution: resolution.Resolution, DecisionVerified: true, ExecutionAuthorized: resolution.ExecutionAuthorized, RepositoryWrites: resolution.RepositoryWrites,
		LocalTestExecutions: resolution.LocalTestExecutions, LiveAuthorized: resolution.LiveAuthorized, LiveState: resolution.LiveState}
	verification.Digest = verificationDigest(verification)
	return verification, nil
}

func resolutionSummary(resolution AuthorizationResolution, id string) CanonicalCase {
	return CanonicalCase{ID: id, ExpectedDecision: resolution.Decision, ExpectedReason: resolution.Reason,
		ActualDecision: resolution.Decision, ActualReason: resolution.Reason, Unknown: resolution.Unknown, Pass: true}
}

func canonicalCase(id string, request AuthorizationRequest, expectedDecision, expectedReason string) (CanonicalCase, error) {
	resolution := ResolveAuthorization(request)
	if resolution.Decision != expectedDecision || resolution.Reason != expectedReason {
		return CanonicalCase{}, fmt.Errorf("canonical case %s resolved %s/%s, expected %s/%s", id, resolution.Decision, resolution.Reason, expectedDecision, expectedReason)
	}
	if err := VerifyAuthorizationResolution(request, resolution); err != nil {
		return CanonicalCase{}, fmt.Errorf("canonical case %s independent verification: %w", id, err)
	}
	result := resolutionSummary(resolution, id)
	result.ExpectedDecision, result.ExpectedReason = expectedDecision, expectedReason
	return result, nil
}

type authorizationCaseSpec struct {
	id, decision, reason string
	request              AuthorizationRequest
}

func cloneAuthorizationRequest(request AuthorizationRequest) AuthorizationRequest {
	raw, _ := json.Marshal(request)
	var clone AuthorizationRequest
	_ = json.Unmarshal(raw, &clone)
	return clone
}

func canonicalAuthorizationCaseSpecs(request AuthorizationRequest) []authorizationCaseSpec {
	missingField := request
	missingField.Candidate.CandidateDigest = ""
	missingField.Digest = requestDigest(missingField)
	expired := request
	expired.Artifact.Expired = true
	expired.Digest = requestDigest(expired)
	policyMissing := request
	policyMissing.Candidate.PolicyDigest = ""
	policyMissing.Digest = requestDigest(policyMissing)
	candidateMismatch := cloneAuthorizationRequest(request)
	candidateMismatch.Candidate.ExecutionInput.CandidateDigest = digestBytes([]byte("wrong-candidate"))
	candidateMismatch.Digest = requestDigest(candidateMismatch)
	observationMismatch := cloneAuthorizationRequest(request)
	observationMismatch.Candidate.ExecutionInput.ObservationDigest = digestBytes([]byte("wrong-observation"))
	observationMismatch.Digest = requestDigest(observationMismatch)
	authorityMismatch := request
	authorityMismatch.Candidate.Authority.RepositoryWrites = 1
	authorityMismatch.Digest = requestDigest(authorityMismatch)
	return []authorizationCaseSpec{
		{"exact-system-evidence", AuthorizationClosed, AuthorizationClosedAllow, request},
		{"deterministic-replay", AuthorizationClosed, AuthorizationClosedAllow, request},
		{"independent-system-decision", AuthorizationClosed, AuthorizationClosedAllow, request},
		{"missing-adapter-field", AuthorizationUnknown, AuthorizationFieldReason, missingField},
		{"expired-or-missing-artifact", AuthorizationUnknown, AuthorizationArtifactReason, expired},
		{"missing-policy-evidence", AuthorizationUnknown, AuthorizationFieldReason, policyMissing},
		{"candidate-digest-mismatch", AuthorizationRefuted, AuthorizationRefutedReason, candidateMismatch},
		{"observation-contract-mismatch", AuthorizationRefuted, AuthorizationRefutedReason, observationMismatch},
		{"authority-contradiction", AuthorizationRefuted, "CANDIDATE_AUTHORITY_CONTRADICTION", authorityMismatch},
	}
}

func newCanonicalCaseReport(request AuthorizationRequest) CanonicalCaseReport {
	return CanonicalCaseReport{Schema: AuthorizationCasesSchema, RequestDigest: request.Digest,
		CaseDenominator: 9, Counts: map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0},
		LiveAuthorized: 0, LiveState: "UNKNOWN", Decision: AuthorizationClosed,
		Resolution: ResolutionExact, Reason: "EXACT_CANONICAL_AUTHORIZATION_CASES",
		Roundtrip: AuthorizationRoundtripEvidence{
			AuthorizationRoundtripExactBefore: 0, AuthorizationRoundtripExactAfter: 1,
			PointerIdentityDependencyBefore: 1, PointerIdentityDependencyAfter: 0,
			CounterexampleRunID: 33926584593,
		}}
}

func BuildCanonicalCases(request AuthorizationRequest) (CanonicalCaseReport, error) {
	if err := ValidateAuthorizationRequest(request); err != nil {
		return CanonicalCaseReport{}, err
	}
	caseSpecs := canonicalAuthorizationCaseSpecs(request)
	report := newCanonicalCaseReport(request)
	for _, spec := range caseSpecs {
		result, err := canonicalCase(spec.id, spec.request, spec.decision, spec.reason)
		if err != nil {
			return CanonicalCaseReport{}, err
		}
		report.Cases = append(report.Cases, result)
		report.Counts[result.ActualDecision]++
	}
	first := ResolveAuthorization(request)
	second := ResolveAuthorization(request)
	firstBytes, _ := json.Marshal(first)
	secondBytes, _ := json.Marshal(second)
	report.ReplayEqual = bytes.Equal(firstBytes, secondBytes)
	if !report.ReplayEqual {
		return CanonicalCaseReport{}, errors.New("authorization replay is not deterministic")
	}
	report.ClosedCases = report.Counts[AuthorizationClosed]
	report.UnknownCases = report.Counts[AuthorizationUnknown]
	report.RefutedCases = report.Counts[AuthorizationRefuted]
	report.Metrics = AuthorizationMetrics{StructuralUnboundEdgesBefore: 1, StructuralUnboundEdgesAfter: 0,
		AdapterEdges: 1, Requests: 9, Decisions: 3, SystemDerivedDecisions: 3, UnknownSixField: 3, RefutedContradictions: 3,
		FallbackAccepted: 0, IndependentReplayComparisons: 1, ArtifactFiles: 9, ArtifactTypes: 3,
		RepositoryWrites: 0, LocalTestExecutions: 0}
	report.Digest = canonicalCasesDigest(report)
	return report, nil
}

func canonicalCasesDigest(report CanonicalCaseReport) string {
	report.Digest = ""
	return digestJSON(report)
}

func ValidateCanonicalCases(report CanonicalCaseReport) error {
	if report.Schema != AuthorizationCasesSchema || report.CaseDenominator != 9 || report.ClosedCases != 3 ||
		report.UnknownCases != 3 || report.RefutedCases != 3 || report.Counts[AuthorizationClosed] != 3 ||
		report.Counts[AuthorizationUnknown] != 3 || report.Counts[AuthorizationRefuted] != 3 || !report.ReplayEqual ||
		report.LiveAuthorized != 0 || report.LiveState != "UNKNOWN" || report.Decision != AuthorizationClosed ||
		report.Resolution != ResolutionExact || report.Roundtrip.AuthorizationRoundtripExactBefore != 0 ||
		report.Roundtrip.AuthorizationRoundtripExactAfter != 1 || report.Roundtrip.PointerIdentityDependencyBefore != 1 ||
		report.Roundtrip.PointerIdentityDependencyAfter != 0 || report.Roundtrip.CounterexampleRunID != 33926584593 ||
		report.Digest != canonicalCasesDigest(report) {
		return errors.New("canonical authorization cases are not exact")
	}
	if report.Metrics.Decisions != 3 || report.Metrics.SystemDerivedDecisions != 3 ||
		report.Metrics.UnknownSixField != 3 || report.Metrics.RefutedContradictions != 3 ||
		report.Metrics.FallbackAccepted != 0 || report.Metrics.RepositoryWrites != 0 || report.Metrics.LocalTestExecutions != 0 {
		return errors.New("canonical system decision metrics are not exact")
	}
	if len(report.Cases) != report.CaseDenominator {
		return errors.New("canonical authorization case denominator mismatch")
	}
	for _, item := range report.Cases {
		if !item.Pass || item.ExpectedDecision != item.ActualDecision || item.ExpectedReason != item.ActualReason {
			return errors.New("canonical authorization case failed")
		}
		if item.ActualDecision == AuthorizationUnknown && item.Unknown == nil {
			return errors.New("canonical UNKNOWN case lacks six-field evidence")
		}
		if item.ActualDecision != AuthorizationUnknown && item.Unknown != nil {
			return errors.New("canonical non-UNKNOWN case contains unknown evidence")
		}
	}
	return nil
}
