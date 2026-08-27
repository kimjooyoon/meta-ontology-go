package intervention

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/executor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                         = "gooo/invariant-transformation-intervention-report/v2"
	DenominatorID                  = "gooo/invariant-transformation-intervention-denominator/v2"
	SemanticExpectedDenominatorID  = "gooo/invariant-transformation-intervention-semantic-expected-denominator/v2"
	SemanticOperationDenominatorID = "gooo/invariant-transformation-intervention-semantic-operation-denominator/v2"
	NonSemanticDenominatorID       = "gooo/invariant-transformation-intervention-nonsemantic-denominator/v2"
	SemanticExpectedCaseID         = "semantic-expected-intervention"
	SemanticOperationCaseID        = "semantic-operation-intervention"
	NonSemanticCaseID              = "nonsemantic-source-intervention"
	SemanticExpectedClaimID        = SemanticExpectedCaseID + "::claim"
	SemanticOperationClaimID       = SemanticOperationCaseID + "::claim"
	NonSemanticClaimID             = NonSemanticCaseID + "::claim"
	InterventionStage              = "INTERVENTION"
	SemanticExpectedStep           = "compare-semantic-expected-projection-and-decision"
	SemanticExpectedReason         = "SEMANTIC_EXPECTED_VALUE_AND_DECISION_CHANGED"
	SemanticOperationStep          = "compare-semantic-operation-projection-and-decision"
	SemanticOperationReason        = "SEMANTIC_OPERATION_AND_DECISION_CHANGED"
	NonSemanticStep                = "compare-nonsemantic-projection-and-decision"
	NonSemanticReason              = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
	SemanticContradictionReason    = "SEMANTIC_INTERVENTION_CONTRADICTED"
	NonSemanticContradictionReason = "NONSEMANTIC_INTERVENTION_CONTRADICTED"
	EvidenceUnobservableReason     = "INTERVENTION_EVIDENCE_UNOBSERVABLE"
	FailClosedDecision             = model.DecisionFailClosed
	PreservedCaseID                = "preserved-translation"
	ExpectedSemanticMutation       = "expected=4"
	OriginalSemanticMutation       = "expected=3"
	ExpectedOperationMutation      = "add:2"
	OriginalOperationMutation      = "add:1"
	NonSemanticInterventionLabel   = "comment-and-whitespace-only"
)

type FixtureProjection struct {
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

type Claim struct {
	ID                string             `json:"id"`
	Status            string             `json:"status"`
	Resolution        string             `json:"resolution"`
	Reason            string             `json:"reason"`
	VerificationCheck string             `json:"verification_check"`
	Coordinate        model.Coordinate   `json:"coordinate"`
	TargetDigest      string             `json:"target_digest"`
	PriorStateDigest  string             `json:"prior_state_digest"`
	EvidenceDigest    string             `json:"evidence_digest"`
	Transitions       []model.Transition `json:"transitions"`
}

type SliceDenominator struct {
	ID             string `json:"id"`
	CasesTotal     int    `json:"cases_total"`
	CasesSatisfied int    `json:"cases_satisfied"`
	CoverageBPS    int    `json:"coverage_bps"`
}

type FixedDenominator struct {
	ID                      string           `json:"id"`
	CasesTotal              int              `json:"cases_total"`
	SemanticExpectedChange  SliceDenominator `json:"semantic_expected_change"`
	SemanticOperationChange SliceDenominator `json:"semantic_operation_change"`
	NonSemantic             SliceDenominator `json:"nonsemantic_change"`
}

type Failure struct {
	CaseID string `json:"case_id"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type EffectGateCase struct {
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

type Case struct {
	ID                                        string                       `json:"id"`
	Kind                                      string                       `json:"kind"`
	SourceEdit                                string                       `json:"source_edit"`
	BaselineProjection                        FixtureProjection            `json:"baseline_projection"`
	MutatedProjection                         FixtureProjection            `json:"mutated_projection"`
	BaselineProjectionDigest                  string                       `json:"baseline_projection_digest"`
	MutatedProjectionDigest                   string                       `json:"mutated_projection_digest"`
	BaselineSourceDigest                      string                       `json:"baseline_source_digest"`
	MutatedSourceDigest                       string                       `json:"mutated_source_digest"`
	BaselineProvenanceDigest                  string                       `json:"baseline_provenance_digest"`
	MutatedProvenanceDigest                   string                       `json:"mutated_provenance_digest"`
	ProvenanceDigestChanged                   bool                         `json:"provenance_digest_changed"`
	BaselineSemanticDigest                    string                       `json:"baseline_semantic_digest"`
	MutatedSemanticDigest                     string                       `json:"mutated_semantic_digest"`
	SemanticDigestEqual                       bool                         `json:"semantic_digest_equal"`
	BaselineReceiptDigest                     string                       `json:"baseline_receipt_digest"`
	MutatedReceiptDigest                      string                       `json:"mutated_receipt_digest"`
	BaselineReceiptDecision                   string                       `json:"baseline_receipt_decision"`
	MutatedReceiptDecision                    string                       `json:"mutated_receipt_decision"`
	BaselineJudgment                          model.Judgment               `json:"baseline_judgment"`
	MutatedJudgment                           model.Judgment               `json:"mutated_judgment"`
	BaselineEvidence                          model.TransformationEvidence `json:"baseline_evidence"`
	MutatedEvidence                           model.TransformationEvidence `json:"mutated_evidence"`
	BaselineClaimTransitions                  []model.Transition           `json:"baseline_claim_transitions"`
	MutatedClaimTransitions                   []model.Transition           `json:"mutated_claim_transitions"`
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
	Claim                                     Claim                        `json:"claim"`
	Satisfied                                 bool                         `json:"satisfied"`
}

type Report struct {
	Schema                            string           `json:"schema"`
	HeadSHA                           string           `json:"head_sha"`
	SourcePath                        string           `json:"source_path"`
	SourceDigest                      string           `json:"source_digest"`
	Denominator                       FixedDenominator `json:"denominator"`
	CaseCount                         int              `json:"case_count"`
	Cases                             []Case           `json:"cases"`
	EffectGates                       []EffectGateCase `json:"effect_gates"`
	EffectGateDenominator             int              `json:"effect_gate_denominator"`
	EffectGateSatisfied               int              `json:"effect_gate_satisfied"`
	Decision                          string           `json:"decision"`
	Resolution                        string           `json:"resolution"`
	Reason                            string           `json:"reason"`
	RepositoryWrites                  int              `json:"repository_writes"`
	RepositoryMutationAuthorized      bool             `json:"repository_mutation_authorized"`
	TempArtifactWriteAuthorized       bool             `json:"temp_artifact_write_authorized"`
	RepositoryNetStatusUnchanged      bool             `json:"repository_net_status_unchanged"`
	RepositoryActualOrTransientWrites string           `json:"repository_actual_or_transient_writes"`
	RepositoryNetStatusObserved       bool             `json:"repository_net_status_observed"`
	ExecutedEffects                   int              `json:"executed_effects"`
	IndependentlyObservedEffects      int              `json:"independently_observed_effects"`
	UnknownEffectScopes               int              `json:"unknown_effect_scopes"`
	RepositoryPathAuthorization       bool             `json:"repository_path_authorization"`
	AmbientProcessAuthority           string           `json:"ambient_process_authority"`
	CorrectionCount                   int              `json:"correction_count"`
	CorrectionDenominator             int              `json:"correction_denominator"`
	Failure                           *Failure         `json:"failure,omitempty"`
	Digest                            string           `json:"digest"`
}

func Build(source []byte, headSHA string) (Report, error) {
	if !model.ValidHead(headSHA) {
		return Report{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	if _, err := project(source); err != nil {
		return Report{}, err
	}
	semanticExpectedCase, err := buildCase(source, headSHA, SemanticExpectedCaseID, "SEMANTIC_EXPECTED", "semantic-expected-value-change", mutateSemantic, SemanticExpectedClaimID, SemanticExpectedStep, SemanticExpectedReason)
	if err != nil {
		return Report{}, err
	}
	semanticOperationCase, err := buildCase(source, headSHA, SemanticOperationCaseID, "SEMANTIC_OPERATION", "semantic-candidate-operation-change", mutateOperation, SemanticOperationClaimID, SemanticOperationStep, SemanticOperationReason)
	if err != nil {
		return Report{}, err
	}
	nonSemanticCase, err := buildCase(source, headSHA, NonSemanticCaseID, "NON_SEMANTIC", NonSemanticInterventionLabel, mutateNonSemantic, NonSemanticClaimID, NonSemanticStep, NonSemanticReason)
	if err != nil {
		return Report{}, err
	}
	cases := []Case{semanticExpectedCase, semanticOperationCase, nonSemanticCase}
	gates, err := buildEffectGates(source, headSHA)
	if err != nil {
		return Report{}, err
	}
	decision, resolution, reason, failure := deriveReport(cases, gates)
	writes, mutationAuthority := effectTotals(cases)
	report := Report{Schema: Schema, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: model.DigestBytes(source),
		Denominator: FixedDenominator{ID: DenominatorID, CasesTotal: 3,
			SemanticExpectedChange:  SliceDenominator{ID: SemanticExpectedDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticExpectedCase.Satisfied), CoverageBPS: boolInt(semanticExpectedCase.Satisfied) * 10000},
			SemanticOperationChange: SliceDenominator{ID: SemanticOperationDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticOperationCase.Satisfied), CoverageBPS: boolInt(semanticOperationCase.Satisfied) * 10000},
			NonSemantic:             SliceDenominator{ID: NonSemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(nonSemanticCase.Satisfied), CoverageBPS: boolInt(nonSemanticCase.Satisfied) * 10000}},
		CaseCount: 3, Cases: cases, EffectGates: gates, EffectGateDenominator: len(gates), Decision: decision, Resolution: resolution, Reason: reason,
		RepositoryWrites: writes, RepositoryMutationAuthorized: mutationAuthority, TempArtifactWriteAuthorized: effectGateObserved(gates), RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false,
		RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope,
		ExecutedEffects: boolInt(effectGateObserved(gates)), IndependentlyObservedEffects: boolInt(effectGateObserved(gates)),
		UnknownEffectScopes: boolInt(effectGateObserved(gates)), CorrectionCount: 12, CorrectionDenominator: 12, Failure: failure}
	for _, gate := range gates {
		if gate.Satisfied {
			report.EffectGateSatisfied++
		}
	}
	return seal(report), nil
}

// DeterministicReplay is explicitly a producer replay check, not independent
// evidence. The production consumer is judge.Judge in the separate package.
func DeterministicReplay(report Report, source []byte, headSHA string) error {
	if report.Schema != Schema || report.HeadSHA != headSHA || report.SourcePath != model.SourcePath || report.SourceDigest != model.DigestBytes(source) || report.CaseCount != 3 || report.Digest == "" || report.Digest != seal(report).Digest {
		return fmt.Errorf("intervention report identity or digest mismatch")
	}
	expected, err := Build(source, headSHA)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(report, expected) {
		return fmt.Errorf("intervention report deterministic replay mismatch")
	}
	return nil
}

func deriveReport(cases []Case, gates []EffectGateCase) (string, string, string, *Failure) {
	for _, item := range cases {
		if !item.Satisfied {
			return failFor(item.ID, item.Claim)
		}
	}
	for _, gate := range gates {
		if !gate.Satisfied {
			return FailClosedDecision, model.ResolutionLower, fmt.Sprintf("CASE=%s;STAGE=EFFECT_GATE;STEP=execute-authorized-temp-artifact;REASON=%s", gate.ID, gate.Reason), &Failure{CaseID: gate.ID, Stage: "EFFECT_GATE", Step: "execute-authorized-temp-artifact", Reason: gate.Reason}
		}
	}
	return model.DecisionPass, model.ResolutionExact, "ALL_INTERVENTION_OBSERVATIONS_SATISFIED", nil
}

func failFor(id string, claim Claim) (string, string, string, *Failure) {
	failure := &Failure{CaseID: id, Stage: claim.Coordinate.Stage, Step: claim.Coordinate.Step, Reason: claim.Reason}
	resolution := claim.Resolution
	if resolution == "" {
		resolution = model.ResolutionLower
	}
	return FailClosedDecision, resolution, fmt.Sprintf("CASE=%s;STAGE=%s;STEP=%s;REASON=%s", failure.CaseID, failure.Stage, failure.Step, failure.Reason), failure
}

func effectTotals(cases []Case) (int, bool) {
	writes := -1
	authority := false
	for _, item := range cases {
		if item.BaselineRepositoryWritesObserved || item.MutatedRepositoryWritesObserved {
			if writes < 0 {
				writes = 0
			}
			writes += item.BaselineRepositoryWrites + item.MutatedRepositoryWrites
		}
		authority = authority || item.BaselineRepositoryMutationAuthorized || item.MutatedRepositoryMutationAuthorized
	}
	return writes, authority
}

func buildCase(source []byte, headSHA, id, kind, edit string, mutate func([]byte) ([]byte, error), claimID, step, satisfiedReason string) (Case, error) {
	mutated, err := mutate(source)
	if err != nil {
		return Case{}, err
	}
	baselineProjection, err := project(source)
	if err != nil {
		return Case{}, err
	}
	mutatedProjection, err := project(mutated)
	if err != nil {
		return Case{}, err
	}
	baselineReceipt, err := producer.Build(source, headSHA, PreservedCaseID)
	if err != nil {
		return Case{}, err
	}
	mutatedReceipt, err := producer.Build(mutated, headSHA, PreservedCaseID)
	if err != nil {
		return Case{}, err
	}
	baselineJudgment := judge.Judge(baselineReceipt, source)
	mutatedJudgment := judge.Judge(mutatedReceipt, mutated)
	baselineTransitions, mutatedTransitions := transitions(baselineReceipt), transitions(mutatedReceipt)
	item := Case{ID: id, Kind: kind, SourceEdit: edit, BaselineProjection: baselineProjection, MutatedProjection: mutatedProjection,
		BaselineProjectionDigest: model.Digest(baselineProjection), MutatedProjectionDigest: model.Digest(mutatedProjection), BaselineSourceDigest: baselineReceipt.SourceDigest, MutatedSourceDigest: mutatedReceipt.SourceDigest,
		BaselineProvenanceDigest: provenanceDigest(baselineReceipt.SourceDigest, headSHA, id), MutatedProvenanceDigest: provenanceDigest(mutatedReceipt.SourceDigest, headSHA, id), ProvenanceDigestChanged: baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest,
		BaselineSemanticDigest: baselineProjection.SemanticSourceDigest, MutatedSemanticDigest: mutatedProjection.SemanticSourceDigest, SemanticDigestEqual: baselineProjection.SemanticSourceDigest == mutatedProjection.SemanticSourceDigest,
		BaselineReceiptDigest: baselineReceipt.Digest, MutatedReceiptDigest: mutatedReceipt.Digest, BaselineReceiptDecision: baselineReceipt.Decision, MutatedReceiptDecision: mutatedReceipt.Decision,
		BaselineJudgment: baselineJudgment, MutatedJudgment: mutatedJudgment, BaselineClaimTransitions: baselineTransitions, MutatedClaimTransitions: mutatedTransitions,
		RawSourceDigestChanged: baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest, ReceiptChanged: baselineReceipt.Digest != mutatedReceipt.Digest,
		SemanticProjectionEqual: reflect.DeepEqual(baselineProjection, mutatedProjection), DecisionEqual: baselineJudgment.Decision == mutatedJudgment.Decision && baselineReceipt.Decision == mutatedReceipt.Decision,
		ResolutionEqual: baselineJudgment.Resolution == mutatedJudgment.Resolution && baselineReceipt.Resolution == mutatedReceipt.Resolution, ReasonEqual: baselineJudgment.Reason == mutatedJudgment.Reason && baselineReceipt.Reason == mutatedReceipt.Reason,
		EffectsEqual: reflect.DeepEqual(baselineReceipt.Effects, mutatedReceipt.Effects), ReplayObservationEqual: replayEqual(baselineReceipt.Evidence, mutatedReceipt.Evidence), EvidenceObservable: baselineJudgment.Independent && mutatedJudgment.Independent,
		BaselineEvidence: baselineReceipt.Evidence, MutatedEvidence: mutatedReceipt.Evidence, BaselineRepositoryWrites: baselineReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites,
		BaselineRepositoryWritesObserved: baselineReceipt.RepositoryWritesObserved, MutatedRepositoryWritesObserved: mutatedReceipt.RepositoryWritesObserved,
		BaselineRepositoryNetStatusUnchanged: baselineReceipt.RepositoryNetStatusUnchanged, MutatedRepositoryNetStatusUnchanged: mutatedReceipt.RepositoryNetStatusUnchanged,
		BaselineRepositoryActualOrTransientWrites: baselineReceipt.RepositoryActualOrTransientWrites, MutatedRepositoryActualOrTransientWrites: mutatedReceipt.RepositoryActualOrTransientWrites,
		BaselineRepositoryMutationAuthorized: baselineReceipt.RepositoryMutationAuthorized, MutatedRepositoryMutationAuthorized: mutatedReceipt.RepositoryMutationAuthorized}
	item.DecisionChanged = !item.DecisionEqual
	item.ClaimTransitionsEqual = transitionOutcomeEqual(baselineTransitions, mutatedTransitions)
	item.RepositoryWritesNotClaimed = !item.BaselineRepositoryWritesObserved && !item.MutatedRepositoryWritesObserved && item.BaselineRepositoryWrites == -1 && item.MutatedRepositoryWrites == -1 && item.BaselineRepositoryActualOrTransientWrites == model.UnknownEffectScope && item.MutatedRepositoryActualOrTransientWrites == model.UnknownEffectScope && !item.BaselineRepositoryMutationAuthorized && !item.MutatedRepositoryMutationAuthorized
	item.Satisfied, item.Claim.Resolution, item.Claim.Reason = adjudicate(kind, item, item.EvidenceObservable, satisfiedReason)
	status := statusForAdjudication(item.Satisfied, item.EvidenceObservable)
	coordinate := model.Coordinate{Stage: InterventionStage, Step: step, Reason: item.Claim.Reason}
	transitionEvidence := model.Digest([]any{item.BaselineProjectionDigest, item.MutatedProjectionDigest, item.BaselineJudgment.Decision, item.MutatedJudgment.Decision, item.BaselineJudgment.Resolution, item.MutatedJudgment.Resolution, item.BaselineJudgment.Reason, item.MutatedJudgment.Reason})
	transition := model.NewTransition(claimID, model.StatusOpen, status, coordinate, transitionEvidence)
	item.Claim.ID, item.Claim.Status, item.Claim.Coordinate, item.Claim.VerificationCheck = claimID, status, coordinate, "intervention-observation-derived-from-two-independent-receipts"
	item.Claim.TargetDigest, item.Claim.PriorStateDigest, item.Claim.EvidenceDigest = transition.PropositionDigest, transition.PriorStateDigest, transition.EvidenceDigest
	item.Claim.Transitions = []model.Transition{transition}
	return item, nil
}

func adjudicate(kind string, item Case, observable bool, satisfiedReason string) (bool, string, string) {
	if !observable {
		return false, model.ResolutionLower, EvidenceUnobservableReason
	}
	if observationSatisfied(kind, item) {
		return true, model.ResolutionExact, satisfiedReason
	}
	if kind == "SEMANTIC_EXPECTED" || kind == "SEMANTIC_OPERATION" {
		return false, model.ResolutionInvariant, SemanticContradictionReason
	}
	return false, model.ResolutionInvariant, NonSemanticContradictionReason
}

func observationSatisfied(kind string, item Case) bool {
	switch kind {
	case "SEMANTIC_EXPECTED", "SEMANTIC_OPERATION":
		return item.RawSourceDigestChanged && item.ReceiptChanged && !item.SemanticDigestEqual && !item.SemanticProjectionEqual && !item.DecisionEqual && !item.ResolutionEqual && !item.ReasonEqual && !item.ClaimTransitionsEqual && item.RepositoryWritesNotClaimed && item.BaselineJudgment.Decision == model.DecisionAllowed && item.MutatedJudgment.Decision == model.DecisionRefuted && item.MutatedJudgment.Reason == "SEMANTIC_POSTCONDITION_REFUTED"
	case "NON_SEMANTIC":
		return item.RawSourceDigestChanged && item.ProvenanceDigestChanged && item.ReceiptChanged && item.SemanticDigestEqual && item.SemanticProjectionEqual && item.DecisionEqual && item.ResolutionEqual && item.ReasonEqual && item.ClaimTransitionsEqual && item.EffectsEqual && item.ReplayObservationEqual && item.RepositoryWritesNotClaimed && item.BaselineJudgment.Decision == model.DecisionAllowed && item.MutatedJudgment.Decision == model.DecisionAllowed
	default:
		return false
	}
}

func provenanceDigest(sourceDigest, headSHA, caseID string) string {
	return model.Digest([]string{"invariant-transformation-source-provenance", sourceDigest, headSHA, caseID})
}

func replayEqual(left, right model.TransformationEvidence) bool {
	return left.ReplayCount == right.ReplayCount && left.ReplayOperation == right.ReplayOperation && left.ReplayOutput == right.ReplayOutput && left.ReplayDigest == right.ReplayDigest && left.ReplaySemanticDigest == right.ReplaySemanticDigest && left.ReplayEvidenceDigest == right.ReplayEvidenceDigest
}

func transitionOutcomeEqual(left, right []model.Transition) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ClaimID != right[index].ClaimID || left[index].From != right[index].From || left[index].To != right[index].To || left[index].Coordinate != right[index].Coordinate {
			return false
		}
	}
	return true
}

func statusForAdjudication(satisfied, observable bool) string {
	if satisfied {
		return model.StatusDischarged
	}
	if observable {
		return model.StatusRefuted
	}
	return model.StatusOpen
}

func project(source []byte) (FixtureProjection, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return FixtureProjection{}, fmt.Errorf("parse intervention source: %s", diagnostics.Error())
	}
	semanticDigest, err := canonicalSemanticDigest(source)
	if err != nil {
		return FixtureProjection{}, err
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != "PreservedTranslation" {
			continue
		}
		if len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			return FixtureProjection{}, fmt.Errorf("preserved fixture is not executable")
		}
		fields, err := decode(activity.ValueProgram)
		if err != nil {
			return FixtureProjection{}, err
		}
		input, err := strconv.ParseInt(fields["input"], 10, 64)
		if err != nil {
			return FixtureProjection{}, err
		}
		expected, err := strconv.ParseInt(fields["expected"], 10, 64)
		if err != nil {
			return FixtureProjection{}, err
		}
		candidateResult, err := evaluateAdd(fields["candidate"], input)
		if err != nil {
			return FixtureProjection{}, err
		}
		if fields["case"] != PreservedCaseID || fields["invariant"] != "candidate-output-equals-expected" || fields["invariant-id"] != model.InvariantID || fields["domain"] != model.InputDomainID {
			return FixtureProjection{}, fmt.Errorf("preserved fixture is outside the bounded projection")
		}
		return FixtureProjection{Activity: activity.Name, CaseID: fields["case"], CaseKind: fields["kind"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult, Expected: expected, Invariant: fields["invariant"], InvariantID: fields["invariant-id"], DomainID: fields["domain"], OperationID: model.Digest([]string{"operation", fields["candidate"]}), ReplayRecipe: fields["replay"], SemanticSourceDigest: semanticDigest, EffectIntent: fields["effect"]}, nil
	}
	return FixtureProjection{}, fmt.Errorf("preserved fixture activity is missing")
}

func canonicalSemanticDigest(source []byte) (string, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return "", fmt.Errorf("canonical semantic parse: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("canonical semantic lowering: %w", err)
	}
	return "sha256:" + ir.StableHash(), nil
}

func decode(program string) (map[string]string, error) {
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
			return nil, fmt.Errorf("fixture field %q is duplicated", key)
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
	name, operandText, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || operandText == "" || strings.Contains(operandText, ":") {
		return 0, fmt.Errorf("candidate operation %q is unsupported", operation)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return 0, err
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("candidate operation %q overflows int64", operation)
	}
	return input + operand, nil
}

func mutateSemantic(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;" + OriginalSemanticMutation)
	replacement := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;" + ExpectedSemanticMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("semantic intervention target count is not 1")
	}
	return bytes.Replace(source, old, replacement, 1), nil
}
func mutateOperation(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=" + OriginalOperationMutation + ";expected=3")
	replacement := []byte("case=preserved-translation;kind=PRESERVED;input=2;candidate=" + ExpectedOperationMutation + ";expected=3")
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("operation intervention target count is not 1")
	}
	return bytes.Replace(source, old, replacement, 1), nil
}
func mutateNonSemantic(source []byte) ([]byte, error) {
	return append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...), nil
}

func transitions(receipt model.Receipt) []model.Transition {
	result := []model.Transition{}
	for _, claim := range receipt.Claims {
		result = append(result, claim.Transitions...)
	}
	return result
}

func buildEffectGates(source []byte, headSHA string) ([]EffectGateCase, error) {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	validReceipt, err := producer.Build(source, headSHA, "approved-artifact")
	if err != nil {
		return nil, err
	}
	validAuth := judge.Judge(validReceipt, source)
	validPath := executor.Path(root, "effect-gate-valid-"+headSHA[:8])
	validEffect, validErr := executor.Emit(validReceipt, validAuth, headSHA, validPath)
	validObserved, observeErr := model.ArtifactEvidence{}, validErr
	if validErr == nil {
		validObserved, observeErr = executor.Observe(validEffect)
	}
	valid := EffectGateCase{ID: "effect-valid-authorized", Scenario: "valid-authorized-case", CaseID: validReceipt.CaseID, SubjectSHA: headSHA, Stage: "EFFECT_EXECUTION", Step: "write-authorized-temp-artifact", AttemptPath: validPath, AuthorizationAttempted: true, AuthorizationAccepted: validAuth.Decision == model.DecisionAllowed && validAuth.Independent, ExecutorAccepted: validErr == nil, ArtifactCount: boolInt(validErr == nil && observeErr == nil), ArtifactExists: validErr == nil && observeErr == nil, Artifact: validObserved, Reason: "AUTHORIZED_TEMP_ARTIFACT_OBSERVED"}
	valid.Satisfied = valid.AuthorizationAccepted && valid.ExecutorAccepted && valid.ArtifactCount == 1 && valid.ArtifactExists
	if !valid.Satisfied {
		valid.Reason = "AUTHORIZED_ARTIFACT_OBSERVATION_FAILED"
	}
	makeRejected := func(id, scenario, caseID string, receipt model.Receipt, judgment model.Judgment, subject, reason string) EffectGateCase {
		path := executor.Path(root, "effect-gate-"+id+"-"+headSHA[:8])
		beforeExists, beforeDigest := snapshotTarget(path)
		_, emitErr := executor.Emit(receipt, judgment, subject, path)
		afterExists, afterDigest := snapshotTarget(path)
		return EffectGateCase{ID: id, Scenario: scenario, CaseID: caseID, SubjectSHA: subject, Stage: "EFFECT_AUTHORIZATION", Step: "validate-authorization", AttemptPath: path, TargetBeforeExists: beforeExists, TargetAfterExists: afterExists, TargetBeforeDigest: beforeDigest, TargetAfterDigest: afterDigest, TargetBytesUnchanged: beforeExists == afterExists && beforeDigest == afterDigest, AuthorizationAttempted: true, AuthorizationAccepted: judgment.Independent && judgment.Decision == model.DecisionAllowed && judgment.AuthorizationDigest == receipt.AuthorizationDigest, ExecutorAccepted: emitErr == nil, ArtifactCount: boolInt(afterExists), ArtifactExists: afterExists, Reason: reason, Satisfied: emitErr != nil && !afterExists}
	}
	refutedReceipt, err := producer.Build(source, headSHA, "semantic-violation")
	if err != nil {
		return nil, err
	}
	refutedJudgment := judge.Judge(refutedReceipt, source)
	unauthorizedJudgment := refutedJudgment
	unauthorizedJudgment.Decision = "UNAUTHORIZED"
	unauthorizedJudgment.AuthorizationDigest = refutedReceipt.AuthorizationDigest
	unauthorized := makeRejected("effect-unauthorized", "unauthorized-decision", refutedReceipt.CaseID, refutedReceipt, unauthorizedJudgment, headSHA, "UNAUTHORIZED_DECISION_REJECTED")
	refuted := makeRejected("effect-refuted", "refuted-receipt", refutedReceipt.CaseID, refutedReceipt, refutedJudgment, headSHA, "REFUTED_RECEIPT_REJECTED")
	openReceipt, err := producer.Build(source, headSHA, "missing-regression-witness")
	if err != nil {
		return nil, err
	}
	openJudgment := judge.Judge(openReceipt, source)
	open := makeRejected("effect-open", "open-replay-evidence", openReceipt.CaseID, openReceipt, openJudgment, headSHA, "OPEN_RECEIPT_REJECTED")
	staleReceipt, err := producer.Build(source, strings.Repeat("0", 40), "approved-artifact")
	if err != nil {
		return nil, err
	}
	staleJudgment := judge.Judge(staleReceipt, source)
	stale := makeRejected("effect-stale-sha", "stale-subject-sha", staleReceipt.CaseID, staleReceipt, staleJudgment, headSHA, "STALE_SUBJECT_SHA_REJECTED")
	tamperedAuth := validAuth
	tamperedAuth.AuthorizationDigest = model.Digest([]string{"tampered-authorization"})
	tampered := makeRejected("effect-tampered-auth", "tampered-authorization", validReceipt.CaseID, validReceipt, tamperedAuth, headSHA, "TAMPERED_AUTHORIZATION_REJECTED")
	repository, ok := repositoryRoot()
	if !ok {
		return nil, fmt.Errorf("effect gate repository root is unavailable")
	}
	repositoryTarget := filepath.Join(repository, ".ci-tmp", "invariant-transformation-effect-gate-"+headSHA[:8]+"-repository.bin")
	repositoryBeforeExists, repositoryBeforeDigest := snapshotTarget(repositoryTarget)
	_, repositoryErr := executor.Emit(validReceipt, validAuth, headSHA, repositoryTarget)
	repositoryAfterExists, repositoryAfterDigest := snapshotTarget(repositoryTarget)
	repositoryGate := EffectGateCase{ID: "effect-valid-repository-path", Scenario: "valid-auth-repository-path", CaseID: validReceipt.CaseID, SubjectSHA: headSHA, Stage: "EFFECT_AUTHORIZATION", Step: "validate-temp-root-containment", AttemptPath: repositoryTarget, TargetPath: repositoryTarget, TargetBeforeExists: repositoryBeforeExists, TargetAfterExists: repositoryAfterExists, TargetBeforeDigest: repositoryBeforeDigest, TargetAfterDigest: repositoryAfterDigest, TargetBytesUnchanged: repositoryBeforeExists == repositoryAfterExists && repositoryBeforeDigest == repositoryAfterDigest, AuthorizationAttempted: true, AuthorizationAccepted: validAuth.Decision == model.DecisionAllowed && validAuth.Independent, ExecutorAccepted: repositoryErr == nil, Reason: "REPOSITORY_TARGET_REJECTED"}
	repositoryGate.ArtifactCount = boolInt(repositoryAfterExists)
	repositoryGate.ArtifactExists = repositoryAfterExists
	repositoryGate.Satisfied = repositoryErr != nil && !repositoryAfterExists && repositoryGate.TargetBytesUnchanged
	if !repositoryGate.Satisfied {
		repositoryGate.Reason = "REPOSITORY_TARGET_MUTATION_OBSERVED"
	}
	symlinkPath := executor.Path(root, "effect-gate-symlink-"+headSHA[:8])
	_ = os.Remove(symlinkPath)
	symlinkErr := os.Symlink(repositoryTarget, symlinkPath)
	symlinkBeforeExists, symlinkBeforeDigest := snapshotTarget(repositoryTarget)
	_, escapeErr := executor.Emit(validReceipt, validAuth, headSHA, symlinkPath)
	symlinkAfterExists, symlinkAfterDigest := snapshotTarget(repositoryTarget)
	symlinkGate := EffectGateCase{ID: "effect-valid-temp-symlink", Scenario: "valid-auth-temp-symlink-escape", CaseID: validReceipt.CaseID, SubjectSHA: headSHA, Stage: "EFFECT_AUTHORIZATION", Step: "validate-rooted-target", AttemptPath: symlinkPath, TargetPath: repositoryTarget, TargetBeforeExists: symlinkBeforeExists, TargetAfterExists: symlinkAfterExists, TargetBeforeDigest: symlinkBeforeDigest, TargetAfterDigest: symlinkAfterDigest, TargetBytesUnchanged: symlinkBeforeExists == symlinkAfterExists && symlinkBeforeDigest == symlinkAfterDigest, AuthorizationAttempted: true, AuthorizationAccepted: validAuth.Decision == model.DecisionAllowed && validAuth.Independent, ExecutorAccepted: escapeErr == nil, Reason: "TEMP_SYMLINK_ESCAPE_REJECTED"}
	symlinkGate.ArtifactCount = boolInt(regularArtifactExists(symlinkPath))
	symlinkGate.ArtifactExists = regularArtifactExists(symlinkPath)
	symlinkGate.Satisfied = symlinkErr == nil && escapeErr != nil && !symlinkGate.ArtifactExists && symlinkGate.TargetBytesUnchanged
	if !symlinkGate.Satisfied {
		symlinkGate.Reason = "TEMP_SYMLINK_ESCAPE_MUTATION_OBSERVED"
	}
	return []EffectGateCase{unauthorized, refuted, open, stale, tampered, repositoryGate, symlinkGate, valid}, nil
}

func snapshotTarget(path string) (bool, string) {
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

func regularArtifactExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
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

func effectGateObserved(gates []EffectGateCase) bool {
	for _, gate := range gates {
		if gate.ID == "effect-valid-authorized" {
			return gate.Satisfied
		}
	}
	return false
}
func seal(report Report) Report {
	report.Digest = ""
	report.Digest = model.Digest(report)
	return report
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
