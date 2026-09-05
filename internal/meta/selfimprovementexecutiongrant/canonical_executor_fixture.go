package selfimprovementexecutiongrant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	valuewitnessinput "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

const (
	CanonicalExecutorRequestSchema          = "gooo/self-improvement-execution-grant-canonical-request/v1"
	CanonicalExecutorDecisionSchema         = "gooo/self-improvement-execution-grant-canonical-decision/v1"
	CanonicalExecutorReceiptSchema          = "gooo/self-improvement-execution-grant-canonical-receipt/v1"
	CanonicalExecutorVerificationSchema     = "gooo/self-improvement-execution-grant-canonical-verification/v1"
	CanonicalExecutorBindingManifestSchema  = "gooo/self-improvement-execution-grant-canonical-binding-manifest/v1"
	CanonicalExecutorDecisionType           = "CANONICAL_EXECUTOR_CONFORMANCE_ONLY"
	CanonicalExecutorScope                  = "CALLER_OWNED_TEMP_ONLY"
	CanonicalExecutorMaterializedCase       = "explicit-allow"
	CanonicalExecutorUnknownReason          = "MISSING_EXACT_CANONICAL_EXECUTOR_GRANT_INPUT"
	CanonicalExecutorUnknownNext            = "restore-exact-canonical-executor-grant-input"
	CanonicalExecutorUnknownBlockedBy       = "canonical_executor_grant_input"
	CanonicalExecutorWorkflowName           = "Self-improvement candidate execution contract"
	CanonicalExecutorWorkflowPath           = ".github/workflows/self-improvement-execution-contract.yml"
	CanonicalExecutorWorkflowEvent          = "workflow_dispatch"
	CanonicalExecutorWorkflowRef            = "refs/heads/dev"
	CanonicalExecutorArtifactPrefix         = "self-improvement-execution-contract-"
	CanonicalExecutorSourceSchema           = "gooo/self-improvement-execution-grant-source-artifact/v1"
)

// CanonicalExecutorBindingNames is the complete identity surface that must be
// bound before a canonical-only grant may be materialized.
var canonicalExecutorBindingNames = []string{
	"v24_request_digest", "v24_resolution_digest", "v24_verification_digest",
	"candidate_stable_id", "candidate_digest", "subject_sha", "observation_digest",
	"candidate_input_digest", "v24_contract_digest", "v25_contract_digest",
	"v24_authorization_contract_digest",
	"v25_verification_digest", "operation_id", "evaluator_registry_digest",
	"toolchain_test_contract_identity", "max_executions", "repository_writes_allowed",
	"source_workflow_name", "source_workflow_path", "source_event", "source_ref",
	"source_head_sha", "source_workflow_run_id", "source_workflow_run_attempt",
	"source_artifact_name", "source_artifact_id", "source_artifact_archive_digest",
	"source_artifact_observed_digest", "source_artifact_expiry",
	"executor_semantic_contract_digest", "materialize_activity_digest", "verify_activity_digest",
}

// CanonicalExecutorBindingNames returns a copy so callers cannot alter the
// typed binding contract used by the manifest and independent verifier.
func CanonicalExecutorBindingNames() []string {
	return append([]string(nil), canonicalExecutorBindingNames...)
}

func CanonicalExecutorArtifactNames() []string {
	return []string{
		"canonical-executor-grant-request.json",
		"canonical-executor-grant-decision.json",
		"canonical-executor-grant-receipt.json",
		"canonical-executor-grant-verification.json",
		"canonical-executor-grant-binding-manifest.json",
	}
}

type CanonicalExecutorSourceArtifact struct {
	Schema                    string `json:"schema"`
	WorkflowName              string `json:"workflow_name"`
	WorkflowPath              string `json:"workflow_path"`
	Repository                string `json:"repository"`
	Event                     string `json:"event"`
	Ref                       string `json:"ref"`
	HeadSHA                   string `json:"head_sha"`
	WorkflowRunID             int64  `json:"workflow_run_id"`
	WorkflowRunAttempt        int    `json:"workflow_run_attempt"`
	ArtifactName              string `json:"artifact_name"`
	ArtifactID                int64  `json:"artifact_id"`
	APIArchiveDigest          string `json:"api_archive_digest"`
	ObservedArchiveDigest     string `json:"observed_archive_digest"`
	EmbeddedVerificationDigest string `json:"embedded_verification_digest"`
	ArtifactExpired           bool   `json:"artifact_expired"`
	ArtifactExpiryKnown       bool   `json:"artifact_expiry_known"`
	ArtifactRetrieved         bool   `json:"artifact_retrieved"`
	ArtifactRetrievalError    string `json:"artifact_retrieval_error,omitempty"`
}

func (source CanonicalExecutorSourceArtifact) ToSourceArtifact() SourceArtifact {
	return SourceArtifact{
		Repository: source.Repository, WorkflowRunID: source.WorkflowRunID,
		WorkflowRunAttempt: source.WorkflowRunAttempt, ArtifactID: source.ArtifactID,
		ArtifactDigest: source.APIArchiveDigest, ObservedArtifactDigest: source.ObservedArchiveDigest,
		ArtifactExpired: source.ArtifactExpired, ArtifactExpiryKnown: source.ArtifactExpiryKnown,
		ArtifactRetrieved: source.ArtifactRetrieved, ArtifactRetrievalError: source.ArtifactRetrievalError,
	}
}

type CanonicalExecutorGrantRequest struct {
	Schema          string                                       `json:"schema"`
	GrantRequest    GrantRequest                                 `json:"grant_request"`
	V24Request      selfimprovementcandidate.AuthorizationRequest `json:"v24_request"`
	V24Resolution   selfimprovementcandidate.AuthorizationResolution `json:"v24_resolution"`
	V24Verification selfimprovementcandidate.AuthorizationVerification `json:"v24_verification"`
	V25Contract     v25.ContractResolution                        `json:"v25_contract"`
	V25Verification v25.Verification                              `json:"v25_verification"`
	CandidateInput  valuewitnessinput.ExecutionInput               `json:"candidate_input"`
	SourceArtifact  CanonicalExecutorSourceArtifact                `json:"source_artifact"`
	Digest          string                                         `json:"digest"`
}

type CanonicalExecutorGrantDecision struct {
	Schema                   string            `json:"schema"`
	DecisionType             string            `json:"decision_type"`
	DecisionInput            GrantDecisionInput `json:"decision_input"`
	Fixture                  bool              `json:"fixture"`
	LiveAuthority            bool              `json:"live_authority"`
	UserDecision             bool              `json:"user_decision"`
	ProductUtilityEvidence   bool              `json:"product_utility_evidence"`
	CanonicalExecutionAllowed bool            `json:"canonical_execution_allowed"`
	ExecutionScope           string            `json:"execution_scope"`
	Digest                   string            `json:"digest"`
}

type CanonicalExecutorGrantReceipt struct {
	Schema                   string                      `json:"schema"`
	GrantReceipt             GrantReceipt                 `json:"grant_receipt"`
	DecisionType             string                      `json:"decision_type"`
	Fixture                  bool                        `json:"fixture"`
	LiveAuthority            bool                        `json:"live_authority"`
	CanonicalExecutionAllowed bool                      `json:"canonical_execution_allowed"`
	ExecutionScope           string                      `json:"execution_scope"`
	MaxExecutions            int                         `json:"max_executions"`
	RequestDigest            string                      `json:"request_digest"`
	DecisionDigest           string                      `json:"decision_digest"`
	CandidateStableID        string                      `json:"candidate_stable_id"`
	CandidateDigest          string                      `json:"candidate_digest"`
	CandidateInputDigest     string                      `json:"candidate_input_digest"`
	V24RequestDigest         string                      `json:"v24_request_digest"`
	V24ResolutionDigest      string                      `json:"v24_resolution_digest"`
	V24VerificationDigest    string                      `json:"v24_verification_digest"`
	V24AuthorizationContractDigest string                 `json:"v24_authorization_contract_digest"`
	V25ContractDigest        string                      `json:"v25_contract_digest"`
	V25VerificationDigest    string                      `json:"v25_verification_digest"`
	SubjectSHA               string                      `json:"subject_sha"`
	SourceArtifact           CanonicalExecutorSourceArtifact `json:"source_artifact"`
	Digest                   string                      `json:"digest"`
}

type CanonicalExecutorVerificationCase struct {
	ID                 string     `json:"id"`
	ExpectedDecision   Decision   `json:"expected_decision"`
	ExpectedResolution Resolution `json:"expected_resolution"`
	ExpectedReason     string     `json:"expected_reason"`
	ActualDecision     Decision   `json:"actual_decision"`
	ActualResolution   Resolution `json:"actual_resolution"`
	ActualReason       string     `json:"actual_reason"`
	MissingFields      []string   `json:"missing_fields,omitempty"`
	ContradictoryFields []string  `json:"contradictory_fields,omitempty"`
	Pass               bool       `json:"pass"`
}

type CanonicalExecutorVerification struct {
	Schema                       string                              `json:"schema"`
	RequestDigest                string                              `json:"request_digest"`
	DecisionDigest               string                              `json:"decision_digest"`
	ReceiptDigest                string                              `json:"receipt_digest"`
	IndependentDecision          Decision                            `json:"independent_decision"`
	IndependentResolution        Resolution                          `json:"independent_resolution"`
	IndependentReason            string                              `json:"independent_reason"`
	Verified                     bool                                `json:"verified"`
	IndependentReplayComparisons int                                 `json:"independent_replay_comparisons"`
	CaseDenominator              int                                 `json:"case_denominator"`
	Counts                      map[string]int                      `json:"counts"`
	Cases                       []CanonicalExecutorVerificationCase `json:"cases"`
	Fixture                     bool                                `json:"fixture"`
	LiveAuthority               bool                                `json:"live_authority"`
	CanonicalExecutionCount     int                                 `json:"canonical_execution_count"`
	GrantConsumedUses           int                                 `json:"grant_consumed_uses"`
	RepositoryWrites            int                                 `json:"repository_writes"`
	LocalTestExecutions         int                                 `json:"local_test_executions"`
	FallbackAccepted            int                                 `json:"fallback_accepted"`
	PerformanceImprovement     string                              `json:"performance_improvement"`
	Digest                      string                              `json:"digest"`
}

type CanonicalExecutorBindingManifest struct {
	Schema                                  string                         `json:"schema"`
	FixtureSchema                           string                         `json:"fixture_schema"`
	MaterializedFixtureOf                   string                         `json:"materialized_fixture_of"`
	MaterializedFixtureCountBefore          int                            `json:"materialized_fixture_count_before"`
	MaterializedFixtureCountAfter           int                            `json:"materialized_fixture_count_after"`
	ExactGrantIdentityBindingsBefore       int                            `json:"exact_grant_identity_bindings_before"`
	ExactGrantIdentityBindingsAfter        int                            `json:"exact_grant_identity_bindings_after"`
	ExactGrantIdentityBindings              int                            `json:"exact_grant_identity_bindings"`
	ExactGrantIdentityBindingDenominator    int                            `json:"exact_grant_identity_binding_denominator"`
	PlaceholderFieldsBefore                 int                            `json:"placeholder_fields_before"`
	PlaceholderFieldsAfter                  int                            `json:"placeholder_fields_after"`
	ArtifactFiles                           int                            `json:"artifact_files"`
	ArtifactTypes                           int                            `json:"artifact_types"`
	ArtifactNames                           []string                       `json:"artifact_names"`
	RequestDigest                           string                         `json:"request_digest"`
	DecisionDigest                          string                         `json:"decision_digest"`
	ReceiptDigest                           string                         `json:"receipt_digest"`
	VerificationDigest                      string                         `json:"verification_digest"`
	CandidateStableID                       string                         `json:"candidate_stable_id"`
	CandidateDigest                         string                         `json:"candidate_digest"`
	CandidateInputDigest                    string                         `json:"candidate_input_digest"`
	SubjectSHA                              string                         `json:"subject_sha"`
	V24RequestDigest                        string                         `json:"v24_request_digest"`
	V24ResolutionDigest                     string                         `json:"v24_resolution_digest"`
	V24VerificationDigest                   string                         `json:"v24_verification_digest"`
	V24AuthorizationContractDigest           string                         `json:"v24_authorization_contract_digest"`
	V25ContractDigest                       string                         `json:"v25_contract_digest"`
	V25VerificationDigest                   string                         `json:"v25_verification_digest"`
	SourceArtifact                          CanonicalExecutorSourceArtifact `json:"source_artifact"`
	DecisionType                            string                         `json:"decision_type"`
	Fixture                                 bool                           `json:"fixture"`
	LiveAuthority                           bool                           `json:"live_authority"`
	UserDecision                            bool                           `json:"user_decision"`
	ProductUtilityEvidence                  bool                           `json:"product_utility_evidence"`
	CanonicalExecutionCount                 int                            `json:"canonical_execution_count"`
	ReceiptRemainingUses                   int                            `json:"receipt_remaining_uses"`
	ReceiptConsumedUses                    int                            `json:"receipt_consumed_uses"`
	ReceiptExecutionCount                  int                            `json:"receipt_execution_count"`
	RepositoryWrites                       int                            `json:"repository_writes"`
	LocalTestExecutions                     int                            `json:"local_test_executions"`
	ManualDecisions                         int                            `json:"manual_decisions"`
	IndependentReplayComparisons            int                            `json:"independent_replay_comparisons"`
	FallbackAccepted                        int                            `json:"fallback_accepted"`
	PerformanceImprovement                  string                         `json:"performance_improvement"`
	GoPhysicalLines                         int                            `json:"go_physical_lines"`
	GoooPhysicalLines                       int                            `json:"gooo_physical_lines"`
	ExecutorContractDigest                  string                         `json:"executor_contract_digest"`
	MaterializeActivityDigest               string                         `json:"materialize_activity_digest"`
	VerifyActivityDigest                    string                         `json:"verify_activity_digest"`
	Decision                                Decision                       `json:"decision"`
	Resolution                              Resolution                     `json:"resolution"`
	Reason                                  string                         `json:"reason"`
	Digest                                  string                         `json:"digest"`
}

type CanonicalExecutorGrantFixture struct {
	Request      CanonicalExecutorGrantRequest
	Decision     CanonicalExecutorGrantDecision
	Receipt      CanonicalExecutorGrantReceipt
	Verification CanonicalExecutorVerification
	Manifest     CanonicalExecutorBindingManifest
}

func BuildCanonicalExecutorGrantFixture(program PolicyProgram, contractProgram v25.PolicyProgram, v24Request selfimprovementcandidate.AuthorizationRequest, v24Resolution selfimprovementcandidate.AuthorizationResolution, v24Verification selfimprovementcandidate.AuthorizationVerification, v25Report v25.LiveReport, source CanonicalExecutorSourceArtifact) (CanonicalExecutorGrantFixture, error) {
	materializeProgram, _, err := canonicalExecutorProgramBinding(program)
	if err != nil {
		return CanonicalExecutorGrantFixture{}, fmt.Errorf("canonical executor semantic activity contract: %w", err)
	}
	if err := validateCanonicalExecutorRoots(contractProgram, v24Request, v24Resolution, v24Verification, v25Report, source); err != nil {
		return CanonicalExecutorGrantFixture{}, err
	}
	if source.HeadSHA != v25Report.ContractResolution.SubjectSHA {
		return CanonicalExecutorGrantFixture{}, errors.New("canonical executor source subject differs from v25 subject")
	}
	v24Binding := ProjectV24(v24Request, v24Resolution)
	v25Binding := ProjectV25(v25Report.ContractResolution)
	grantRequest := BuildRequest(program, v24Binding, v25Binding, source.ToSourceArtifact())
	request := CanonicalExecutorGrantRequest{
		Schema: materializeProgram["fixture_schema"], GrantRequest: grantRequest,
		V24Request: v24Request, V24Resolution: v24Resolution, V24Verification: v24Verification,
		V25Contract: v25Report.ContractResolution, V25Verification: v25Report.Verification,
		CandidateInput: *v25Report.ContractResolution.ExecutionInput, SourceArtifact: source,
	}
	request.Digest = canonicalExecutorRequestDigest(request)
	decisionInput := BuildDecisionInput(grantRequest, DecisionAllow, DecisionSourceCanonical, ActorEvidence{
		Repository: source.Repository, Actor: CanonicalExecutorDecisionType,
		WorkflowRunID: source.WorkflowRunID, WorkflowRunAttempt: source.WorkflowRunAttempt,
		Event: "canonical-fixture", EvidenceLabel: CanonicalEvidenceLabel,
	})
	decision := CanonicalExecutorGrantDecision{
		Schema: CanonicalExecutorDecisionSchema, DecisionType: materializeProgram["decision_type"],
		DecisionInput: decisionInput, Fixture: true, LiveAuthority: materializeProgram["live_authority"] == "true", UserDecision: materializeProgram["user_decision"] == "true",
		ProductUtilityEvidence: materializeProgram["product_utility_evidence"] == "true", CanonicalExecutionAllowed: true, ExecutionScope: materializeProgram["scope"],
	}
	decision.Digest = canonicalExecutorDecisionDigest(decision)
	grantReceipt := buildReceipt(grantRequest.Digest, decisionInput, true)
	if grantReceipt == nil || ValidateGrantReceipt(*grantReceipt) != nil {
		return CanonicalExecutorGrantFixture{}, errors.New("canonical executor grant receipt did not validate")
	}
	receipt := CanonicalExecutorGrantReceipt{
		Schema: CanonicalExecutorReceiptSchema, GrantReceipt: *grantReceipt,
		DecisionType: decision.DecisionType, Fixture: true, LiveAuthority: decision.LiveAuthority,
		CanonicalExecutionAllowed: decision.CanonicalExecutionAllowed, ExecutionScope: decision.ExecutionScope, MaxExecutions: MaxExecutions,
		RequestDigest: grantRequest.Digest, DecisionDigest: decision.Digest,
		CandidateStableID: request.CandidateInput.CandidateStableID, CandidateDigest: request.CandidateInput.CandidateDigest,
		CandidateInputDigest: request.CandidateInput.Digest, V24RequestDigest: v24Request.Digest,
		V24ResolutionDigest: v24Resolution.Digest, V24VerificationDigest: v24Verification.Digest,
		V24AuthorizationContractDigest: v24Request.Contract.CanonicalDigest,
		V25ContractDigest: v25Report.ContractResolution.Digest, V25VerificationDigest: v25Report.Verification.Digest,
		SubjectSHA: source.HeadSHA, SourceArtifact: source,
	}
	receipt.Digest = canonicalExecutorReceiptDigest(receipt)
	fixture := CanonicalExecutorGrantFixture{Request: request, Decision: decision, Receipt: receipt}
	fixture.Verification = VerifyCanonicalExecutorGrantFixture(program, fixture)
	fixture.Manifest = buildCanonicalExecutorBindingManifest(program, fixture)
	fixture.Manifest.Digest = canonicalExecutorManifestDigest(fixture.Manifest)
	if !fixture.Verification.Verified {
		return CanonicalExecutorGrantFixture{}, errors.New("canonical executor grant fixture independent verification failed")
	}
	return fixture, nil
}

func validateCanonicalExecutorRoots(contractProgram v25.PolicyProgram, v24Request selfimprovementcandidate.AuthorizationRequest, v24Resolution selfimprovementcandidate.AuthorizationResolution, v24Verification selfimprovementcandidate.AuthorizationVerification, v25Report v25.LiveReport, source CanonicalExecutorSourceArtifact) error {
	if err := selfimprovementcandidate.VerifyAuthorizationResolution(v24Request, v24Resolution); err != nil {
		return fmt.Errorf("v24 authorization root is not exact: %w", err)
	}
	expectedV24Verification, err := selfimprovementcandidate.BuildAuthorizationVerification(v24Request, v24Resolution)
	if err != nil || !reflect.DeepEqual(expectedV24Verification, v24Verification) {
		return errors.New("v24 authorization verification root is not exact")
	}
	if err := v25.VerifyResolution(v25Report.ContractResolution); err != nil {
		return fmt.Errorf("v25 contract root is not exact: %w", err)
	}
	if v25Report.Verification.ContractDigest != v25Report.ContractResolution.Digest || !v25Report.Verification.Verified || v25Report.Verification.IndependentReplayComparisons != 1 || v25Report.Verification.RepositoryWrites != 0 || v25Report.Verification.LocalTestExecutions != 0 || v25Report.Verification.ExecutionGrants != 0 {
		return errors.New("v25 contract verification root is not exact")
	}
	input := v25.ProjectAuthorizationRequest(v24Request, v25.KnownRegistry())
	expectedV25Verification := v25.Verify(contractProgram, input, v25Report.ContractResolution)
	if !reflect.DeepEqual(expectedV25Verification, v25Report.Verification) {
		return errors.New("v25 independent verification root is not exact")
	}
	if v25Report.ContractResolution.ExecutionInput == nil || v24Request.Candidate.ExecutionInput == nil || !reflect.DeepEqual(*v25Report.ContractResolution.ExecutionInput, *v24Request.Candidate.ExecutionInput) {
		return errors.New("candidate execution input is not identically propagated from v24 to v25")
	}
	if err := valuewitnessinput.Validate(*v25Report.ContractResolution.ExecutionInput); err != nil {
		return fmt.Errorf("candidate execution input is not exact: %w", err)
	}
	if err := validateCanonicalExecutorSource(source); err != nil {
		return err
	}
	if source.HeadSHA != v24Request.Candidate.SubjectSHA || source.HeadSHA != v25Report.ContractResolution.SubjectSHA {
		return errors.New("canonical executor source subject is not bound to candidate and v25")
	}
	return nil
}

func validateCanonicalExecutorSource(source CanonicalExecutorSourceArtifact) error {
	if source.Schema != CanonicalExecutorSourceSchema || source.WorkflowName != CanonicalExecutorWorkflowName || source.WorkflowPath != CanonicalExecutorWorkflowPath || source.Repository == "" || source.Event != CanonicalExecutorWorkflowEvent || source.Ref != CanonicalExecutorWorkflowRef || !validSHA(source.HeadSHA) || source.WorkflowRunID <= 0 || source.WorkflowRunAttempt <= 0 || source.ArtifactName != CanonicalExecutorArtifactPrefix+source.HeadSHA || source.ArtifactID <= 0 || !validDigest(source.APIArchiveDigest) || !validDigest(source.ObservedArchiveDigest) || source.APIArchiveDigest != source.ObservedArchiveDigest || !validDigest(source.EmbeddedVerificationDigest) || source.ArtifactExpired || !source.ArtifactExpiryKnown || !source.ArtifactRetrieved || source.ArtifactRetrievalError != "" {
		return errors.New("canonical executor source artifact identity is not exact")
	}
	return nil
}

func VerifyCanonicalExecutorGrantFixture(program PolicyProgram, fixture CanonicalExecutorGrantFixture) CanonicalExecutorVerification {
	_, verifyProgram, contractErr := canonicalExecutorProgramBinding(program)
	decision, resolution, reason, missing, contradictory := classifyCanonicalExecutorFixture(fixture)
	verification := CanonicalExecutorVerification{
		Schema: CanonicalExecutorVerificationSchema, RequestDigest: fixture.Request.GrantRequest.Digest,
		DecisionDigest: fixture.Decision.Digest, ReceiptDigest: fixture.Receipt.Digest,
		IndependentDecision: decision, IndependentResolution: resolution, IndependentReason: reason,
		CaseDenominator: 0,
		Counts: map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}, Fixture: true,
		LiveAuthority: false, CanonicalExecutionCount: 0, GrantConsumedUses: 0,
		RepositoryWrites: 0, LocalTestExecutions: 0, FallbackAccepted: 0,
		PerformanceImprovement: PerformanceUnknown,
	}
	verification.Cases = canonicalExecutorVerificationCases(program, fixture)
	verification.CaseDenominator = len(verification.Cases)
	firstDecision, firstResolution, firstReason, firstMissing, firstContradictory := classifyCanonicalExecutorFixture(fixture)
	secondDecision, secondResolution, secondReason, secondMissing, secondContradictory := classifyCanonicalExecutorFixture(fixture)
	replayEqual := firstDecision == secondDecision && firstResolution == secondResolution && firstReason == secondReason && reflect.DeepEqual(firstMissing, secondMissing) && reflect.DeepEqual(firstContradictory, secondContradictory)
	if replayEqual {
		verification.IndependentReplayComparisons = 1
	}
	for _, item := range verification.Cases {
		verification.Counts[string(item.ActualDecision)]++
	}
	verification.Verified = contractErr == nil && verifyProgram["verification"] == "independent" && verifyProgram["candidate_execution"] == "0" && verifyProgram["grant_consumption"] == "0" && verifyProgram["repository_writes"] == "0" && verifyProgram["local_test_executions"] == "0" && verifyProgram["refuted_dominates_unknown"] == "true" && decision == DecisionClosed && resolution == ResolutionGrantedUnconsumed && reason == ReasonAllow && len(missing) == 0 && len(contradictory) == 0 && fixture.Decision.DecisionType == CanonicalExecutorDecisionType && fixture.Decision.Fixture && !fixture.Decision.LiveAuthority && !fixture.Decision.UserDecision && !fixture.Decision.ProductUtilityEvidence && fixture.Decision.CanonicalExecutionAllowed && fixture.Decision.ExecutionScope == CanonicalExecutorScope && fixture.Receipt.GrantReceipt.GrantAllowsExecution && fixture.Receipt.GrantReceipt.RemainingUses == 1 && fixture.Receipt.GrantReceipt.ConsumedUses == 0 && fixture.Receipt.GrantReceipt.ExecutionCount == 0 && fixture.Receipt.MaxExecutions == 1 && !fixture.Receipt.LiveAuthority && fixture.Receipt.CanonicalExecutionAllowed && fixture.Receipt.GrantReceipt.Digest == receiptDigest(fixture.Receipt.GrantReceipt) && fixture.Decision.Digest == canonicalExecutorDecisionDigest(fixture.Decision) && fixture.Receipt.Digest == canonicalExecutorReceiptDigest(fixture.Receipt) && fixture.Request.Digest == canonicalExecutorRequestDigest(fixture.Request) && verification.IndependentReplayComparisons == 1 && verification.Counts["CLOSED"] == 1 && verification.Counts["UNKNOWN"] == 3 && verification.Counts["REFUTED"] == 9 && allCanonicalExecutorCasesPass(verification.Cases)
	verification.Digest = canonicalExecutorVerificationDigest(verification)
	return verification
}

func canonicalExecutorVerificationCases(program PolicyProgram, fixture CanonicalExecutorGrantFixture) []CanonicalExecutorVerificationCase {
	variants := []struct {
		id       string
		mutate   func(*CanonicalExecutorGrantFixture)
		expected Decision
		resolution Resolution
		reason   string
	}{
		{"exact-materialized-grant", func(*CanonicalExecutorGrantFixture) {}, DecisionClosed, ResolutionGrantedUnconsumed, ReasonAllow},
		{"missing-candidate-input", func(current *CanonicalExecutorGrantFixture) { current.Request.CandidateInput = valuewitnessinput.ExecutionInput{}; current.Request.V24Request.Candidate.ExecutionInput = nil; current.Request.V25Contract.ExecutionInput = nil; current.Receipt.CandidateStableID = ""; current.Receipt.CandidateDigest = ""; current.Receipt.CandidateInputDigest = ""; current.Receipt.SubjectSHA = ""; current.Receipt.GrantReceipt.Digest = receiptDigest(current.Receipt.GrantReceipt); current.Receipt.Digest = canonicalExecutorReceiptDigest(current.Receipt); current.Request.Digest = canonicalExecutorRequestDigest(current.Request) }, DecisionUnknown, ResolutionLower, CanonicalExecutorUnknownReason},
		{"missing-source-artifact", func(current *CanonicalExecutorGrantFixture) { current.Request.SourceArtifact = CanonicalExecutorSourceArtifact{}; current.Request.GrantRequest.Source = SourceArtifact{}; current.Request.GrantRequest.Digest = requestDigest(current.Request.GrantRequest); current.Request.Digest = canonicalExecutorRequestDigest(current.Request); current.Decision.DecisionInput.Source = SourceArtifact{}; current.Decision.DecisionInput.DecisionDigest = decisionDigest(current.Decision.DecisionInput); current.Decision.Digest = canonicalExecutorDecisionDigest(current.Decision); current.Receipt.SourceArtifact = CanonicalExecutorSourceArtifact{}; current.Receipt.GrantReceipt.Digest = receiptDigest(current.Receipt.GrantReceipt); current.Receipt.Digest = canonicalExecutorReceiptDigest(current.Receipt) }, DecisionUnknown, ResolutionLower, CanonicalExecutorUnknownReason},
		{"missing-source-freshness", func(current *CanonicalExecutorGrantFixture) { current.Request.SourceArtifact.ArtifactExpiryKnown = false; refreshCanonicalExecutorDerived(current) }, DecisionUnknown, ResolutionLower, CanonicalExecutorUnknownReason},
		{"expired-source-artifact", func(current *CanonicalExecutorGrantFixture) { current.Request.SourceArtifact.ArtifactExpired = true; refreshCanonicalExecutorDerived(current) }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SOURCE_BINDING_MISMATCH"},
		{"tampered-candidate-digest", func(current *CanonicalExecutorGrantFixture) { current.Request.CandidateInput.CandidateDigest = digestBytes([]byte("tampered-candidate")) }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_CANDIDATE_BINDING_MISMATCH"},
		{"tampered-v24-binding", func(current *CanonicalExecutorGrantFixture) { current.Request.V24Request.Digest = digestBytes([]byte("tampered-v24-request")) }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_V24_BINDING_MISMATCH"},
		{"tampered-v25-binding", func(current *CanonicalExecutorGrantFixture) { current.Request.V25Contract.Digest = digestBytes([]byte("tampered-v25-contract")) }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_V25_BINDING_MISMATCH"},
		{"tampered-source-artifact", func(current *CanonicalExecutorGrantFixture) { current.Request.SourceArtifact.APIArchiveDigest = digestBytes([]byte("tampered-source")) }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SOURCE_BINDING_MISMATCH"},
		{"tampered-scope", func(current *CanonicalExecutorGrantFixture) { current.Request.GrantRequest.Target = "unbounded" }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SCOPE_BINDING_MISMATCH"},
		{"tampered-ref", func(current *CanonicalExecutorGrantFixture) { current.Request.SourceArtifact.Ref = "refs/heads/main" }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SOURCE_BINDING_MISMATCH"},
		{"tampered-decision", func(current *CanonicalExecutorGrantFixture) { current.Decision.DecisionInput.Decision = DecisionDeny }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_DECISION_BINDING_MISMATCH"},
		{"missing-input-with-tampered-scope", func(current *CanonicalExecutorGrantFixture) { current.Request.CandidateInput = valuewitnessinput.ExecutionInput{}; current.Request.V24Request.Candidate.ExecutionInput = nil; current.Request.V25Contract.ExecutionInput = nil; current.Request.GrantRequest.Target = "unbounded" }, DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SCOPE_BINDING_MISMATCH"},
	}
	result := make([]CanonicalExecutorVerificationCase, 0, len(variants))
	for _, variant := range variants {
		current := fixture
		variant.mutate(&current)
		actual, actualResolution, actualReason, missing, contradictory := classifyCanonicalExecutorFixture(current)
		result = append(result, CanonicalExecutorVerificationCase{ID: variant.id,
			ExpectedDecision: variant.expected, ExpectedResolution: variant.resolution, ExpectedReason: variant.reason,
			ActualDecision: actual, ActualResolution: actualResolution, ActualReason: actualReason,
			MissingFields: missing, ContradictoryFields: contradictory,
			Pass: actual == variant.expected && actualResolution == variant.resolution && actualReason == variant.reason,
		})
	}
	return result
}

func refreshCanonicalExecutorDerived(fixture *CanonicalExecutorGrantFixture) {
	fixture.Request.GrantRequest.Source = fixture.Request.SourceArtifact.ToSourceArtifact()
	fixture.Request.GrantRequest.Digest = requestDigest(fixture.Request.GrantRequest)
	fixture.Request.Digest = canonicalExecutorRequestDigest(fixture.Request)
	fixture.Decision.DecisionInput.RequestDigest = fixture.Request.GrantRequest.Digest
	fixture.Decision.DecisionInput.Source = fixture.Request.GrantRequest.Source
	fixture.Decision.DecisionInput.DecisionDigest = decisionDigest(fixture.Decision.DecisionInput)
	fixture.Decision.Digest = canonicalExecutorDecisionDigest(fixture.Decision)
	fixture.Receipt.RequestDigest = fixture.Request.GrantRequest.Digest
	fixture.Receipt.DecisionDigest = fixture.Decision.Digest
	fixture.Receipt.SourceArtifact = fixture.Request.SourceArtifact
	fixture.Receipt.GrantReceipt.RequestDigest = fixture.Request.GrantRequest.Digest
	fixture.Receipt.GrantReceipt.Digest = receiptDigest(fixture.Receipt.GrantReceipt)
	fixture.Receipt.Digest = canonicalExecutorReceiptDigest(fixture.Receipt)
}

func classifyCanonicalExecutorFixture(fixture CanonicalExecutorGrantFixture) (Decision, Resolution, string, []string, []string) {
	missing, contradictory := canonicalExecutorBindingState(fixture)
	if contains(contradictory, "scope") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SCOPE_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "decision") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_DECISION_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "candidate") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_CANDIDATE_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "v24") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_V24_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "v25") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_V25_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "source") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_SOURCE_BINDING_MISMATCH", missing, contradictory
	}
	if contains(contradictory, "digest") {
		return DecisionRefuted, ResolutionExact, "CANONICAL_EXECUTOR_DIGEST_BINDING_MISMATCH", missing, contradictory
	}
	if len(missing) > 0 {
		return DecisionUnknown, ResolutionLower, CanonicalExecutorUnknownReason, missing, contradictory
	}
	return DecisionClosed, ResolutionGrantedUnconsumed, ReasonAllow, nil, nil
}

func canonicalExecutorBindingState(fixture CanonicalExecutorGrantFixture) ([]string, []string) {
	missing, contradictory := []string{}, []string{}
	addMissing := func(name string) { if !contains(missing, name) { missing = append(missing, name) } }
	addContradictory := func(name string) { if !contains(contradictory, name) { contradictory = append(contradictory, name) } }
	request := fixture.Request
	input := request.CandidateInput
	if request.Schema == "" { addMissing("request") } else if request.Schema != CanonicalExecutorRequestSchema { addContradictory("digest") }
	if request.GrantRequest.Digest == "" || request.Digest == "" { addMissing("request_digest") } else if request.GrantRequest.Digest != requestDigest(request.GrantRequest) || request.Digest != canonicalExecutorRequestDigest(request) { addContradictory("digest") }
	if request.V24Request.Digest == "" { addMissing("v24_request_digest") }
	if request.V24Resolution.Digest == "" { addMissing("v24_resolution_digest") }
	if request.V24Verification.Digest == "" { addMissing("v24_verification_digest") }
	if request.V24Request.Candidate.ContractCanonicalDigest == "" { addMissing("v24_contract_digest") }
	if request.V24Request.Contract.CanonicalDigest == "" { addMissing("v24_authorization_contract_digest") }
	if request.V25Contract.Digest == "" { addMissing("v25_contract_digest") }
	if request.V25Verification.Digest == "" { addMissing("v25_verification_digest") }
	if input.Digest == "" { addMissing("candidate_input_digest") }
	if request.SourceArtifact.ArtifactID == 0 || request.SourceArtifact.APIArchiveDigest == "" || request.SourceArtifact.ObservedArchiveDigest == "" { addMissing("source_artifact") }
	if request.SourceArtifact.WorkflowRunID == 0 || request.SourceArtifact.WorkflowRunAttempt == 0 { addMissing("source_workflow_run") }
	if !request.SourceArtifact.ArtifactExpiryKnown || !request.SourceArtifact.ArtifactRetrieved { addMissing("source_artifact_expiry") }
	if request.SourceArtifact.ArtifactExpired { addContradictory("source") }
	if request.SourceArtifact.ArtifactRetrievalError != "" { addMissing("source_artifact_expiry") }
	if request.GrantRequest.Target != GrantTarget || request.GrantRequest.Mode != GrantMode { addContradictory("scope") }
	if request.GrantRequest.V24 != ProjectV24(request.V24Request, request.V24Resolution) { addContradictory("v24") }
	if request.GrantRequest.V25 != ProjectV25(request.V25Contract) { addContradictory("v25") }
	sourceMissing := request.SourceArtifact.Schema == "" || request.SourceArtifact.WorkflowName == "" || request.SourceArtifact.WorkflowPath == "" || request.SourceArtifact.Repository == "" || request.SourceArtifact.Event == "" || request.SourceArtifact.Ref == "" || request.SourceArtifact.HeadSHA == "" || request.SourceArtifact.ArtifactName == "" || request.SourceArtifact.ArtifactID == 0 || request.SourceArtifact.APIArchiveDigest == "" || request.SourceArtifact.ObservedArchiveDigest == "" || request.SourceArtifact.WorkflowRunID == 0 || request.SourceArtifact.WorkflowRunAttempt == 0
	if sourceMissing { addMissing("source_workflow_identity") }
	if !sourceMissing && (request.SourceArtifact.Schema != CanonicalExecutorSourceSchema || request.SourceArtifact.WorkflowName != CanonicalExecutorWorkflowName || request.SourceArtifact.WorkflowPath != CanonicalExecutorWorkflowPath || request.SourceArtifact.Event != CanonicalExecutorWorkflowEvent || request.SourceArtifact.Ref != CanonicalExecutorWorkflowRef || request.SourceArtifact.ArtifactName != CanonicalExecutorArtifactPrefix+request.SourceArtifact.HeadSHA || !validDigest(request.SourceArtifact.APIArchiveDigest) || !validDigest(request.SourceArtifact.ObservedArchiveDigest) || request.SourceArtifact.APIArchiveDigest != request.SourceArtifact.ObservedArchiveDigest || !validDigest(request.SourceArtifact.EmbeddedVerificationDigest)) { addContradictory("source") }
	if !sourceMissing && request.GrantRequest.Source != request.SourceArtifact.ToSourceArtifact() { addContradictory("source") }
	inputMissing := input.Digest == ""
	if !inputMissing && request.V24Request.Candidate.ExecutionInput != nil && !reflect.DeepEqual(input, *request.V24Request.Candidate.ExecutionInput) { addContradictory("candidate") }
	if !inputMissing && request.V25Contract.ExecutionInput != nil && !reflect.DeepEqual(input, *request.V25Contract.ExecutionInput) { addContradictory("candidate") }
	if input.CandidateStableID == "" || input.CandidateDigest == "" || input.SubjectSHA == "" || input.ObservationDigest == "" { addMissing("candidate_identity") }
	if !inputMissing && (input.CandidateStableID != request.V24Request.Candidate.CandidateID || input.CandidateDigest != request.V24Request.Candidate.CandidateDigest || input.SubjectSHA != request.V25Contract.SubjectSHA || input.ObservationDigest != request.V24Request.Candidate.SourceObservationDigest) { addContradictory("candidate") }
	if !inputMissing {
		if valuewitnessinput.Validate(input) != nil {
			addContradictory("candidate")
		}
		if selfimprovementcandidate.ValidateAuthorizationRequest(request.V24Request) != nil || selfimprovementcandidate.VerifyAuthorizationResolution(request.V24Request, request.V24Resolution) != nil {
			addContradictory("v24")
		} else if expected, err := selfimprovementcandidate.BuildAuthorizationVerification(request.V24Request, request.V24Resolution); err != nil || !reflect.DeepEqual(expected, request.V24Verification) {
			addContradictory("v24")
		}
	}
	if request.V25Contract.Digest != "" && v25.ValidateResolution(request.V25Contract) != nil { addContradictory("v25") }
	if request.V25Verification.ContractDigest != request.V25Contract.Digest || !request.V25Verification.Verified { addContradictory("v25") }
	decision := fixture.Decision
	decisionPresent := decision.Schema != "" || decision.DecisionInput.Decision != "" || decision.Digest != ""
	if !decisionPresent {
		addMissing("decision")
	} else {
		if decision.Schema != CanonicalExecutorDecisionSchema || decision.DecisionType != CanonicalExecutorDecisionType || !decision.Fixture || decision.LiveAuthority || decision.UserDecision || decision.ProductUtilityEvidence || !decision.CanonicalExecutionAllowed || decision.ExecutionScope != CanonicalExecutorScope { addContradictory("decision") }
		if decision.DecisionInput.Decision != DecisionAllow || decision.DecisionInput.RequestDigest != request.GrantRequest.Digest || decision.DecisionInput.V24 != request.GrantRequest.V24 || decision.DecisionInput.V25 != request.GrantRequest.V25 || decision.DecisionInput.Source != request.GrantRequest.Source || decision.DecisionInput.DecisionSource != DecisionSourceCanonical || decision.DecisionInput.ActorEvidence.EvidenceLabel != CanonicalEvidenceLabel { addContradictory("decision") }
		if decision.DecisionInput.DecisionDigest == "" { addMissing("decision_digest") } else if decision.DecisionInput.DecisionDigest != decisionDigest(decision.DecisionInput) { addContradictory("digest") }
		if decision.Digest == "" { addMissing("decision_digest") } else if decision.Digest != canonicalExecutorDecisionDigest(decision) { addContradictory("digest") }
	}
	receipt := fixture.Receipt
	receiptPresent := receipt.Schema != "" || receipt.GrantReceipt.Digest != "" || receipt.Digest != ""
	if !receiptPresent {
		addMissing("receipt")
	} else {
		if receipt.Schema != CanonicalExecutorReceiptSchema || receipt.DecisionType != CanonicalExecutorDecisionType || !receipt.Fixture || receipt.LiveAuthority || !receipt.CanonicalExecutionAllowed || receipt.ExecutionScope != CanonicalExecutorScope || receipt.MaxExecutions != MaxExecutions { addContradictory("receipt") }
		if receipt.RequestDigest != request.GrantRequest.Digest || receipt.DecisionDigest != decision.Digest || receipt.CandidateStableID != input.CandidateStableID || receipt.CandidateDigest != input.CandidateDigest || receipt.CandidateInputDigest != input.Digest || receipt.V24RequestDigest != request.V24Request.Digest || receipt.V24ResolutionDigest != request.V24Resolution.Digest || receipt.V24VerificationDigest != request.V24Verification.Digest || receipt.V24AuthorizationContractDigest != request.V24Request.Contract.CanonicalDigest || receipt.V25ContractDigest != request.V25Contract.Digest || receipt.V25VerificationDigest != request.V25Verification.Digest || receipt.SubjectSHA != input.SubjectSHA || receipt.SourceArtifact != request.SourceArtifact { addContradictory("receipt") }
		if receipt.GrantReceipt.RequestDigest != request.GrantRequest.Digest || receipt.GrantReceipt.Decision != DecisionAllow || receipt.GrantReceipt.DecisionSource != DecisionSourceCanonical || !receipt.GrantReceipt.GrantAllowsExecution || receipt.GrantReceipt.RemainingUses != 1 || receipt.GrantReceipt.ConsumedUses != 0 || receipt.GrantReceipt.ExecutionCount != 0 || receipt.GrantReceipt.ConsumptionStatus != ConsumptionPending || receipt.GrantReceipt.ConsumptionObligation != ConsumptionObligation || receipt.GrantReceipt.OneUseEnforcementState != "PENDING_NEXT_EXECUTOR" || receipt.GrantReceipt.Digest != receiptDigest(receipt.GrantReceipt) { addContradictory("receipt") }
		if receipt.Digest == "" { addMissing("receipt_digest") } else if receipt.Digest != canonicalExecutorReceiptDigest(receipt) { addContradictory("digest") }
	}
	return sortedStrings(missing), sortedStrings(contradictory)
}

func buildCanonicalExecutorBindingManifest(program PolicyProgram, fixture CanonicalExecutorGrantFixture) CanonicalExecutorBindingManifest {
	bound := countCanonicalExecutorBindings(program, fixture)
	return CanonicalExecutorBindingManifest{
		Schema: CanonicalExecutorBindingManifestSchema, FixtureSchema: CanonicalExecutorRequestSchema,
		MaterializedFixtureOf: CanonicalExecutorMaterializedCase, MaterializedFixtureCountBefore: 0, MaterializedFixtureCountAfter: 1,
		ExactGrantIdentityBindingsBefore: 0, ExactGrantIdentityBindingsAfter: bound, ExactGrantIdentityBindings: bound, ExactGrantIdentityBindingDenominator: bound,
		PlaceholderFieldsBefore: countLegacyCanonicalPlaceholders(program), PlaceholderFieldsAfter: countCanonicalExecutorPlaceholders(fixture),
		ArtifactFiles: 5, ArtifactTypes: 5, ArtifactNames: []string{"canonical-executor-grant-request.json", "canonical-executor-grant-decision.json", "canonical-executor-grant-receipt.json", "canonical-executor-grant-verification.json", "canonical-executor-grant-binding-manifest.json"},
		RequestDigest: fixture.Request.GrantRequest.Digest, DecisionDigest: fixture.Decision.Digest, ReceiptDigest: fixture.Receipt.Digest, VerificationDigest: fixture.Verification.Digest,
		CandidateStableID: fixture.Request.CandidateInput.CandidateStableID, CandidateDigest: fixture.Request.CandidateInput.CandidateDigest, CandidateInputDigest: fixture.Request.CandidateInput.Digest, SubjectSHA: fixture.Request.CandidateInput.SubjectSHA,
		V24RequestDigest: fixture.Request.V24Request.Digest, V24ResolutionDigest: fixture.Request.V24Resolution.Digest, V24VerificationDigest: fixture.Request.V24Verification.Digest, V24AuthorizationContractDigest: fixture.Request.V24Request.Contract.CanonicalDigest,
		V25ContractDigest: fixture.Request.V25Contract.Digest, V25VerificationDigest: fixture.Request.V25Verification.Digest, SourceArtifact: fixture.Request.SourceArtifact,
		DecisionType: CanonicalExecutorDecisionType, Fixture: true, LiveAuthority: false, UserDecision: false, ProductUtilityEvidence: false,
		CanonicalExecutionCount: 0, ReceiptRemainingUses: fixture.Receipt.GrantReceipt.RemainingUses, ReceiptConsumedUses: fixture.Receipt.GrantReceipt.ConsumedUses, ReceiptExecutionCount: fixture.Receipt.GrantReceipt.ExecutionCount,
		RepositoryWrites: 0, LocalTestExecutions: 0, ManualDecisions: 0, IndependentReplayComparisons: 1, FallbackAccepted: 0, PerformanceImprovement: PerformanceUnknown,
		GoPhysicalLines: program.Inventory.GoPhysicalLines, GoooPhysicalLines: program.Inventory.GoooPhysicalLines,
		ExecutorContractDigest: program.ExecutorContract.Digest,
		MaterializeActivityDigest: canonicalExecutorActivityDigest(program.ExecutorContract, "MaterializeCanonicalExecutorGrantFixture"),
		VerifyActivityDigest: canonicalExecutorActivityDigest(program.ExecutorContract, "VerifyCanonicalExecutorGrantFixture"),
		Decision: DecisionClosed, Resolution: ResolutionGrantedUnconsumed, Reason: ReasonAllow,
	}
}

// FinalizeCanonicalExecutorBindingManifest records the observations made by
// the artifact writer. The in-memory builder uses zero pre-materialization
// values; the writer supplies its actual directory counts and names here.
func FinalizeCanonicalExecutorBindingManifest(program PolicyProgram, fixture CanonicalExecutorGrantFixture, artifactNames []string, artifactTypes, materializedBefore, bindingsBefore, placeholdersBefore int) (CanonicalExecutorGrantFixture, error) {
	if len(artifactNames) == 0 || artifactTypes <= 0 || artifactTypes > len(artifactNames) || materializedBefore < 0 || bindingsBefore < 0 || placeholdersBefore < 0 {
		return CanonicalExecutorGrantFixture{}, errors.New("canonical executor artifact observations are incomplete")
	}
	manifest := buildCanonicalExecutorBindingManifest(program, fixture)
	manifest.MaterializedFixtureCountBefore = materializedBefore
	manifest.MaterializedFixtureCountAfter = materializedBefore + 1
	manifest.ExactGrantIdentityBindingsBefore = bindingsBefore
	manifest.ExactGrantIdentityBindingsAfter = countCanonicalExecutorBindings(program, fixture)
	manifest.ExactGrantIdentityBindings = manifest.ExactGrantIdentityBindingsAfter
	manifest.PlaceholderFieldsBefore = placeholdersBefore
	manifest.PlaceholderFieldsAfter = countCanonicalExecutorPlaceholders(fixture)
	manifest.ArtifactFiles = len(artifactNames)
	manifest.ArtifactTypes = artifactTypes
	manifest.ArtifactNames = append([]string(nil), artifactNames...)
	sort.Strings(manifest.ArtifactNames)
	manifest.Digest = canonicalExecutorManifestDigest(manifest)
	fixture.Manifest = manifest
	return fixture, nil
}

func ValidateCanonicalExecutorFixture(program PolicyProgram, fixture CanonicalExecutorGrantFixture) error {
	verification := VerifyCanonicalExecutorGrantFixture(program, fixture)
	if !verification.Verified || verification.Digest != fixture.Verification.Digest {
		return errors.New("canonical executor fixture independent verification failed")
	}
	if fixture.Manifest.Schema != CanonicalExecutorBindingManifestSchema || fixture.Manifest.FixtureSchema != CanonicalExecutorRequestSchema || fixture.Manifest.MaterializedFixtureCountAfter != fixture.Manifest.MaterializedFixtureCountBefore+1 || fixture.Manifest.ExactGrantIdentityBindings != fixture.Manifest.ExactGrantIdentityBindingsAfter || fixture.Manifest.ExactGrantIdentityBindingsAfter != len(canonicalExecutorBindingNames) || fixture.Manifest.ArtifactFiles != len(fixture.Manifest.ArtifactNames) || fixture.Manifest.ArtifactTypes != len(fixture.Manifest.ArtifactNames) || fixture.Manifest.LiveAuthority || !fixture.Manifest.Fixture || fixture.Manifest.CanonicalExecutionCount != 0 || fixture.Manifest.ReceiptConsumedUses != 0 || fixture.Manifest.ReceiptExecutionCount != 0 || fixture.Manifest.Digest != canonicalExecutorManifestDigest(fixture.Manifest) {
		return errors.New("canonical executor fixture manifest is not exact")
	}
	return nil
}

func countCanonicalExecutorBindings(program PolicyProgram, fixture CanonicalExecutorGrantFixture) int {
	request, input, source := fixture.Request, fixture.Request.CandidateInput, fixture.Request.SourceArtifact
	v25 := request.GrantRequest.V25
	bindings := map[string]bool{
		"v24_request_digest": request.V24Request.Digest != "",
		"v24_resolution_digest": request.V24Resolution.Digest != "",
		"v24_verification_digest": request.V24Verification.Digest != "",
		"candidate_stable_id": input.CandidateStableID != "",
		"candidate_digest": validDigest(input.CandidateDigest),
		"subject_sha": validSHA(input.SubjectSHA),
		"observation_digest": validDigest(input.ObservationDigest),
		"candidate_input_digest": validDigest(input.Digest),
		"v24_contract_digest": validDigest(request.V24Request.Candidate.ContractCanonicalDigest),
		"v25_contract_digest": validDigest(v25.ContractDigest),
		"v24_authorization_contract_digest": validDigest(request.V24Request.Contract.CanonicalDigest),
		"v25_verification_digest": validDigest(request.V25Verification.Digest),
		"operation_id": v25.OperationID != "",
		"evaluator_registry_digest": validDigest(v25.EvaluatorRegistryDigest),
		"toolchain_test_contract_identity": v25.ToolchainTestContractIdentity != "",
		"max_executions": v25.MaxExecutions == MaxExecutions,
		"repository_writes_allowed": !v25.RepositoryWritesAllowed,
		"source_workflow_name": source.WorkflowName == CanonicalExecutorWorkflowName,
		"source_workflow_path": source.WorkflowPath == CanonicalExecutorWorkflowPath,
		"source_event": source.Event == CanonicalExecutorWorkflowEvent,
		"source_ref": source.Ref == CanonicalExecutorWorkflowRef,
		"source_head_sha": validSHA(source.HeadSHA),
		"source_workflow_run_id": source.WorkflowRunID > 0,
		"source_workflow_run_attempt": source.WorkflowRunAttempt > 0,
		"source_artifact_name": source.ArtifactName == CanonicalExecutorArtifactPrefix+source.HeadSHA,
		"source_artifact_id": source.ArtifactID > 0,
		"source_artifact_archive_digest": validDigest(source.APIArchiveDigest),
		"source_artifact_observed_digest": validDigest(source.ObservedArchiveDigest) && source.ObservedArchiveDigest == source.APIArchiveDigest,
		"source_artifact_expiry": source.ArtifactExpiryKnown && source.ArtifactRetrieved && !source.ArtifactExpired && source.ArtifactRetrievalError == "",
		"executor_semantic_contract_digest": validDigest(program.ExecutorContract.Digest),
		"materialize_activity_digest": validDigest(canonicalExecutorActivityDigest(program.ExecutorContract, "MaterializeCanonicalExecutorGrantFixture")),
		"verify_activity_digest": validDigest(canonicalExecutorActivityDigest(program.ExecutorContract, "VerifyCanonicalExecutorGrantFixture")),
	}
	count := 0
	for _, name := range canonicalExecutorBindingNames {
		if bindings[name] {
			count++
		}
	}
	return count
}

func allCanonicalExecutorCasesPass(cases []CanonicalExecutorVerificationCase) bool {
	for _, item := range cases { if !item.Pass { return false } }
	return true
}

func countCanonicalExecutorPlaceholders(fixture CanonicalExecutorGrantFixture) int {
	raw, _ := json.Marshal(fixture)
	markers := []string{"PLACEHOLDER", "placeholder", "<missing>", "TODO", "synthetic"}
	count := 0
	for _, marker := range markers { if bytes.Contains(raw, []byte(marker)) { count++ } }
	return count
}

func countLegacyCanonicalPlaceholders(program PolicyProgram) int {
	raw, _ := json.Marshal(program)
	markers := []string{"PLACEHOLDER", "placeholder", "<missing>", "TODO", "synthetic"}
	count := 0
	for _, marker := range markers { if bytes.Contains(raw, []byte(marker)) { count++ } }
	return count
}

func canonicalExecutorRequestDigest(value CanonicalExecutorGrantRequest) string { value.Digest = ""; return digestJSON(value) }
func canonicalExecutorDecisionDigest(value CanonicalExecutorGrantDecision) string { value.Digest = ""; return digestJSON(value) }
func canonicalExecutorReceiptDigest(value CanonicalExecutorGrantReceipt) string { value.Digest = ""; return digestJSON(value) }
func canonicalExecutorVerificationDigest(value CanonicalExecutorVerification) string { value.Digest = ""; return digestJSON(value) }
func canonicalExecutorManifestDigest(value CanonicalExecutorBindingManifest) string { value.Digest = ""; return digestJSON(value) }
