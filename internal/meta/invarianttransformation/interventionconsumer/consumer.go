package interventionconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	reportSchema                   = "gooo/invariant-transformation-intervention-report/v2"
	consumerSchema                 = "gooo/invariant-transformation-intervention-consumer/v2"
	semanticExpectedCaseID         = "semantic-expected-intervention"
	semanticOperationCaseID        = "semantic-operation-intervention"
	nonSemanticCaseID              = "nonsemantic-source-intervention"
	preservedCaseID                = "preserved-translation"
	semanticExpectedStep           = "compare-semantic-expected-projection-and-decision"
	semanticExpectedReason         = "SEMANTIC_EXPECTED_VALUE_AND_DECISION_CHANGED"
	semanticOperationStep          = "compare-semantic-operation-projection-and-decision"
	semanticOperationReason        = "SEMANTIC_OPERATION_AND_DECISION_CHANGED"
	nonSemanticStep                = "compare-nonsemantic-projection-and-decision"
	nonSemanticReason              = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
	semanticContradictionReason    = "SEMANTIC_INTERVENTION_CONTRADICTED"
	nonSemanticContradictionReason = "NONSEMANTIC_INTERVENTION_CONTRADICTED"
	evidenceUnobservableReason     = "INTERVENTION_EVIDENCE_UNOBSERVABLE"
)

type DependencyBoundary struct {
	ProducerDependencyImports        int                    `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int                    `json:"allowed_producer_dependency_imports"`
	ArtifactEvidence                 model.ArtifactEvidence `json:"artifact_evidence"`
	UnknownEffectScopes              int                    `json:"unknown_effect_scopes"`
}

type Audit struct {
	Schema                            string                 `json:"schema"`
	HeadSHA                           string                 `json:"head_sha"`
	ProducerDependencyImports         int                    `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports  int                    `json:"allowed_producer_dependency_imports"`
	ReconstructedCases                int                    `json:"reconstructed_cases"`
	ExpectedCases                     int                    `json:"expected_cases"`
	ActualReplay                      int                    `json:"actual_replay"`
	ExpectedActualReplay              int                    `json:"expected_actual_replay"`
	ArtifactEvidence                  model.ArtifactEvidence `json:"artifact_evidence"`
	ArtifactObserved                  bool                   `json:"artifact_observed"`
	CoherentTamperRejected            int                    `json:"coherent_tamper_rejected"`
	ExpectedCoherentTamperRejections  int                    `json:"expected_coherent_tamper_rejections"`
	Decision                          string                 `json:"decision"`
	Resolution                        string                 `json:"resolution"`
	Reason                            string                 `json:"reason"`
	RepositoryNetStatusUnchanged      bool                   `json:"repository_net_status_unchanged"`
	RepositoryNetStatusObserved       bool                   `json:"repository_net_status_observed"`
	RepositoryNetState                string                 `json:"repository_net_state"`
	RepositoryActualOrTransientWrites string                 `json:"repository_actual_or_transient_writes"`
	RepositoryPathAuthorization       bool                   `json:"repository_path_authorization"`
	AmbientProcessAuthority           string                 `json:"ambient_process_authority"`
	UnknownEffectScopes               int                    `json:"unknown_effect_scopes"`
	Digest                            string                 `json:"digest"`
}

type coordinateWire struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type transitionWire struct {
	ClaimID                  string         `json:"claim_id"`
	From                     string         `json:"from"`
	To                       string         `json:"to"`
	Coordinate               coordinateWire `json:"coordinate"`
	PropositionDigest        string         `json:"proposition_digest"`
	PriorStateDigest         string         `json:"prior_state_digest"`
	EvidenceDigest           string         `json:"evidence_digest"`
	PreviousTransitionDigest string         `json:"previous_transition_digest"`
	CurrentTransitionDigest  string         `json:"current_transition_digest"`
}
type claimWire struct {
	ID                string           `json:"id"`
	Status            string           `json:"status"`
	Resolution        string           `json:"resolution"`
	Reason            string           `json:"reason"`
	VerificationCheck string           `json:"verification_check"`
	Coordinate        coordinateWire   `json:"coordinate"`
	TargetDigest      string           `json:"target_digest"`
	PriorStateDigest  string           `json:"prior_state_digest"`
	EvidenceDigest    string           `json:"evidence_digest"`
	Transitions       []transitionWire `json:"transitions"`
}
type projectionWire struct {
	Activity             string `json:"activity"`
	CaseID               string `json:"case_id"`
	CaseKind             string `json:"case_kind"`
	Input                int64  `json:"input"`
	CandidateOperation   string `json:"candidate_operation"`
	CandidateResult      int64  `json:"candidate_result"`
	Expected             int64  `json:"expected"`
	Invariant            string `json:"invariant"`
	InvariantID          string `json:"invariant_id"`
	DomainID             string `json:"input_domain_id"`
	OperationID          string `json:"operation_id"`
	ReplayRecipe         string `json:"replay_recipe"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	EffectIntent         string `json:"effect_intent"`
}
type caseWire struct {
	ID                                        string                       `json:"id"`
	Kind                                      string                       `json:"kind"`
	SourceEdit                                string                       `json:"source_edit"`
	BaselineProjection                        projectionWire               `json:"baseline_projection"`
	MutatedProjection                         projectionWire               `json:"mutated_projection"`
	BaselineProjectionDigest                  string                       `json:"baseline_projection_digest"`
	MutatedProjectionDigest                   string                       `json:"mutated_projection_digest"`
	BaselineSourceDigest                      string                       `json:"baseline_source_digest"`
	MutatedSourceDigest                       string                       `json:"mutated_source_digest"`
	BaselineReceiptDigest                     string                       `json:"baseline_receipt_digest"`
	MutatedReceiptDigest                      string                       `json:"mutated_receipt_digest"`
	BaselineReceiptDecision                   string                       `json:"baseline_receipt_decision"`
	MutatedReceiptDecision                    string                       `json:"mutated_receipt_decision"`
	BaselineJudgment                          model.Judgment               `json:"baseline_judgment"`
	MutatedJudgment                           model.Judgment               `json:"mutated_judgment"`
	BaselineEvidence                          model.TransformationEvidence `json:"baseline_evidence"`
	MutatedEvidence                           model.TransformationEvidence `json:"mutated_evidence"`
	BaselineClaimTransitions                  []transitionWire             `json:"baseline_claim_transitions"`
	MutatedClaimTransitions                   []transitionWire             `json:"mutated_claim_transitions"`
	RawSourceDigestChanged                    bool                         `json:"raw_source_digest_changed"`
	ReceiptChanged                            bool                         `json:"receipt_changed"`
	SemanticProjectionEqual                   bool                         `json:"semantic_projection_equal"`
	DecisionEqual                             bool                         `json:"decision_equal"`
	ResolutionEqual                           bool                         `json:"resolution_equal"`
	ReasonEqual                               bool                         `json:"reason_equal"`
	DecisionChanged                           bool                         `json:"decision_changed"`
	ClaimTransitionsEqual                     bool                         `json:"claim_transitions_equal"`
	EffectsEqual                              bool                         `json:"effects_equal"`
	ReplayObservationEqual                    bool                         `json:"replay_observation_equal"`
	EvidenceObservable                        bool                         `json:"evidence_observable"`
	RepositoryWritesNotClaimed                bool                         `json:"repository_writes_not_claimed"`
	BaselineRepositoryWrites                  int                          `json:"baseline_repository_writes"`
	MutatedRepositoryWrites                   int                          `json:"mutated_repository_writes"`
	BaselineRepositoryWritesObserved          bool                         `json:"baseline_repository_writes_observed"`
	MutatedRepositoryWritesObserved           bool                         `json:"mutated_repository_writes_observed"`
	BaselineRepositoryNetStatusUnchanged      bool                         `json:"baseline_repository_net_status_unchanged"`
	MutatedRepositoryNetStatusUnchanged       bool                         `json:"mutated_repository_net_status_unchanged"`
	BaselineRepositoryActualOrTransientWrites string                       `json:"baseline_repository_actual_or_transient_writes"`
	MutatedRepositoryActualOrTransientWrites  string                       `json:"mutated_repository_actual_or_transient_writes"`
	BaselineRepositoryMutationAuthorized      bool                         `json:"baseline_repository_mutation_authorized"`
	MutatedRepositoryMutationAuthorized       bool                         `json:"mutated_repository_mutation_authorized"`
	Claim                                     claimWire                    `json:"claim"`
	Satisfied                                 bool                         `json:"satisfied"`
}
type denominatorWire struct {
	ID             string `json:"id"`
	CasesTotal     int    `json:"cases_total"`
	CasesSatisfied int    `json:"cases_satisfied"`
	CoverageBPS    int    `json:"coverage_bps"`
}
type fixedDenominatorWire struct {
	ID                      string          `json:"id"`
	CasesTotal              int             `json:"cases_total"`
	SemanticExpectedChange  denominatorWire `json:"semantic_expected_change"`
	SemanticOperationChange denominatorWire `json:"semantic_operation_change"`
	NonSemantic             denominatorWire `json:"nonsemantic_change"`
}
type gateWire struct {
	ID                     string                 `json:"id"`
	Scenario               string                 `json:"scenario"`
	CaseID                 string                 `json:"case_id"`
	SubjectSHA             string                 `json:"subject_sha"`
	Stage                  string                 `json:"stage"`
	Step                   string                 `json:"step"`
	AttemptPath            string                 `json:"attempt_path"`
	TargetPath             string                 `json:"target_path"`
	TargetBeforeExists     bool                   `json:"target_before_exists"`
	TargetAfterExists      bool                   `json:"target_after_exists"`
	TargetBeforeDigest     string                 `json:"target_before_digest"`
	TargetAfterDigest      string                 `json:"target_after_digest"`
	TargetBytesUnchanged   bool                   `json:"target_bytes_unchanged"`
	AuthorizationAttempted bool                   `json:"authorization_attempted"`
	AuthorizationAccepted  bool                   `json:"authorization_accepted"`
	ExecutorAccepted       bool                   `json:"executor_accepted"`
	ArtifactCount          int                    `json:"artifact_count"`
	ArtifactExists         bool                   `json:"artifact_exists"`
	Artifact               model.ArtifactEvidence `json:"artifact"`
	Reason                 string                 `json:"reason"`
	Satisfied              bool                   `json:"satisfied"`
}
type reportWire struct {
	Schema                            string               `json:"schema"`
	HeadSHA                           string               `json:"head_sha"`
	SourcePath                        string               `json:"source_path"`
	SourceDigest                      string               `json:"source_digest"`
	Denominator                       fixedDenominatorWire `json:"denominator"`
	CaseCount                         int                  `json:"case_count"`
	Cases                             []caseWire           `json:"cases"`
	EffectGates                       []gateWire           `json:"effect_gates"`
	EffectGateDenominator             int                  `json:"effect_gate_denominator"`
	EffectGateSatisfied               int                  `json:"effect_gate_satisfied"`
	Decision                          string               `json:"decision"`
	Resolution                        string               `json:"resolution"`
	Reason                            string               `json:"reason"`
	RepositoryWrites                  int                  `json:"repository_writes"`
	RepositoryMutationAuthorized      bool                 `json:"repository_mutation_authorized"`
	TempArtifactWriteAuthorized       bool                 `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool                 `json:"repository_net_status_unchanged"`
	RepositoryNetStatusObserved       bool                 `json:"repository_net_status_observed"`
	RepositoryNetState                string               `json:"repository_net_state"`
	RepositoryActualOrTransientWrites string               `json:"repository_actual_or_transient_writes"`
	RepositoryPathAuthorization       bool                 `json:"repository_path_authorization"`
	AmbientProcessAuthority           string               `json:"ambient_process_authority"`
	ExecutedEffects                   int                  `json:"executed_effects"`
	IndependentlyObservedEffects      int                  `json:"independently_observed_effects"`
	UnknownEffectScopes               int                  `json:"unknown_effect_scopes"`
	CorrectionCount                   int                  `json:"correction_count"`
	CorrectionDenominator             int                  `json:"correction_denominator"`
	Failure                           json.RawMessage      `json:"failure,omitempty"`
	Digest                            string               `json:"digest"`
}

type sourceFixture struct {
	Activity, CaseID, CaseKind                                                                      string
	Input                                                                                           int64
	CandidateOperation                                                                              string
	CandidateResult                                                                                 int64
	Expected                                                                                        int64
	Invariant, InvariantID, DomainID, OperationID, ReplayRecipe, EffectIntent, SemanticSourceDigest string
}

func VerifyReport(raw, source []byte, headSHA string, dependency DependencyBoundary) (Audit, error) {
	if dependency.ProducerDependencyImports != 0 || dependency.AllowedProducerDependencyImports != 0 {
		return Audit{}, fmt.Errorf("consumer producer dependency boundary is not 0/0")
	}
	if err := verifyArtifact(dependency.ArtifactEvidence, headSHA, source); err != nil {
		return Audit{}, err
	}
	observed, err := decodeReport(raw)
	if err != nil {
		return Audit{}, err
	}
	if err := independentlyVerify(observed, source, headSHA, dependency); err != nil {
		return Audit{}, err
	}
	tampered := coherentTamper(observed)
	if independentlyVerify(tampered, source, headSHA, dependency) == nil {
		return Audit{}, fmt.Errorf("coherent resealed tamper was accepted")
	}
	return Audit{Schema: consumerSchema, HeadSHA: headSHA, ProducerDependencyImports: dependency.ProducerDependencyImports, AllowedProducerDependencyImports: dependency.AllowedProducerDependencyImports, ReconstructedCases: len(observed.Cases), ExpectedCases: 3, ActualReplay: 3, ExpectedActualReplay: 3, ArtifactEvidence: dependency.ArtifactEvidence, ArtifactObserved: true, CoherentTamperRejected: 1, ExpectedCoherentTamperRejections: 1, Decision: "PASS", Resolution: model.ResolutionExact, Reason: "INDEPENDENT_SOURCE_RECONSTRUCTION_AND_EFFECT_OBSERVATION", RepositoryNetStatusObserved: observed.RepositoryNetStatusObserved, RepositoryNetStatusUnchanged: observed.RepositoryNetStatusUnchanged, RepositoryActualOrTransientWrites: observed.RepositoryActualOrTransientWrites, UnknownEffectScopes: observed.UnknownEffectScopes}, nil
}

func decodeReport(raw []byte) (reportWire, error) {
	var report reportWire
	if err := json.Unmarshal(raw, &report); err != nil {
		return reportWire{}, fmt.Errorf("decode intervention report: %w", err)
	}
	return report, nil
}

func independentlyVerify(report reportWire, source []byte, headSHA string, dependency DependencyBoundary) error {
	if report.Schema != reportSchema || report.HeadSHA != headSHA || report.SourcePath != model.SourcePath || report.SourceDigest != model.DigestBytes(source) || report.CaseCount != 3 || len(report.Cases) != 3 || report.Digest == "" || report.Digest != reseal(report).Digest {
		return fmt.Errorf("intervention report identity or digest is invalid")
	}
	if report.Decision != "PASS" || report.Resolution != model.ResolutionExact || report.Reason != "ALL_INTERVENTION_OBSERVATIONS_SATISFIED" || report.EffectGateDenominator != 8 || report.EffectGateSatisfied != 8 || report.CorrectionCount != 12 || report.CorrectionDenominator != 12 || report.RepositoryWrites != -1 || report.RepositoryMutationAuthorized || report.RepositoryNetStatusObserved || report.RepositoryNetStatusUnchanged || report.RepositoryNetState != model.RepositoryNetStateUnknown || report.RepositoryActualOrTransientWrites != model.UnknownEffectScope || report.RepositoryPathAuthorization || report.AmbientProcessAuthority != model.UnknownEffectScope || report.ExecutedEffects != 1 || report.IndependentlyObservedEffects != 1 || report.UnknownEffectScopes != 1 {
		return fmt.Errorf("intervention report top result or gate denominator is invalid")
	}
	if report.Denominator.ID != "gooo/invariant-transformation-intervention-denominator/v2" || report.Denominator.CasesTotal != 3 || report.Denominator.SemanticExpectedChange.CasesSatisfied != 1 || report.Denominator.SemanticOperationChange.CasesSatisfied != 1 || report.Denominator.NonSemantic.CasesSatisfied != 1 {
		return fmt.Errorf("intervention denominator is invalid")
	}
	mutators := map[string]func([]byte) ([]byte, error){semanticExpectedCaseID: mutateSemantic, semanticOperationCaseID: mutateOperation, nonSemanticCaseID: mutateNonSemantic}
	steps := map[string][2]string{semanticExpectedCaseID: {semanticExpectedStep, semanticExpectedReason}, semanticOperationCaseID: {semanticOperationStep, semanticOperationReason}, nonSemanticCaseID: {nonSemanticStep, nonSemanticReason}}
	seen := map[string]bool{}
	for _, item := range report.Cases {
		mutate, ok := mutators[item.ID]
		if !ok || seen[item.ID] {
			return fmt.Errorf("unknown or duplicate intervention case %q", item.ID)
		}
		seen[item.ID] = true
		mutated, err := mutate(source)
		if err != nil {
			return err
		}
		baselineFixture, err := parseFixture(source)
		if err != nil {
			return err
		}
		mutatedFixture, err := parseFixture(mutated)
		if err != nil {
			return err
		}
		baselineReceipt := reconstructReceipt(baselineFixture, source, headSHA)
		mutatedReceipt := reconstructReceipt(mutatedFixture, mutated, headSHA)
		baselineJudgment := reconstructJudgment(baselineReceipt)
		mutatedJudgment := reconstructJudgment(mutatedReceipt)
		if err := compareCase(item, baselineFixture, mutatedFixture, baselineReceipt, mutatedReceipt, baselineJudgment, mutatedJudgment, steps[item.ID]); err != nil {
			return err
		}
	}
	if !seen[semanticExpectedCaseID] || !seen[semanticOperationCaseID] || !seen[nonSemanticCaseID] {
		return fmt.Errorf("intervention case inventory is incomplete")
	}
	if err := verifyGates(report.EffectGates, source, headSHA); err != nil {
		return err
	}
	if dependency.UnknownEffectScopes != report.UnknownEffectScopes {
		return fmt.Errorf("unknown effect scope metric mismatch")
	}
	return nil
}

func compareCase(item caseWire, baseline, mutated sourceFixture, baselineReceipt, mutatedReceipt model.Receipt, baselineJudgment, mutatedJudgment model.Judgment, step [2]string) error {
	if !reflect.DeepEqual(item.BaselineProjection, projection(baseline)) || !reflect.DeepEqual(item.MutatedProjection, projection(mutated)) || item.BaselineProjectionDigest != model.Digest(projection(baseline)) || item.MutatedProjectionDigest != model.Digest(projection(mutated)) || item.BaselineSourceDigest != baselineReceipt.SourceDigest || item.MutatedSourceDigest != mutatedReceipt.SourceDigest || item.BaselineReceiptDigest != baselineReceipt.Digest || item.MutatedReceiptDigest != mutatedReceipt.Digest || item.BaselineReceiptDecision != baselineReceipt.Decision || item.MutatedReceiptDecision != mutatedReceipt.Decision || !reflect.DeepEqual(item.BaselineEvidence, baselineReceipt.Evidence) || !reflect.DeepEqual(item.MutatedEvidence, mutatedReceipt.Evidence) || !reflect.DeepEqual(item.BaselineJudgment, baselineJudgment) || !reflect.DeepEqual(item.MutatedJudgment, mutatedJudgment) || !reflect.DeepEqual(item.BaselineClaimTransitions, transitionWires(baselineReceipt.Claims)) || !reflect.DeepEqual(item.MutatedClaimTransitions, transitionWires(mutatedReceipt.Claims)) {
		return fmt.Errorf("case %q is not independently reconstructed", item.ID)
	}
	if item.Claim.ID != item.ID+"::claim" || item.Claim.Status != model.StatusDischarged || len(item.Claim.Transitions) != 1 || item.Claim.Transitions[0].From != model.StatusOpen || item.Claim.Transitions[0].To != model.StatusDischarged || item.Claim.Coordinate.Stage != "INTERVENTION" || item.Claim.Coordinate.Step != step[0] || item.Claim.Coordinate.Reason != step[1] {
		return fmt.Errorf("case %q claim provenance is invalid", item.ID)
	}
	rawChanged := baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest
	receiptChanged := baselineReceipt.Digest != mutatedReceipt.Digest
	semanticEqual := reflect.DeepEqual(projection(baseline), projection(mutated))
	decisionEqual := baselineJudgment.Decision == mutatedJudgment.Decision
	resolutionEqual := baselineJudgment.Resolution == mutatedJudgment.Resolution
	reasonEqual := baselineJudgment.Reason == mutatedJudgment.Reason
	transitionEqual := transitionOutcomes(baselineReceipt.Claims, mutatedReceipt.Claims)
	replayEqual := replayObservationEqual(baselineReceipt.Evidence, mutatedReceipt.Evidence)
	transitionEvidence := model.Digest([]any{model.Digest(projection(baseline)), model.Digest(projection(mutated)), baselineJudgment.Decision, mutatedJudgment.Decision, baselineJudgment.Resolution, mutatedJudgment.Resolution, baselineJudgment.Reason, mutatedJudgment.Reason})
	claimCoordinate := model.Coordinate{Stage: "INTERVENTION", Step: step[0], Reason: step[1]}
	expectedTransition := model.NewTransition(item.ID+"::claim", model.StatusOpen, model.StatusDischarged, claimCoordinate, transitionEvidence)
	if item.Claim.VerificationCheck != "intervention-observation-derived-from-two-independent-receipts" || item.Claim.Resolution != model.ResolutionExact || item.Claim.Reason != step[1] || item.Claim.TargetDigest != expectedTransition.PropositionDigest || item.Claim.PriorStateDigest != expectedTransition.PriorStateDigest || item.Claim.EvidenceDigest != expectedTransition.EvidenceDigest || !reflect.DeepEqual(item.Claim.Transitions[0], transitionFromModel(expectedTransition)) {
		return fmt.Errorf("case %q claim ledger provenance is invalid", item.ID)
	}
	if item.RawSourceDigestChanged != rawChanged || item.ReceiptChanged != receiptChanged || item.SemanticProjectionEqual != semanticEqual || item.DecisionEqual != decisionEqual || item.ResolutionEqual != resolutionEqual || item.ReasonEqual != reasonEqual || item.DecisionChanged != !decisionEqual || item.ClaimTransitionsEqual != transitionEqual || item.EffectsEqual != true || item.ReplayObservationEqual != replayEqual || !item.EvidenceObservable || !item.RepositoryWritesNotClaimed || item.BaselineRepositoryWrites != -1 || item.MutatedRepositoryWrites != -1 || item.BaselineRepositoryWritesObserved || item.MutatedRepositoryWritesObserved || item.BaselineRepositoryNetStatusUnchanged || item.MutatedRepositoryNetStatusUnchanged || item.BaselineRepositoryActualOrTransientWrites != model.UnknownEffectScope || item.MutatedRepositoryActualOrTransientWrites != model.UnknownEffectScope || !item.Satisfied {
		return fmt.Errorf("case %q observation fields are not derived", item.ID)
	}
	return nil
}

func verifyGates(gates []gateWire, source []byte, headSHA string) error {
	if len(gates) != 8 {
		return fmt.Errorf("effect gate denominator is not 8")
	}
	expectedIDs := map[string]bool{
		"effect-unauthorized":          false,
		"effect-refuted":               false,
		"effect-open":                  false,
		"effect-stale-sha":             false,
		"effect-tampered-auth":         false,
		"effect-valid-repository-path": false,
		"effect-valid-temp-symlink":    false,
		"effect-valid-authorized":      false,
	}
	expectedReasons := map[string]string{
		"effect-unauthorized":          "UNAUTHORIZED_DECISION_REJECTED",
		"effect-refuted":               "REFUTED_RECEIPT_REJECTED",
		"effect-open":                  "OPEN_RECEIPT_REJECTED",
		"effect-stale-sha":             "STALE_SUBJECT_SHA_REJECTED",
		"effect-tampered-auth":         "TAMPERED_AUTHORIZATION_REJECTED",
		"effect-valid-repository-path": "REPOSITORY_TARGET_REJECTED",
		"effect-valid-temp-symlink":    "TEMP_SYMLINK_ESCAPE_REJECTED",
		"effect-valid-authorized":      "AUTHORIZED_TEMP_ARTIFACT_OBSERVED",
	}
	expectedCoordinates := map[string][2]string{
		"effect-unauthorized":          {"EFFECT_AUTHORIZATION", "validate-authorization"},
		"effect-refuted":               {"EFFECT_AUTHORIZATION", "validate-authorization"},
		"effect-open":                  {"EFFECT_AUTHORIZATION", "validate-authorization"},
		"effect-stale-sha":             {"EFFECT_AUTHORIZATION", "validate-authorization"},
		"effect-tampered-auth":         {"EFFECT_AUTHORIZATION", "validate-authorization"},
		"effect-valid-repository-path": {"EFFECT_AUTHORIZATION", "validate-temp-root-containment"},
		"effect-valid-temp-symlink":    {"EFFECT_AUTHORIZATION", "validate-rooted-target"},
		"effect-valid-authorized":      {"EFFECT_EXECUTION", "write-authorized-temp-artifact"},
	}
	approvedFixture, err := parseFixtureCase(source, "approved-artifact")
	if err != nil {
		return err
	}
	approvedReceipt := reconstructReceipt(approvedFixture, source, headSHA)
	approvedAuth := approvedReceipt.AuthorizationDigest
	for _, gate := range gates {
		seen, ok := expectedIDs[gate.ID]
		if !ok || seen {
			return fmt.Errorf("effect gate inventory contains duplicate or unknown case %q", gate.ID)
		}
		expectedIDs[gate.ID] = true
		if gate.AuthorizationAttempted != true {
			return fmt.Errorf("effect gate %q lacks authorization attempt", gate.ID)
		}
		coordinate := expectedCoordinates[gate.ID]
		if gate.Stage != coordinate[0] || gate.Step != coordinate[1] || gate.Reason != expectedReasons[gate.ID] {
			return fmt.Errorf("effect gate %q stage/step/reason is not bound", gate.ID)
		}
		if gate.ID == "effect-valid-authorized" {
			if !gate.AuthorizationAccepted || !gate.ExecutorAccepted || gate.ArtifactCount != 1 || !gate.ArtifactExists || !gate.Satisfied || gate.SubjectSHA != headSHA || gate.CaseID != "approved-artifact" || gate.AttemptPath != gate.Artifact.Path || gate.TargetPath != "" || gate.TargetBytesUnchanged || gate.Artifact.Path == "" || !allowedTempPath(gate.Artifact.Path) || gate.Artifact.CaseID != "approved-artifact" || gate.Artifact.ExecutionID != approvedReceipt.ExecutionID || gate.Artifact.SubjectSHA != headSHA || !model.ValidDigest(gate.Artifact.ContentDigest) || gate.Artifact.Size <= 0 || gate.Artifact.AuthorizationDigest != approvedAuth || gate.Artifact.Producer != model.ProducerID || gate.Artifact.Consumer != model.ConsumerID || gate.Artifact.Executor != model.ExecutorID || gate.Artifact.RepositoryNetStatusObserved || gate.Artifact.RepositoryNetStatusUnchanged || gate.Artifact.RepositoryNetState != model.RepositoryNetStateUnknown {
				return fmt.Errorf("valid effect gate is not observed")
			}
			data, err := os.ReadFile(gate.Artifact.Path)
			if err != nil || len(data) != gate.Artifact.Size || model.DigestBytes(data) != gate.Artifact.ContentDigest || !bytes.Equal(data, artifactBytes(approvedReceipt)) {
				return fmt.Errorf("valid effect artifact bytes are not observed")
			}
			expectedEffect := model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: gate.Artifact.ContentDigest, ArtifactPath: gate.Artifact.Path, ArtifactSize: gate.Artifact.Size, Artifact: gate.Artifact, CaseID: approvedReceipt.CaseID, SubjectSHA: headSHA, Intent: approvedFixture.EffectIntent, AuthorizationDigest: approvedAuth, Producer: model.ProducerID, Executor: model.ExecutorID, Consumer: model.ConsumerID, MetaOperation: "execute-authorized-temp-artifact", TempArtifactWriteAuthorized: true, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope}
			if gate.Artifact.EffectReceiptDigest != model.EffectExecutionDigest(expectedEffect) {
				return fmt.Errorf("valid effect gate execution digest is not bound")
			}
		} else if gate.ID != "effect-valid-repository-path" && gate.ID != "effect-valid-temp-symlink" {
			expectedAuthorization := gate.ID == "effect-stale-sha"
			if gate.CaseID == "" || gate.SubjectSHA == "" || gate.AuthorizationAccepted != expectedAuthorization || gate.ExecutorAccepted || gate.ArtifactCount != 0 || gate.ArtifactExists || !reflect.DeepEqual(gate.Artifact, model.ArtifactEvidence{}) || gate.Satisfied != true {
				return fmt.Errorf("rejected effect gate %q created an artifact", gate.ID)
			}
		}
		if gate.ID == "effect-valid-repository-path" || gate.ID == "effect-valid-temp-symlink" {
			if !gate.AuthorizationAccepted || gate.ExecutorAccepted || gate.ArtifactCount != 0 || gate.ArtifactExists || !gate.Satisfied || gate.SubjectSHA != headSHA || gate.CaseID != "approved-artifact" || !reflect.DeepEqual(gate.Artifact, model.ArtifactEvidence{}) || !gate.TargetBytesUnchanged || gate.TargetPath == "" || gate.TargetBeforeExists != gate.TargetAfterExists || gate.TargetBeforeDigest != gate.TargetAfterDigest || !targetSnapshotMatches(gate.TargetPath, gate.TargetAfterExists, gate.TargetAfterDigest) {
				return fmt.Errorf("effect gate %q did not prove target bytes unchanged", gate.ID)
			}
			if gate.ID == "effect-valid-repository-path" {
				if gate.AttemptPath != gate.TargetPath || !allowedRepositoryPath(gate.TargetPath) {
					return fmt.Errorf("repository path gate target is not bound")
				}
			} else if !lexicalTempPath(gate.AttemptPath) || !allowedRepositoryPath(gate.TargetPath) || gate.AttemptPath == gate.TargetPath || !isSymlinkTo(gate.AttemptPath, gate.TargetPath) {
				return fmt.Errorf("temporary symlink gate target is not bound")
			}
		}
	}
	for id, seen := range expectedIDs {
		if !seen {
			return fmt.Errorf("effect gate %q is missing", id)
		}
	}
	return nil
}

func verifyArtifact(artifact model.ArtifactEvidence, headSHA string, source []byte) error {
	fixture, err := parseFixtureCase(source, "approved-artifact")
	if err != nil {
		return fmt.Errorf("artifact source reconstruction failed: %w", err)
	}
	receipt := reconstructReceipt(fixture, source, headSHA)
	if artifact.Path == "" || !allowedTempPath(artifact.Path) || !model.ValidDigest(artifact.ContentDigest) || artifact.Size <= 0 || artifact.CaseID != "approved-artifact" || artifact.ExecutionID != receipt.ExecutionID || artifact.SubjectSHA != headSHA || !model.ValidDigest(artifact.AuthorizationDigest) || !model.ValidDigest(artifact.EffectReceiptDigest) || artifact.Producer != model.ProducerID || artifact.Executor != model.ExecutorID || artifact.Consumer != model.ConsumerID || artifact.RepositoryNetStatusObserved || artifact.RepositoryNetStatusUnchanged || artifact.RepositoryNetState != model.RepositoryNetStateUnknown {
		return fmt.Errorf("artifact evidence is incomplete")
	}
	data, err := os.ReadFile(artifact.Path)
	if err != nil || len(data) != artifact.Size || model.DigestBytes(data) != artifact.ContentDigest {
		return fmt.Errorf("artifact evidence bytes are not observed")
	}
	if artifact.CaseID != receipt.CaseID || artifact.AuthorizationDigest != receipt.AuthorizationDigest || !bytes.Equal(data, artifactBytes(receipt)) {
		return fmt.Errorf("artifact evidence is not source-derived")
	}
	expectedEffect := model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: artifact.ContentDigest, ArtifactPath: artifact.Path, ArtifactSize: artifact.Size, Artifact: artifact, CaseID: receipt.CaseID, ExecutionID: receipt.ExecutionID, SubjectSHA: headSHA, Intent: fixture.EffectIntent, AuthorizationDigest: receipt.AuthorizationDigest, Producer: model.ProducerID, Executor: model.ExecutorID, Consumer: model.ConsumerID, MetaOperation: "execute-authorized-temp-artifact", TempArtifactWriteAuthorized: true, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope}
	if artifact.EffectReceiptDigest != model.EffectExecutionDigest(expectedEffect) {
		return fmt.Errorf("artifact effect receipt digest is not source-derived")
	}
	return nil
}

func allowedTempPath(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, root := range tempRoots() {
		canonicalRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		canonicalRoot, err = filepath.Abs(canonicalRoot)
		if err != nil || overlapsRepository(canonicalRoot) || allowedRepositoryPath(path) || !withinPath(canonicalRoot, path) {
			continue
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		resolved, err = filepath.Abs(resolved)
		if err == nil && withinPath(canonicalRoot, resolved) {
			return true
		}
	}
	return false
}

func lexicalTempPath(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, root := range tempRoots() {
		canonicalRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		canonicalRoot, err = filepath.Abs(canonicalRoot)
		if err == nil && !overlapsRepository(canonicalRoot) && !allowedRepositoryPath(path) && withinPath(canonicalRoot, path) {
			return true
		}
	}
	return false
}

func tempRoots() []string {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	return []string{root}
}

func withinPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func overlapsRepository(root string) bool {
	repository, ok := repositoryRoot()
	return ok && (withinPath(root, repository) || withinPath(repository, root))
}

func allowedRepositoryPath(path string) bool {
	repository, ok := repositoryRoot()
	if !ok {
		return false
	}
	path, err := filepath.Abs(path)
	if err != nil || !withinPath(repository, path) {
		return false
	}
	if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
		resolved, err = filepath.Abs(resolved)
		return err == nil && withinPath(repository, resolved)
	}
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		if resolved, resolveErr := filepath.EvalSymlinks(parent); resolveErr == nil {
			resolved, err = filepath.Abs(resolved)
			return err == nil && withinPath(repository, resolved)
		}
		next := filepath.Dir(parent)
		if next == parent {
			return false
		}
	}
}

func repositoryRoot() (string, bool) {
	root := os.Getenv("GITHUB_WORKSPACE")
	if root == "" {
		root, _ = os.Getwd()
		for {
			if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
				break
			}
			parent := filepath.Dir(root)
			if parent == root {
				return "", false
			}
			root = parent
		}
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false
	}
	root, err = filepath.Abs(root)
	return root, err == nil
}

func snapshotPath(path string) (bool, string) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false, ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, ""
	}
	return true, model.DigestBytes(data)
}

func targetSnapshotMatches(path string, exists bool, digest string) bool {
	actualExists, actualDigest := snapshotPath(path)
	return actualExists == exists && actualDigest == digest
}

func isSymlinkTo(path, target string) bool {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	link, err := os.Readlink(path)
	if err != nil {
		return false
	}
	absoluteLink := link
	if !filepath.IsAbs(absoluteLink) {
		absoluteLink = filepath.Join(filepath.Dir(path), link)
	}
	absoluteLink, err = filepath.Abs(absoluteLink)
	if err != nil {
		return false
	}
	absoluteTarget, err := filepath.Abs(target)
	return err == nil && absoluteLink == absoluteTarget
}

func artifactBytes(receipt model.Receipt) []byte {
	return []byte(fmt.Sprintf("gooo bounded transformation artifact\ncase=%s\nexecution=%s\ninput=%d\noperation=%s\noutput=%d\nsource=%s\nsemantic-source=%s\nauthorization=%s\nsubject=%s\n",
		receipt.CaseID, receipt.ExecutionID, receipt.Evidence.InputValue, receipt.Evidence.CandidateOperation, receipt.Evidence.CandidateResult,
		receipt.SourceDigest, receipt.SemanticSourceDigest, receipt.AuthorizationDigest, receipt.HeadSHA))
}

func parseFixture(source []byte) (sourceFixture, error) {
	return parseFixtureCase(source, preservedCaseID)
}

func parseFixtureCase(source []byte, caseID string) (sourceFixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceFixture{}, fmt.Errorf("consumer source syntax: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceFixture{}, err
	}
	semantic := "sha256:" + ir.StableHash()
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			return sourceFixture{}, fmt.Errorf("preserved source activity is not executable")
		}
		fields, err := decodeFields(activity.ValueProgram)
		if err != nil {
			return sourceFixture{}, err
		}
		if fields["case"] != caseID {
			continue
		}
		input, err := strconv.ParseInt(fields["input"], 10, 64)
		if err != nil {
			return sourceFixture{}, err
		}
		expected, err := strconv.ParseInt(fields["expected"], 10, 64)
		if err != nil {
			return sourceFixture{}, err
		}
		result, err := evaluateAdd(fields["candidate"], input)
		if err != nil {
			return sourceFixture{}, err
		}
		return sourceFixture{Activity: activity.Name, CaseID: fields["case"], CaseKind: fields["kind"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: result, Expected: expected, Invariant: fields["invariant"], InvariantID: fields["invariant-id"], DomainID: fields["domain"], OperationID: model.Digest([]string{"operation", fields["candidate"]}), ReplayRecipe: fields["replay"], EffectIntent: fields["effect"], SemanticSourceDigest: semantic}, nil
	}
	return sourceFixture{}, fmt.Errorf("preserved source activity is missing")
}

func decodeFields(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 10 {
		return nil, fmt.Errorf("fixture computes value has %d fields, want 10", len(parts))
	}
	fields := map[string]string{}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("fixture field %q is malformed", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("fixture field %q is duplicated", part)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "kind", "input", "candidate", "expected", "invariant", "invariant-id", "domain", "replay", "effect"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("fixture field %q is missing", key)
		}
	}
	return fields, nil
}
func evaluateAdd(operation string, input int64) (int64, error) {
	name, text, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || text == "" || strings.Contains(text, ":") {
		return 0, fmt.Errorf("operation %q unsupported", operation)
	}
	operand, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, err
	}
	const max = int64(1<<63 - 1)
	const min = -max - 1
	if (operand > 0 && input > max-operand) || (operand < 0 && input < min-operand) {
		return 0, fmt.Errorf("operation overflows int64")
	}
	return input + operand, nil
}
func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("REGRESSION_REPLAY_RECIPE_UNAVAILABLE")
	}
	return evaluateAdd(recipe, input)
}

func reconstructReceipt(fixture sourceFixture, source []byte, headSHA string) model.Receipt {
	sourceDigest := model.DigestBytes(source)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	before := model.SemanticDigest(fixture.Input)
	after := model.SemanticDigest(fixture.CandidateResult)
	expected := model.SemanticDigest(fixture.Expected)
	replayOutput, replayErr := executeReplay(fixture.ReplayRecipe, fixture.Input)
	evidence := model.TransformationEvidence{SourceDigest: sourceDigest, SemanticSourceDigest: fixture.SemanticSourceDigest, CaseStableID: fixture.CaseID, ActivityStableID: fixture.Activity, OperationID: fixture.OperationID, InputDomainID: fixture.DomainID, InvariantID: fixture.InvariantID, EffectIntent: fixture.EffectIntent, InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation, CandidateResult: fixture.CandidateResult, ExpectedValue: fixture.Expected, Invariant: fixture.Invariant, CandidateDigest: candidateDigest, SemanticBeforeDigest: before, SemanticAfterDigest: after, ExpectedSemanticDigest: expected, ReplayRecipe: fixture.ReplayRecipe, BaselineInputValue: fixture.Input, BaselineOperation: fixture.CandidateOperation, BaselineOutput: fixture.CandidateResult, BaselineDigest: candidateDigest, ReplayCount: 1}
	if replayErr != nil {
		evidence.ReplayFailureStage, evidence.ReplayFailureStep, evidence.ReplayFailureReason = "REGRESSION", "execute-replay", "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	} else {
		replayDigest := model.CandidateDigest(fixture.ReplayRecipe, fixture.Input, replayOutput)
		evidence.ReplayInputValue, evidence.ReplayOperation, evidence.ReplayOutput = fixture.Input, fixture.ReplayRecipe, replayOutput
		evidence.ReplayDigest, evidence.ReplaySemanticDigest = replayDigest, model.SemanticDigest(replayOutput)
		evidence.ReplayEvidenceDigest = model.ReplayDigest(candidateDigest, replayDigest)
		evidence.ReplayCount = 2
		evidence.RegressionWitnessPresent = candidateDigest == replayDigest && after == evidence.ReplaySemanticDigest
	}
	post := model.PostconditionDigest(before, after, expected)
	statuses := map[string]string{"precondition": model.StatusDischarged, "transformation": model.StatusDischarged, "postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged}
	reasons := map[string]string{"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED", "postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED"}
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"], reasons["postcondition"] = model.StatusRefuted, "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if evidence.ReplayCount != 2 {
		statuses["regression-witness"], reasons["regression-witness"] = model.StatusOpen, evidence.ReplayFailureReason
	} else if !evidence.RegressionWitnessPresent || fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"], reasons["regression-witness"] = model.StatusRefuted, "REGRESSION_REPLAY_REFUTED"
	}
	claims := []model.Claim{}
	values := []model.MetaValue{}
	for _, spec := range model.CanonicalValueSpecs() {
		evidenceDigest := sourceDigest
		if spec.ID == "transformation" {
			evidenceDigest = candidateDigest
		}
		if spec.ID == "postcondition" {
			evidenceDigest = post
		}
		if spec.ID == "regression-witness" {
			evidenceDigest = evidence.ReplayEvidenceDigest
		}
		coordinate := model.Coordinate{Stage: spec.Coordinate.Stage, Step: spec.Coordinate.Step, Reason: reasons[spec.ID]}
		id := fixture.CaseID + "::" + spec.ID
		transition := model.NewTransition(id, model.StatusOpen, statuses[spec.ID], coordinate, evidenceDigest)
		claims = append(claims, model.Claim{ID: id, Status: statuses[spec.ID], Reason: reasons[spec.ID], VerificationCheck: spec.VerificationCheck, Coordinate: coordinate, TargetDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigests: evidenceDigests(evidenceDigest), Transitions: []model.Transition{transition}})
		values = append(values, model.MetaValue{ID: spec.ID, Kind: spec.Kind, Value: statuses[spec.ID], EvidenceDigest: evidenceDigest, Producer: spec.Producer, Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice, VerificationCheck: spec.VerificationCheck, Coordinate: coordinate})
	}
	decision, resolution, reason := derive(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: fixture.CaseID, CaseKind: fixture.CaseKind, ActivityStableID: fixture.Activity, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: sourceDigest, SemanticSourceDigest: fixture.SemanticSourceDigest, ContractDigest: model.ValueContractDigest(), ValidatorContractDigest: model.ValidatorContractDigest(), Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp, ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence, Decision: decision, Resolution: resolution, Reason: reason, Phase: model.ReceiptProvisional, Effects: []model.Effect{}, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryWritesObserved: false, RepositoryWrites: -1, RepositoryMutationAuthorized: false, RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope, AuthorityScope: model.AuthorityScope}
	receipt.AuthorizationDigest = model.AuthorizationDigest(receipt)
	return model.SealReceipt(receipt)
}

func reconstructJudgment(receipt model.Receipt) model.Judgment {
	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects), AuthorizationDigest: receipt.AuthorizationDigest}
	for _, claim := range receipt.Claims {
		switch claim.Status {
		case model.StatusDischarged:
			judgment.DischargedClaims++
		case model.StatusOpen:
			judgment.OpenClaims++
		case model.StatusRefuted:
			judgment.RefutedClaims++
		}
	}
	judgment.Decision, judgment.Resolution, judgment.Reason = derive(receipt.Claims)
	if judgment.Decision == model.DecisionAllowed {
		judgment.Status = model.StatusDischarged
	} else if judgment.Decision == model.DecisionBlocked {
		judgment.Status = model.StatusOpen
	} else {
		judgment.Status = model.StatusRefuted
	}
	return judgment
}
func derive(claims []model.Claim) (string, string, string) {
	for _, claim := range claims {
		if claim.Status == model.StatusRefuted {
			return model.DecisionRefuted, model.ResolutionInvariant, claim.Reason
		}
	}
	for _, claim := range claims {
		if claim.Status == model.StatusOpen {
			return model.DecisionBlocked, model.ResolutionLower, claim.Reason
		}
	}
	return model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED"
}
func evidenceDigests(value string) []string {
	if value == "" {
		return []string{}
	}
	return []string{value}
}
func projection(fixture sourceFixture) projectionWire {
	return projectionWire{Activity: fixture.Activity, CaseID: fixture.CaseID, CaseKind: fixture.CaseKind, Input: fixture.Input, CandidateOperation: fixture.CandidateOperation, CandidateResult: fixture.CandidateResult, Expected: fixture.Expected, Invariant: fixture.Invariant, InvariantID: fixture.InvariantID, DomainID: fixture.DomainID, OperationID: fixture.OperationID, ReplayRecipe: fixture.ReplayRecipe, SemanticSourceDigest: fixture.SemanticSourceDigest, EffectIntent: fixture.EffectIntent}
}
func transitionOutcomes(left, right []model.Claim) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if len(left[index].Transitions) != len(right[index].Transitions) {
			return false
		}
		for step := range left[index].Transitions {
			if left[index].Transitions[step].ClaimID != right[index].Transitions[step].ClaimID || left[index].Transitions[step].From != right[index].Transitions[step].From || left[index].Transitions[step].To != right[index].Transitions[step].To || left[index].Transitions[step].Coordinate != right[index].Transitions[step].Coordinate {
				return false
			}
		}
	}
	return true
}
func replayObservationEqual(left, right model.TransformationEvidence) bool {
	return left.ReplayCount == right.ReplayCount && left.ReplayOperation == right.ReplayOperation && left.ReplayOutput == right.ReplayOutput && left.ReplayDigest == right.ReplayDigest && left.ReplaySemanticDigest == right.ReplaySemanticDigest && left.ReplayEvidenceDigest == right.ReplayEvidenceDigest
}
func transitionWires(claims []model.Claim) []transitionWire {
	result := []transitionWire{}
	for _, claim := range claims {
		for _, transition := range claim.Transitions {
			result = append(result, transitionWire{ClaimID: transition.ClaimID, From: transition.From, To: transition.To, Coordinate: coordinateWire{Stage: transition.Coordinate.Stage, Step: transition.Coordinate.Step, Reason: transition.Coordinate.Reason}, PropositionDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigest: transition.EvidenceDigest, PreviousTransitionDigest: transition.PreviousTransitionDigest, CurrentTransitionDigest: transition.CurrentTransitionDigest})
		}
	}
	return result
}

func transitionFromModel(transition model.Transition) transitionWire {
	return transitionWire{ClaimID: transition.ClaimID, From: transition.From, To: transition.To, Coordinate: coordinateWire{Stage: transition.Coordinate.Stage, Step: transition.Coordinate.Step, Reason: transition.Coordinate.Reason}, PropositionDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest, EvidenceDigest: transition.EvidenceDigest, PreviousTransitionDigest: transition.PreviousTransitionDigest, CurrentTransitionDigest: transition.CurrentTransitionDigest}
}
func mutateSemantic(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=3")
	newValue := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=4")
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("semantic intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}
func mutateOperation(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=3")
	newValue := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:2;expected=3")
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("operation intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}
func mutateNonSemantic(source []byte) ([]byte, error) {
	return append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...), nil
}
func verifyArtifactPath(path string) bool {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	absRoot, _ := filepath.Abs(root)
	absPath, _ := filepath.Abs(path)
	return filepath.Dir(absPath) == absRoot
}
func coherentTamper(report reportWire) reportWire {
	tampered := report
	tampered.Cases = append([]caseWire(nil), report.Cases...)
	tampered.Cases[2].Claim.Status = model.StatusRefuted
	tampered.Cases[2].Satisfied = false
	tampered.Cases[2].Claim.Reason = nonSemanticContradictionReason
	tampered.Denominator.NonSemantic.CasesSatisfied = 0
	tampered.Denominator.NonSemantic.CoverageBPS = 0
	tampered.Decision = "FAIL_CLOSED"
	tampered.Resolution = model.ResolutionInvariant
	tampered.Reason = "tampered"
	return reseal(tampered)
}
func reseal(report reportWire) reportWire {
	report.Digest = ""
	report.Digest = model.Digest(report)
	return report
}
