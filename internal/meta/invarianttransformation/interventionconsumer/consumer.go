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
	reportSchema                   = "gooo/invariant-transformation-intervention-report/v1"
	consumerSchema                 = "gooo/invariant-transformation-intervention-consumer/v1"
	denominatorID                  = "gooo/invariant-transformation-intervention-denominator/v1"
	semanticExpectedDenominatorID  = "gooo/invariant-transformation-intervention-semantic-expected-denominator/v1"
	semanticOperationDenominatorID = "gooo/invariant-transformation-intervention-semantic-operation-denominator/v1"
	nonSemanticDenominatorID       = "gooo/invariant-transformation-intervention-nonsemantic-denominator/v1"
	semanticExpectedCaseID         = "semantic-expected-intervention"
	semanticOperationCaseID        = "semantic-operation-intervention"
	nonSemanticCaseID              = "nonsemantic-source-intervention"
	semanticExpectedClaimID        = "semantic-expected-intervention-claim"
	semanticOperationClaimID       = "semantic-operation-intervention-claim"
	nonSemanticClaimID             = "nonsemantic-intervention-claim"
	interventionStage              = "INTERVENTION"
	semanticExpectedStep           = "compare-semantic-expected-projection-and-decision"
	semanticExpectedReason         = "SEMANTIC_EXPECTED_VALUE_AND_DECISION_CHANGED"
	semanticOperationStep          = "compare-semantic-operation-projection-and-decision"
	semanticOperationReason        = "SEMANTIC_OPERATION_AND_DECISION_CHANGED"
	nonSemanticStep                = "compare-nonsemantic-projection-and-decision"
	nonSemanticReason              = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
	semanticContradictionReason    = "SEMANTIC_INTERVENTION_CONTRADICTED"
	nonSemanticContradictionReason = "NONSEMANTIC_INTERVENTION_CONTRADICTED"
	evidenceUnobservableReason     = "INTERVENTION_EVIDENCE_UNOBSERVABLE"
	failClosedDecision             = "FAIL_CLOSED"
	preservedCaseID                = "preserved-translation"
	expectedSemanticMutation       = "expected=4"
	originalSemanticMutation       = "expected=3"
	expectedOperationMutation      = "add:2"
	originalOperationMutation      = "add:1"
	nonSemanticInterventionLabel   = "comment-and-whitespace-only"
	replayUnavailableReason        = "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	replayExecutionReason          = "REGRESSION_REPLAY_EXECUTION_FAILED"
	replayMismatchReason           = "REGRESSION_REPLAY_MISMATCH"
)

// DependencyBoundary is production-only evidence supplied by CI. This
// package consumes it as data and does not import the producer package.
type DependencyBoundary struct {
	ProducerDependencyImports        int `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int `json:"allowed_producer_dependency_imports"`
	ArtifactObservation              int `json:"artifact_observation"`
	ExpectedArtifactObservation      int `json:"expected_artifact_observation"`
}

type Audit struct {
	Schema                           string `json:"schema"`
	HeadSHA                          string `json:"head_sha"`
	ProducerDependencyImports        int    `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int    `json:"allowed_producer_dependency_imports"`
	ReconstructedCases               int    `json:"reconstructed_cases"`
	ExpectedCases                    int    `json:"expected_cases"`
	ActualReplay                     int    `json:"actual_replay"`
	ExpectedActualReplay             int    `json:"expected_actual_replay"`
	ArtifactObservation              int    `json:"artifact_observation"`
	ExpectedArtifactObservation      int    `json:"expected_artifact_observation"`
	CoherentTamperRejected           int    `json:"coherent_tamper_rejected"`
	ExpectedCoherentTamperRejections int    `json:"expected_coherent_tamper_rejections"`
	Decision                         string `json:"decision"`
	Resolution                       string `json:"resolution"`
	Reason                           string `json:"reason"`
	RepositoryWrites                 int    `json:"repository_writes"`
	MutationAuthority                bool   `json:"mutation_authority"`
	Digest                           string `json:"digest"`
}

type coordinateWire struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type transitionWire struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Coordinate coordinateWire `json:"coordinate"`
}

type claimWire struct {
	ID          string           `json:"id"`
	Status      string           `json:"status"`
	Resolution  string           `json:"resolution"`
	Reason      string           `json:"reason"`
	Coordinate  coordinateWire   `json:"coordinate"`
	Transitions []transitionWire `json:"transitions"`
}

type projectionWire struct {
	Activity             string `json:"activity"`
	CaseID               string `json:"case_id"`
	Input                int64  `json:"input"`
	CandidateOperation   string `json:"candidate_operation"`
	CandidateResult      int64  `json:"candidate_result"`
	Expected             int64  `json:"expected"`
	Invariant            string `json:"invariant"`
	ReplayRecipe         string `json:"replay_recipe"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	ApprovedArtifact     bool   `json:"approved_artifact"`
}

type evidenceWire struct {
	SourceDigest             string `json:"source_digest"`
	InputValue               int64  `json:"input_value"`
	CandidateOperation       string `json:"candidate_operation"`
	CandidateResult          int64  `json:"candidate_result"`
	ExpectedValue            int64  `json:"expected_value"`
	Invariant                string `json:"invariant"`
	CandidateDigest          string `json:"candidate_digest"`
	SemanticBeforeDigest     string `json:"semantic_before_digest"`
	SemanticAfterDigest      string `json:"semantic_after_digest"`
	ExpectedSemanticDigest   string `json:"expected_semantic_digest"`
	ReplayRecipe             string `json:"replay_recipe"`
	BaselineInputValue       int64  `json:"baseline_input_value"`
	BaselineOperation        string `json:"baseline_operation"`
	BaselineOutput           int64  `json:"baseline_output"`
	BaselineDigest           string `json:"baseline_digest"`
	ReplayInputValue         int64  `json:"replay_input_value"`
	ReplayOperation          string `json:"replay_operation"`
	ReplayOutput             int64  `json:"replay_output"`
	ReplayDigest             string `json:"replay_digest,omitempty"`
	ReplaySemanticDigest     string `json:"replay_semantic_digest,omitempty"`
	ReplayEvidenceDigest     string `json:"replay_evidence_digest,omitempty"`
	RegressionWitnessPresent bool   `json:"regression_witness_present"`
	ReplayCount              int    `json:"replay_count"`
	ReplayFailureStage       string `json:"replay_failure_stage,omitempty"`
	ReplayFailureStep        string `json:"replay_failure_step,omitempty"`
	ReplayFailureReason      string `json:"replay_failure_reason,omitempty"`
	SemanticSourceDigest     string `json:"semantic_source_digest"`
}

type judgmentWire struct {
	Decision         string `json:"decision"`
	Resolution       string `json:"resolution"`
	Reason           string `json:"reason"`
	Status           string `json:"status"`
	Independent      bool   `json:"independent"`
	CheckedClaims    int    `json:"checked_claims"`
	DischargedClaims int    `json:"discharged_claims"`
	OpenClaims       int    `json:"open_claims"`
	RefutedClaims    int    `json:"refuted_claims"`
	Effects          int    `json:"effects"`
}

type sliceDenominatorWire struct {
	ID             string `json:"id"`
	CasesTotal     int    `json:"cases_total"`
	CasesSatisfied int    `json:"cases_satisfied"`
	CoverageBPS    int    `json:"coverage_bps"`
}

type fixedDenominatorWire struct {
	ID                      string               `json:"id"`
	CasesTotal              int                  `json:"cases_total"`
	SemanticExpectedChange  sliceDenominatorWire `json:"semantic_expected_change"`
	SemanticOperationChange sliceDenominatorWire `json:"semantic_operation_change"`
	NonSemantic             sliceDenominatorWire `json:"nonsemantic_change"`
}

type failureWire struct {
	CaseID string `json:"case_id"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type caseWire struct {
	ID                        string           `json:"id"`
	Kind                      string           `json:"kind"`
	SourceEdit                string           `json:"source_edit"`
	BaselineProjection        projectionWire   `json:"baseline_projection"`
	MutatedProjection         projectionWire   `json:"mutated_projection"`
	BaselineProjectionDigest  string           `json:"baseline_projection_digest"`
	MutatedProjectionDigest   string           `json:"mutated_projection_digest"`
	BaselineSourceDigest      string           `json:"baseline_source_digest"`
	MutatedSourceDigest       string           `json:"mutated_source_digest"`
	BaselineReceiptDigest     string           `json:"baseline_receipt_digest"`
	MutatedReceiptDigest      string           `json:"mutated_receipt_digest"`
	BaselineReceiptDecision   string           `json:"baseline_receipt_decision"`
	MutatedReceiptDecision    string           `json:"mutated_receipt_decision"`
	BaselineJudgment          judgmentWire     `json:"baseline_judgment"`
	MutatedJudgment           judgmentWire     `json:"mutated_judgment"`
	BaselineEvidence          evidenceWire     `json:"baseline_evidence"`
	MutatedEvidence           evidenceWire     `json:"mutated_evidence"`
	BaselineClaimTransitions  []transitionWire `json:"baseline_claim_transitions"`
	MutatedClaimTransitions   []transitionWire `json:"mutated_claim_transitions"`
	RawSourceDigestChanged    bool             `json:"raw_source_digest_changed"`
	ReceiptChanged            bool             `json:"receipt_changed"`
	SemanticProjectionEqual   bool             `json:"semantic_projection_equal"`
	DecisionEqual             bool             `json:"decision_equal"`
	ResolutionEqual           bool             `json:"resolution_equal"`
	ReasonEqual               bool             `json:"reason_equal"`
	DecisionChanged           bool             `json:"decision_changed"`
	ClaimTransitionsEqual     bool             `json:"claim_transitions_equal"`
	EffectsEqual              bool             `json:"effects_equal"`
	ReplayObservationEqual    bool             `json:"replay_observation_equal"`
	EvidenceObservable        bool             `json:"evidence_observable"`
	RepositoryWritesZero      bool             `json:"repository_writes_zero"`
	BaselineRepositoryWrites  int              `json:"baseline_repository_writes"`
	MutatedRepositoryWrites   int              `json:"mutated_repository_writes"`
	BaselineMutationAuthority bool             `json:"baseline_mutation_authority"`
	MutatedMutationAuthority  bool             `json:"mutated_mutation_authority"`
	Claim                     claimWire        `json:"claim"`
	Satisfied                 bool             `json:"satisfied"`
}

type reportWire struct {
	Schema            string               `json:"schema"`
	HeadSHA           string               `json:"head_sha"`
	SourcePath        string               `json:"source_path"`
	SourceDigest      string               `json:"source_digest"`
	Denominator       fixedDenominatorWire `json:"denominator"`
	CaseCount         int                  `json:"case_count"`
	Cases             []caseWire           `json:"cases"`
	Decision          string               `json:"decision"`
	Resolution        string               `json:"resolution"`
	Reason            string               `json:"reason"`
	RepositoryWrites  int                  `json:"repository_writes"`
	MutationAuthority bool                 `json:"mutation_authority"`
	Failure           *failureWire         `json:"failure,omitempty"`
	Digest            string               `json:"digest"`
}

type sourceFixture struct {
	Activity             string
	CaseID               string
	Input                int64
	CandidateOperation   string
	CandidateResult      int64
	Expected             int64
	Invariant            string
	ReplayRecipe         string
	SemanticSourceDigest string
	ApprovedArtifact     bool
}

func VerifyReport(raw, source []byte, headSHA string, dependency DependencyBoundary) (Audit, error) {
	if dependency.ProducerDependencyImports != 0 || dependency.AllowedProducerDependencyImports != 0 {
		return Audit{}, fmt.Errorf("consumer producer dependency boundary is not 0/0")
	}
	if dependency.ArtifactObservation != dependency.ExpectedArtifactObservation || dependency.ArtifactObservation != 1 {
		return Audit{}, fmt.Errorf("approved artifact observation is not 1/1")
	}
	observed, err := decodeReport(raw)
	if err != nil {
		return Audit{}, err
	}
	expected, err := reconstructExpected(source, headSHA)
	if err != nil {
		return Audit{}, err
	}
	if !reflect.DeepEqual(observed, expected) {
		return Audit{}, fmt.Errorf("consumer rejected source-incoherent intervention report")
	}
	tampered := coherentTamper(observed)
	coherentTamperRejected := 0
	if !reflect.DeepEqual(tampered, expected) {
		coherentTamperRejected = 1
	}
	if coherentTamperRejected != 1 {
		return Audit{}, fmt.Errorf("consumer failed coherent resealed tamper regression")
	}
	actualReplay := 0
	for _, item := range expected.Cases {
		if item.Kind != "NON_SEMANTIC" && item.BaselineEvidence.ReplayCount == 2 && item.MutatedEvidence.ReplayCount == 2 {
			actualReplay++
		}
	}
	if actualReplay != 2 {
		return Audit{}, fmt.Errorf("actual replay is %d/2", actualReplay)
	}
	audit := Audit{Schema: consumerSchema, HeadSHA: headSHA,
		ProducerDependencyImports: dependency.ProducerDependencyImports, AllowedProducerDependencyImports: dependency.AllowedProducerDependencyImports,
		ReconstructedCases: len(expected.Cases), ExpectedCases: 3, ActualReplay: actualReplay, ExpectedActualReplay: 2,
		ArtifactObservation: dependency.ArtifactObservation, ExpectedArtifactObservation: dependency.ExpectedArtifactObservation,
		CoherentTamperRejected: coherentTamperRejected, ExpectedCoherentTamperRejections: 1, Decision: expected.Decision,
		Resolution: expected.Resolution, Reason: expected.Reason, RepositoryWrites: expected.RepositoryWrites, MutationAuthority: expected.MutationAuthority}
	audit.Digest = sealAudit(audit)
	return audit, nil
}

func decodeReport(raw []byte) (reportWire, error) {
	var report reportWire
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return reportWire{}, fmt.Errorf("decode intervention report: %w", err)
	}
	if report.Schema != reportSchema || !model.ValidHead(report.HeadSHA) || report.SourcePath != model.SourcePath ||
		!model.ValidDigest(report.SourceDigest) || report.Digest == "" || report.Digest != sealReport(report).Digest {
		return reportWire{}, fmt.Errorf("intervention report identity or digest is invalid")
	}
	return report, nil
}

func reconstructExpected(source []byte, headSHA string) (reportWire, error) {
	if !model.ValidHead(headSHA) {
		return reportWire{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	baseline, err := parseFixture(source)
	if err != nil {
		return reportWire{}, err
	}
	if baseline.CaseID != preservedCaseID || baseline.Expected != 3 {
		return reportWire{}, fmt.Errorf("source is not the fixed preserved intervention fixture")
	}
	semanticSource, err := mutateSemantic(source)
	if err != nil {
		return reportWire{}, err
	}
	operationSource, err := mutateOperation(source)
	if err != nil {
		return reportWire{}, err
	}
	nonSemanticSource, err := mutateNonSemantic(source)
	if err != nil {
		return reportWire{}, err
	}
	semanticExpected, err := reconstructCase(source, semanticSource, headSHA, semanticExpectedCaseID, "SEMANTIC_EXPECTED", "semantic-expected-value-change", semanticExpectedClaimID, semanticExpectedStep, semanticExpectedReason)
	if err != nil {
		return reportWire{}, err
	}
	semanticOperation, err := reconstructCase(source, operationSource, headSHA, semanticOperationCaseID, "SEMANTIC_OPERATION", "semantic-candidate-operation-change", semanticOperationClaimID, semanticOperationStep, semanticOperationReason)
	if err != nil {
		return reportWire{}, err
	}
	nonSemantic, err := reconstructCase(source, nonSemanticSource, headSHA, nonSemanticCaseID, "NON_SEMANTIC", nonSemanticInterventionLabel, nonSemanticClaimID, nonSemanticStep, nonSemanticReason)
	if err != nil {
		return reportWire{}, err
	}
	cases := []caseWire{semanticExpected, semanticOperation, nonSemantic}
	decision, resolution, reason, failure := deriveReport(cases)
	repositoryWrites, mutationAuthority := effectTotals(cases)
	report := reportWire{Schema: reportSchema, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: model.DigestBytes(source),
		Denominator: fixedDenominatorWire{ID: denominatorID, CasesTotal: 3,
			SemanticExpectedChange:  sliceDenominatorWire{ID: semanticExpectedDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticExpected.Satisfied), CoverageBPS: boolInt(semanticExpected.Satisfied) * 10_000},
			SemanticOperationChange: sliceDenominatorWire{ID: semanticOperationDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticOperation.Satisfied), CoverageBPS: boolInt(semanticOperation.Satisfied) * 10_000},
			NonSemantic:             sliceDenominatorWire{ID: nonSemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(nonSemantic.Satisfied), CoverageBPS: boolInt(nonSemantic.Satisfied) * 10_000}},
		CaseCount: 3, Cases: cases, Decision: decision, Resolution: resolution, Reason: reason, Failure: failure,
		RepositoryWrites: repositoryWrites, MutationAuthority: mutationAuthority}
	return sealReport(report), nil
}

func deriveReport(cases []caseWire) (string, string, string, *failureWire) {
	if len(cases) == 3 && cases[0].Satisfied && cases[1].Satisfied && cases[2].Satisfied {
		return model.DecisionPass, model.ResolutionExact, "ALL_INTERVENTION_OBSERVATIONS_SATISFIED", nil
	}
	for _, item := range cases {
		if item.Satisfied {
			continue
		}
		failure := &failureWire{CaseID: item.ID, Stage: item.Claim.Coordinate.Stage, Step: item.Claim.Coordinate.Step, Reason: item.Claim.Reason}
		resolution := item.Claim.Resolution
		if resolution == "" {
			resolution = model.ResolutionLower
		}
		return failClosedDecision, resolution, "CASE=" + failure.CaseID + ";STAGE=" + failure.Stage + ";STEP=" + failure.Step + ";REASON=" + failure.Reason, failure
	}
	failure := &failureWire{CaseID: "intervention-denominator", Stage: interventionStage, Step: "adjudicate-fixed-cases", Reason: evidenceUnobservableReason}
	return failClosedDecision, model.ResolutionLower, "CASE=" + failure.CaseID + ";STAGE=" + failure.Stage + ";STEP=" + failure.Step + ";REASON=" + failure.Reason, failure
}

func effectTotals(cases []caseWire) (int, bool) {
	writes := 0
	mutationAuthority := false
	for _, item := range cases {
		writes += item.BaselineRepositoryWrites + item.MutatedRepositoryWrites
		mutationAuthority = mutationAuthority || item.BaselineMutationAuthority || item.MutatedMutationAuthority
	}
	return writes, mutationAuthority
}

func reconstructCase(source, mutated []byte, headSHA, id, kind, edit, claimID, step, satisfiedReason string) (caseWire, error) {
	baselineProjection, err := parseFixture(source)
	if err != nil {
		return caseWire{}, err
	}
	mutatedProjection, err := parseFixture(mutated)
	if err != nil {
		return caseWire{}, err
	}
	baselineReceipt, err := buildReceipt(source, headSHA)
	if err != nil {
		return caseWire{}, err
	}
	mutatedReceipt, err := buildReceipt(mutated, headSHA)
	if err != nil {
		return caseWire{}, err
	}
	baselineJudgment := judgmentFromReceipt(baselineReceipt)
	mutatedJudgment := judgmentFromReceipt(mutatedReceipt)
	baselineProjectionWire := projectionFromFixture(baselineProjection)
	mutatedProjectionWire := projectionFromFixture(mutatedProjection)
	baselineTransitions := transitionsFromReceipt(baselineReceipt)
	mutatedTransitions := transitionsFromReceipt(mutatedReceipt)
	semanticEqual := reflect.DeepEqual(baselineProjectionWire, mutatedProjectionWire)
	decisionEqual := baselineJudgment.Decision == mutatedJudgment.Decision && baselineReceipt.Decision == mutatedReceipt.Decision
	resolutionEqual := baselineJudgment.Resolution == mutatedJudgment.Resolution && baselineReceipt.Resolution == mutatedReceipt.Resolution
	reasonEqual := baselineJudgment.Reason == mutatedJudgment.Reason && baselineReceipt.Reason == mutatedReceipt.Reason
	rawDigestChanged := baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest
	receiptChanged := baselineReceipt.Digest != mutatedReceipt.Digest
	transitionsEqual := reflect.DeepEqual(baselineTransitions, mutatedTransitions)
	effectsEqual := reflect.DeepEqual(baselineReceipt.Effects, mutatedReceipt.Effects)
	replayEqual := baselineReceipt.Evidence.ReplayCount == mutatedReceipt.Evidence.ReplayCount && baselineReceipt.Evidence.ReplayOperation == mutatedReceipt.Evidence.ReplayOperation &&
		baselineReceipt.Evidence.ReplayOutput == mutatedReceipt.Evidence.ReplayOutput && baselineReceipt.Evidence.ReplayDigest == mutatedReceipt.Evidence.ReplayDigest &&
		baselineReceipt.Evidence.ReplaySemanticDigest == mutatedReceipt.Evidence.ReplaySemanticDigest && baselineReceipt.Evidence.ReplayEvidenceDigest == mutatedReceipt.Evidence.ReplayEvidenceDigest
	writesZero := baselineReceipt.RepositoryWrites == 0 && mutatedReceipt.RepositoryWrites == 0 && !baselineReceipt.MutationAuthority && !mutatedReceipt.MutationAuthority
	evidenceObservable := baselineJudgment.Independent && mutatedJudgment.Independent
	caseResult := caseWire{ID: id, Kind: kind, SourceEdit: edit, BaselineProjection: baselineProjectionWire, MutatedProjection: mutatedProjectionWire,
		BaselineProjectionDigest: model.Digest(baselineProjectionWire), MutatedProjectionDigest: model.Digest(mutatedProjectionWire),
		BaselineSourceDigest: baselineReceipt.SourceDigest, MutatedSourceDigest: mutatedReceipt.SourceDigest, BaselineReceiptDigest: baselineReceipt.Digest, MutatedReceiptDigest: mutatedReceipt.Digest,
		BaselineReceiptDecision: baselineReceipt.Decision, MutatedReceiptDecision: mutatedReceipt.Decision, BaselineJudgment: judgmentFromModel(baselineJudgment), MutatedJudgment: judgmentFromModel(mutatedJudgment),
		BaselineEvidence: evidenceFromModel(baselineReceipt.Evidence), MutatedEvidence: evidenceFromModel(mutatedReceipt.Evidence), BaselineClaimTransitions: baselineTransitions, MutatedClaimTransitions: mutatedTransitions,
		RawSourceDigestChanged: rawDigestChanged, ReceiptChanged: receiptChanged, SemanticProjectionEqual: semanticEqual, DecisionEqual: decisionEqual, ResolutionEqual: resolutionEqual,
		ReasonEqual: reasonEqual, DecisionChanged: !decisionEqual, ClaimTransitionsEqual: transitionsEqual, EffectsEqual: effectsEqual, ReplayObservationEqual: replayEqual,
		EvidenceObservable: evidenceObservable, RepositoryWritesZero: writesZero, BaselineRepositoryWrites: baselineReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites,
		BaselineMutationAuthority: baselineReceipt.MutationAuthority, MutatedMutationAuthority: mutatedReceipt.MutationAuthority}
	observationSatisfied := (kind == "SEMANTIC_EXPECTED" || kind == "SEMANTIC_OPERATION") && rawDigestChanged && receiptChanged && !semanticEqual && !decisionEqual && !resolutionEqual && !reasonEqual && !transitionsEqual && writesZero && baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionRefuted && mutatedJudgment.Reason == "SEMANTIC_POSTCONDITION_REFUTED"
	if kind == "NON_SEMANTIC" {
		observationSatisfied = rawDigestChanged && receiptChanged && semanticEqual && decisionEqual && resolutionEqual && reasonEqual && transitionsEqual && effectsEqual && replayEqual && writesZero && baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionAllowed
	}
	claimReason := satisfiedReason
	claimResolution := model.ResolutionExact
	claimStatus := model.StatusDischarged
	if !observationSatisfied {
		claimResolution = model.ResolutionInvariant
		claimStatus = model.StatusRefuted
		if kind == "NON_SEMANTIC" {
			claimReason = nonSemanticContradictionReason
		} else {
			claimReason = semanticContradictionReason
		}
	}
	coordinate := coordinateWire{Stage: interventionStage, Step: step, Reason: claimReason}
	caseResult.Claim = claimWire{ID: claimID, Status: claimStatus, Resolution: claimResolution, Reason: claimReason, Coordinate: coordinate,
		Transitions: []transitionWire{{From: model.StatusOpen, To: claimStatus, Coordinate: coordinate}}}
	caseResult.Satisfied = observationSatisfied
	return caseResult, nil
}

func buildReceipt(source []byte, headSHA string) (model.Receipt, error) {
	fixture, err := parseFixture(source)
	if err != nil {
		return model.Receipt{}, err
	}
	contract := model.CanonicalContract()
	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.SemanticDigest(fixture.Input)
	semanticAfter := model.SemanticDigest(fixture.CandidateResult)
	expectedSemantic := model.SemanticDigest(fixture.Expected)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	replayOutput, replayErr := executeReplay(fixture.ReplayRecipe, fixture.Input)
	evidence := model.TransformationEvidence{SourceDigest: sourceDigest, InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation, CandidateResult: fixture.CandidateResult,
		ExpectedValue: fixture.Expected, Invariant: fixture.Invariant, CandidateDigest: candidateDigest, SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
		ExpectedSemanticDigest: expectedSemantic, ReplayRecipe: fixture.ReplayRecipe, BaselineInputValue: fixture.Input, BaselineOperation: fixture.CandidateOperation,
		BaselineOutput: fixture.CandidateResult, BaselineDigest: candidateDigest, ReplayCount: 1}
	if replayErr != nil {
		evidence.ReplayFailureStage = "REGRESSION"
		evidence.ReplayFailureStep = "execute-replay"
		evidence.ReplayFailureReason = replayFailureReason(fixture.ReplayRecipe)
	} else {
		replayDigest := model.CandidateDigest(fixture.ReplayRecipe, fixture.Input, replayOutput)
		evidence.ReplayInputValue = fixture.Input
		evidence.ReplayOperation = fixture.ReplayRecipe
		evidence.ReplayOutput = replayOutput
		evidence.ReplayDigest = replayDigest
		evidence.ReplaySemanticDigest = model.SemanticDigest(replayOutput)
		evidence.ReplayEvidenceDigest = model.ReplayDigest(candidateDigest, replayDigest)
		evidence.ReplayCount = 2
		evidence.RegressionWitnessPresent = candidateDigest == replayDigest && semanticAfter == evidence.ReplaySemanticDigest
	}
	postconditionDigest := model.PostconditionDigest(semanticBefore, semanticAfter, expectedSemantic)
	regressionDigest := ""
	if evidence.ReplayCount == 2 {
		regressionDigest = evidence.ReplayEvidenceDigest
	}
	statuses := map[string]string{"precondition": model.StatusDischarged, "transformation": model.StatusDischarged, "postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged}
	reasons := map[string]string{"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED", "postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED"}
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if evidence.ReplayCount != 2 {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = evidence.ReplayFailureReason
	} else if !evidence.RegressionWitnessPresent {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = replayMismatchReason
	} else if fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}
	claims := make([]model.Claim, 0, len(contract.Values))
	values := make([]model.MetaValue, 0, len(contract.Values))
	for _, valueSpec := range contract.Values {
		evidenceDigest := evidenceFor(valueSpec.ID, sourceDigest, candidateDigest, postconditionDigest, regressionDigest)
		coordinate := model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]}
		claim := model.Claim{ID: valueSpec.ID, Status: statuses[valueSpec.ID], Reason: reasons[valueSpec.ID], Coordinate: coordinate, EvidenceDigests: evidenceDigests(evidenceDigest), Transitions: []model.Transition{{From: model.StatusOpen, To: statuses[valueSpec.ID], Coordinate: coordinate}}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: valueSpec.ID, Kind: valueSpec.Kind, Value: statuses[valueSpec.ID], EvidenceDigest: evidenceDigest, Producer: valueSpec.Producer,
			Consumer: valueSpec.Consumer, MetaOperation: valueSpec.MetaOperation, ProofChoice: valueSpec.ProofChoice, Coordinate: coordinate})
	}
	decision, resolution, reason := deriveDecision(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: fixture.CaseID, CaseKind: "PRESERVED", HeadSHA: headSHA, SourcePath: model.SourcePath,
		SourceDigest: sourceDigest, ContractDigest: model.Digest(contract), Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp,
		ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence, Decision: decision, Resolution: resolution, Reason: reason,
		Effects: []model.Effect{}, RepositoryWrites: 0, MutationAuthority: false, AuthorityScope: model.AuthorityScope}
	if fixture.ApprovedArtifact {
		data := approvedArtifactBytes(fixture)
		path := approvedArtifactPath()
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return model.Receipt{}, err
		}
		receipt.Effects = append(receipt.Effects, model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: model.DigestBytes(data), ArtifactPath: path, ArtifactSize: len(data), Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: "record-approved-artifact-effect", RepositoryWrites: 0, MutationAuthority: false})
	}
	return model.SealReceipt(receipt), nil
}

func parseFixture(source []byte) (sourceFixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceFixture{}, fmt.Errorf("consumer source syntax: %s", diagnostics.Error())
	}
	var found *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != "PreservedTranslation" {
			continue
		}
		if found != nil {
			return sourceFixture{}, fmt.Errorf("duplicate source activity %q", activity.Name)
		}
		found = activity
	}
	if found == nil || len(found.Parameters) != 0 || found.Result.Name != "Transformation" || !found.ValueProgramPresent {
		return sourceFixture{}, fmt.Errorf("preserved source activity is not executable")
	}
	fields, err := decodeFields(found.ValueProgram)
	if err != nil {
		return sourceFixture{}, err
	}
	input, err := strconv.ParseInt(fields["input"], 10, 64)
	if err != nil {
		return sourceFixture{}, fmt.Errorf("source input is not int64: %w", err)
	}
	expected, err := strconv.ParseInt(fields["expected"], 10, 64)
	if err != nil {
		return sourceFixture{}, fmt.Errorf("source expected is not int64: %w", err)
	}
	candidateResult, err := evaluateAdd(fields["candidate"], input)
	if err != nil {
		return sourceFixture{}, err
	}
	if fields["case"] != preservedCaseID || fields["invariant"] != "candidate-output-equals-expected" || fields["replay"] == "" || (fields["effect"] != "none" && fields["effect"] != "approved-artifact") {
		return sourceFixture{}, fmt.Errorf("source fixture is outside the bounded intervention contract")
	}
	semanticDigest, err := canonicalSemanticDigest(source)
	if err != nil {
		return sourceFixture{}, err
	}
	return sourceFixture{Activity: found.Name, CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult, Expected: expected, Invariant: fields["invariant"], ReplayRecipe: fields["replay"], SemanticSourceDigest: semanticDigest, ApprovedArtifact: fields["effect"] == "approved-artifact"}, nil
}

func projectionFromFixture(fixture sourceFixture) projectionWire {
	return projectionWire{Activity: fixture.Activity, CaseID: fixture.CaseID, Input: fixture.Input, CandidateOperation: fixture.CandidateOperation, CandidateResult: fixture.CandidateResult,
		Expected: fixture.Expected, Invariant: fixture.Invariant, ReplayRecipe: fixture.ReplayRecipe, SemanticSourceDigest: fixture.SemanticSourceDigest, ApprovedArtifact: fixture.ApprovedArtifact}
}

func decodeFields(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 7 {
		return nil, fmt.Errorf("fixture computes value has %d fields, want 7", len(parts))
	}
	fields := make(map[string]string, len(parts))
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
	for _, key := range []string{"case", "input", "candidate", "expected", "invariant", "replay", "effect"} {
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
		return 0, fmt.Errorf("candidate operand is not int64: %w", err)
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("candidate operation %q overflows int64", operation)
	}
	return input + operand, nil
}

func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("%s", replayUnavailableReason)
	}
	return evaluateAdd(recipe, input)
}

func replayFailureReason(recipe string) string {
	if recipe == "unavailable" {
		return replayUnavailableReason
	}
	return replayExecutionReason
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

func mutateSemantic(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;input=2;candidate=add:1;" + originalSemanticMutation)
	newValue := []byte("case=preserved-translation;input=2;candidate=add:1;" + expectedSemanticMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("semantic intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}

func mutateOperation(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;input=2;candidate=" + originalOperationMutation + ";expected=3;invariant=candidate-output-equals-expected;replay=" + originalOperationMutation)
	newValue := []byte("case=preserved-translation;input=2;candidate=" + expectedOperationMutation + ";expected=3;invariant=candidate-output-equals-expected;replay=" + expectedOperationMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("operation intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}

func mutateNonSemantic(source []byte) ([]byte, error) {
	return append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...), nil
}

func transitionsFromReceipt(receipt model.Receipt) []transitionWire {
	result := make([]transitionWire, 0, len(receipt.Claims))
	for _, claim := range receipt.Claims {
		for _, transition := range claim.Transitions {
			result = append(result, transitionFromModel(transition))
		}
	}
	return result
}

func judgmentFromReceipt(receipt model.Receipt) model.Judgment {
	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects)}
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
	judgment.Decision, judgment.Resolution, judgment.Reason = deriveDecision(receipt.Claims)
	if judgment.Decision == model.DecisionAllowed {
		judgment.Status = model.StatusDischarged
	} else if judgment.Decision == model.DecisionBlocked {
		judgment.Status = model.StatusOpen
	} else {
		judgment.Status = model.StatusRefuted
	}
	return judgment
}

func deriveDecision(claims []model.Claim) (string, string, string) {
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

func evidenceFor(id, sourceDigest, candidate, postcondition, regression string) string {
	switch id {
	case "precondition":
		return sourceDigest
	case "transformation":
		return candidate
	case "postcondition":
		return postcondition
	case "regression-witness":
		return regression
	default:
		return ""
	}
}

func evidenceDigests(digest string) []string {
	if digest == "" {
		return []string{}
	}
	return []string{digest}
}

func coherentTamper(report reportWire) reportWire {
	tampered := report
	tampered.Cases = append([]caseWire(nil), report.Cases...)
	tampered.Cases[2] = report.Cases[2]
	tampered.Cases[2].RawSourceDigestChanged = false
	tampered.Cases[2].ReceiptChanged = false
	tampered.Cases[2].DecisionEqual = false
	tampered.Cases[2].ResolutionEqual = false
	tampered.Cases[2].ReasonEqual = false
	tampered.Cases[2].DecisionChanged = true
	tampered.Cases[2].ClaimTransitionsEqual = false
	tampered.Cases[2].Claim.Status = model.StatusRefuted
	tampered.Cases[2].Claim.Resolution = model.ResolutionInvariant
	tampered.Cases[2].Claim.Reason = nonSemanticContradictionReason
	tampered.Cases[2].Claim.Coordinate.Reason = nonSemanticContradictionReason
	tampered.Cases[2].Claim.Transitions = []transitionWire{{From: model.StatusOpen, To: model.StatusRefuted, Coordinate: tampered.Cases[2].Claim.Coordinate}}
	tampered.Cases[2].Satisfied = false
	tampered.Denominator.NonSemantic.CasesSatisfied = 0
	tampered.Denominator.NonSemantic.CoverageBPS = 0
	tampered.Decision = failClosedDecision
	tampered.Resolution = model.ResolutionInvariant
	tampered.Reason = "CASE=" + nonSemanticCaseID + ";STAGE=" + interventionStage + ";STEP=" + nonSemanticStep + ";REASON=" + nonSemanticContradictionReason
	tampered.Failure = &failureWire{CaseID: nonSemanticCaseID, Stage: interventionStage, Step: nonSemanticStep, Reason: nonSemanticContradictionReason}
	return sealReport(tampered)
}

func approvedArtifactPath() string {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	return filepath.Join(root, "gooo-invariant-transformation-approved-artifact.bin")
}

func approvedArtifactBytes(fixture sourceFixture) []byte {
	return []byte(fmt.Sprintf("gooo approved artifact\ncase=%s\ninput=%d\noperation=%s\noutput=%d\n", fixture.CaseID, fixture.Input, fixture.CandidateOperation, fixture.CandidateResult))
}

func sealReport(report reportWire) reportWire {
	report.Digest = ""
	report.Digest = model.Digest(report)
	return report
}

func sealAudit(audit Audit) string {
	audit.Digest = ""
	return model.Digest(audit)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func coordinateFromModel(coordinate model.Coordinate) coordinateWire {
	return coordinateWire{Stage: coordinate.Stage, Step: coordinate.Step, Reason: coordinate.Reason}
}

func transitionFromModel(transition model.Transition) transitionWire {
	return transitionWire{From: transition.From, To: transition.To, Coordinate: coordinateFromModel(transition.Coordinate)}
}

func judgmentFromModel(judgment model.Judgment) judgmentWire {
	return judgmentWire{Decision: judgment.Decision, Resolution: judgment.Resolution, Reason: judgment.Reason, Status: judgment.Status, Independent: judgment.Independent,
		CheckedClaims: judgment.CheckedClaims, DischargedClaims: judgment.DischargedClaims, OpenClaims: judgment.OpenClaims, RefutedClaims: judgment.RefutedClaims, Effects: judgment.Effects}
}

func evidenceFromModel(evidence model.TransformationEvidence) evidenceWire {
	return evidenceWire{SourceDigest: evidence.SourceDigest, InputValue: evidence.InputValue, CandidateOperation: evidence.CandidateOperation, CandidateResult: evidence.CandidateResult,
		ExpectedValue: evidence.ExpectedValue, Invariant: evidence.Invariant, CandidateDigest: evidence.CandidateDigest, SemanticBeforeDigest: evidence.SemanticBeforeDigest,
		SemanticAfterDigest: evidence.SemanticAfterDigest, ExpectedSemanticDigest: evidence.ExpectedSemanticDigest, ReplayRecipe: evidence.ReplayRecipe,
		BaselineInputValue: evidence.BaselineInputValue, BaselineOperation: evidence.BaselineOperation, BaselineOutput: evidence.BaselineOutput, BaselineDigest: evidence.BaselineDigest,
		ReplayInputValue: evidence.ReplayInputValue, ReplayOperation: evidence.ReplayOperation, ReplayOutput: evidence.ReplayOutput, ReplayDigest: evidence.ReplayDigest,
		ReplaySemanticDigest: evidence.ReplaySemanticDigest, ReplayEvidenceDigest: evidence.ReplayEvidenceDigest, RegressionWitnessPresent: evidence.RegressionWitnessPresent,
		ReplayCount: evidence.ReplayCount, ReplayFailureStage: evidence.ReplayFailureStage, ReplayFailureStep: evidence.ReplayFailureStep, ReplayFailureReason: evidence.ReplayFailureReason,
		SemanticSourceDigest: evidence.SemanticSourceDigest}
}
