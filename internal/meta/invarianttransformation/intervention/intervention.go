package intervention

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                         = "gooo/invariant-transformation-intervention-report/v1"
	DenominatorID                  = "gooo/invariant-transformation-intervention-denominator/v1"
	SemanticExpectedDenominatorID  = "gooo/invariant-transformation-intervention-semantic-expected-denominator/v1"
	SemanticOperationDenominatorID = "gooo/invariant-transformation-intervention-semantic-operation-denominator/v1"
	NonSemanticDenominatorID       = "gooo/invariant-transformation-intervention-nonsemantic-denominator/v1"
	SemanticExpectedCaseID         = "semantic-expected-intervention"
	SemanticOperationCaseID        = "semantic-operation-intervention"
	NonSemanticCaseID              = "nonsemantic-source-intervention"
	SemanticExpectedClaimID        = "semantic-expected-intervention-claim"
	SemanticOperationClaimID       = "semantic-operation-intervention-claim"
	NonSemanticClaimID             = "nonsemantic-intervention-claim"
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
	FailClosedDecision             = "FAIL_CLOSED"
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
	Input                int64  `json:"input"`
	CandidateOperation   string `json:"candidate_operation"`
	CandidateResult      int64  `json:"candidate_result"`
	Expected             int64  `json:"expected"`
	Invariant            string `json:"invariant"`
	ReplayRecipe         string `json:"replay_recipe"`
	SemanticSourceDigest string `json:"semantic_source_digest"`
	ApprovedArtifact     bool   `json:"approved_artifact"`
}

type Claim struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	Resolution  string             `json:"resolution"`
	Reason      string             `json:"reason"`
	Coordinate  model.Coordinate   `json:"coordinate"`
	Transitions []model.Transition `json:"transitions"`
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

type Case struct {
	ID                        string                       `json:"id"`
	Kind                      string                       `json:"kind"`
	SourceEdit                string                       `json:"source_edit"`
	BaselineProjection        FixtureProjection            `json:"baseline_projection"`
	MutatedProjection         FixtureProjection            `json:"mutated_projection"`
	BaselineProjectionDigest  string                       `json:"baseline_projection_digest"`
	MutatedProjectionDigest   string                       `json:"mutated_projection_digest"`
	BaselineSourceDigest      string                       `json:"baseline_source_digest"`
	MutatedSourceDigest       string                       `json:"mutated_source_digest"`
	BaselineReceiptDigest     string                       `json:"baseline_receipt_digest"`
	MutatedReceiptDigest      string                       `json:"mutated_receipt_digest"`
	BaselineReceiptDecision   string                       `json:"baseline_receipt_decision"`
	MutatedReceiptDecision    string                       `json:"mutated_receipt_decision"`
	BaselineJudgment          model.Judgment               `json:"baseline_judgment"`
	MutatedJudgment           model.Judgment               `json:"mutated_judgment"`
	BaselineEvidence          model.TransformationEvidence `json:"baseline_evidence"`
	MutatedEvidence           model.TransformationEvidence `json:"mutated_evidence"`
	BaselineClaimTransitions  []model.Transition           `json:"baseline_claim_transitions"`
	MutatedClaimTransitions   []model.Transition           `json:"mutated_claim_transitions"`
	RawSourceDigestChanged    bool                         `json:"raw_source_digest_changed"`
	ReceiptChanged            bool                         `json:"receipt_changed"`
	SemanticProjectionEqual   bool                         `json:"semantic_projection_equal"`
	DecisionEqual             bool                         `json:"decision_equal"`
	ResolutionEqual           bool                         `json:"resolution_equal"`
	ReasonEqual               bool                         `json:"reason_equal"`
	DecisionChanged           bool                         `json:"decision_changed"`
	ClaimTransitionsEqual     bool                         `json:"claim_transitions_equal"`
	EffectsEqual              bool                         `json:"effects_equal"`
	ReplayObservationEqual    bool                         `json:"replay_observation_equal"`
	EvidenceObservable        bool                         `json:"evidence_observable"`
	RepositoryWritesZero      bool                         `json:"repository_writes_zero"`
	BaselineRepositoryWrites  int                          `json:"baseline_repository_writes"`
	MutatedRepositoryWrites   int                          `json:"mutated_repository_writes"`
	BaselineMutationAuthority bool                         `json:"baseline_mutation_authority"`
	MutatedMutationAuthority  bool                         `json:"mutated_mutation_authority"`
	Claim                     Claim                        `json:"claim"`
	Satisfied                 bool                         `json:"satisfied"`
}

type Report struct {
	Schema            string           `json:"schema"`
	HeadSHA           string           `json:"head_sha"`
	SourcePath        string           `json:"source_path"`
	SourceDigest      string           `json:"source_digest"`
	Denominator       FixedDenominator `json:"denominator"`
	CaseCount         int              `json:"case_count"`
	Cases             []Case           `json:"cases"`
	Decision          string           `json:"decision"`
	Resolution        string           `json:"resolution"`
	Reason            string           `json:"reason"`
	RepositoryWrites  int              `json:"repository_writes"`
	MutationAuthority bool             `json:"mutation_authority"`
	Failure           *Failure         `json:"failure,omitempty"`
	Digest            string           `json:"digest"`
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
	decision, resolution, reason, failure := deriveReport(cases)
	repositoryWrites, mutationAuthority := effectTotals(cases)
	report := Report{
		Schema: Schema, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: model.DigestBytes(source),
		Denominator: FixedDenominator{
			ID: DenominatorID, CasesTotal: 3,
			SemanticExpectedChange:  SliceDenominator{ID: SemanticExpectedDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticExpectedCase.Satisfied), CoverageBPS: boolInt(semanticExpectedCase.Satisfied) * 10_000},
			SemanticOperationChange: SliceDenominator{ID: SemanticOperationDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticOperationCase.Satisfied), CoverageBPS: boolInt(semanticOperationCase.Satisfied) * 10_000},
			NonSemantic:             SliceDenominator{ID: NonSemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(nonSemanticCase.Satisfied), CoverageBPS: boolInt(nonSemanticCase.Satisfied) * 10_000},
		},
		CaseCount: 3, Cases: cases, Decision: decision, Resolution: resolution,
		Reason: reason, RepositoryWrites: repositoryWrites, MutationAuthority: mutationAuthority, Failure: failure,
	}
	return seal(report), nil
}

// DeterministicReplay compares a report with a second production replay. It
// is a determinism check, not the independent consumer judgment.
func DeterministicReplay(report Report, source []byte, headSHA string) error {
	if report.Schema != Schema || report.HeadSHA != headSHA || report.SourcePath != model.SourcePath ||
		report.SourceDigest != model.DigestBytes(source) || report.CaseCount != 3 ||
		report.Digest == "" || report.Digest != seal(report).Digest {
		return fmt.Errorf("intervention report identity or write boundary mismatch")
	}
	expected, err := Build(source, headSHA)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(report, expected) {
		return fmt.Errorf("intervention report is not independently reproducible")
	}
	return nil
}

func deriveReport(cases []Case) (string, string, string, *Failure) {
	if len(cases) == 3 && cases[0].Satisfied && cases[1].Satisfied && cases[2].Satisfied {
		return model.DecisionPass, model.ResolutionExact, "ALL_INTERVENTION_OBSERVATIONS_SATISFIED", nil
	}
	for _, item := range cases {
		if item.Satisfied {
			continue
		}
		failure := &Failure{CaseID: item.ID, Stage: item.Claim.Coordinate.Stage, Step: item.Claim.Coordinate.Step, Reason: item.Claim.Reason}
		resolution := item.Claim.Resolution
		if resolution == "" {
			resolution = model.ResolutionLower
		}
		return FailClosedDecision, resolution, formatFailureReason(*failure), failure
	}
	failure := &Failure{CaseID: "intervention-denominator", Stage: InterventionStage, Step: "adjudicate-fixed-cases", Reason: EvidenceUnobservableReason}
	return FailClosedDecision, model.ResolutionLower, formatFailureReason(*failure), failure
}

func formatFailureReason(failure Failure) string {
	return fmt.Sprintf("CASE=%s;STAGE=%s;STEP=%s;REASON=%s", failure.CaseID, failure.Stage, failure.Step, failure.Reason)
}

func effectTotals(cases []Case) (int, bool) {
	writes := 0
	mutationAuthority := false
	for _, item := range cases {
		writes += item.BaselineRepositoryWrites + item.MutatedRepositoryWrites
		mutationAuthority = mutationAuthority || item.BaselineMutationAuthority || item.MutatedMutationAuthority
	}
	return writes, mutationAuthority
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
	baselineTransitions := transitions(baselineReceipt)
	mutatedTransitions := transitions(mutatedReceipt)
	semanticEqual := reflect.DeepEqual(baselineProjection, mutatedProjection)
	decisionEqual := baselineJudgment.Decision == mutatedJudgment.Decision && baselineReceipt.Decision == mutatedReceipt.Decision
	resolutionEqual := baselineJudgment.Resolution == mutatedJudgment.Resolution && baselineReceipt.Resolution == mutatedReceipt.Resolution
	reasonEqual := baselineJudgment.Reason == mutatedJudgment.Reason && baselineReceipt.Reason == mutatedReceipt.Reason
	rawDigestChanged := baselineReceipt.SourceDigest != mutatedReceipt.SourceDigest
	receiptChanged := baselineReceipt.Digest != mutatedReceipt.Digest
	transitionsEqual := reflect.DeepEqual(baselineTransitions, mutatedTransitions)
	effectsEqual := reflect.DeepEqual(baselineReceipt.Effects, mutatedReceipt.Effects)
	replayEqual := baselineReceipt.Evidence.ReplayCount == mutatedReceipt.Evidence.ReplayCount &&
		baselineReceipt.Evidence.ReplayOperation == mutatedReceipt.Evidence.ReplayOperation &&
		baselineReceipt.Evidence.ReplayOutput == mutatedReceipt.Evidence.ReplayOutput &&
		baselineReceipt.Evidence.ReplayDigest == mutatedReceipt.Evidence.ReplayDigest &&
		baselineReceipt.Evidence.ReplaySemanticDigest == mutatedReceipt.Evidence.ReplaySemanticDigest &&
		baselineReceipt.Evidence.ReplayEvidenceDigest == mutatedReceipt.Evidence.ReplayEvidenceDigest
	writesZero := baselineReceipt.RepositoryWrites == 0 && mutatedReceipt.RepositoryWrites == 0 &&
		!baselineReceipt.MutationAuthority && !mutatedReceipt.MutationAuthority
	evidenceObservable := baselineJudgment.Independent && mutatedJudgment.Independent
	caseResult := Case{
		ID: id, Kind: kind, SourceEdit: edit, BaselineProjection: baselineProjection, MutatedProjection: mutatedProjection,
		BaselineProjectionDigest: model.Digest(baselineProjection), MutatedProjectionDigest: model.Digest(mutatedProjection),
		BaselineSourceDigest: baselineReceipt.SourceDigest, MutatedSourceDigest: mutatedReceipt.SourceDigest,
		BaselineReceiptDigest: baselineReceipt.Digest, MutatedReceiptDigest: mutatedReceipt.Digest,
		BaselineReceiptDecision: baselineReceipt.Decision, MutatedReceiptDecision: mutatedReceipt.Decision,
		BaselineJudgment: baselineJudgment, MutatedJudgment: mutatedJudgment,
		BaselineClaimTransitions: baselineTransitions, MutatedClaimTransitions: mutatedTransitions,
		RawSourceDigestChanged: rawDigestChanged, ReceiptChanged: receiptChanged, SemanticProjectionEqual: semanticEqual,
		DecisionEqual: decisionEqual, ResolutionEqual: resolutionEqual, ReasonEqual: reasonEqual,
		DecisionChanged: !decisionEqual, ClaimTransitionsEqual: transitionsEqual, RepositoryWritesZero: writesZero,
		EffectsEqual: effectsEqual, ReplayObservationEqual: replayEqual, EvidenceObservable: evidenceObservable,
		BaselineEvidence: baselineReceipt.Evidence, MutatedEvidence: mutatedReceipt.Evidence,
		BaselineRepositoryWrites: baselineReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites,
		BaselineMutationAuthority: baselineReceipt.MutationAuthority, MutatedMutationAuthority: mutatedReceipt.MutationAuthority,
	}
	satisfied, claimResolution, claimReason := adjudicate(kind, caseResult, evidenceObservable, satisfiedReason)
	caseResult.Satisfied = satisfied
	claimCoordinate := model.Coordinate{Stage: InterventionStage, Step: step, Reason: claimReason}
	caseResult.Claim = Claim{ID: claimID, Status: statusForAdjudication(caseResult.Satisfied, evidenceObservable), Resolution: claimResolution,
		Reason: claimReason, Coordinate: claimCoordinate,
		Transitions: []model.Transition{{From: model.StatusOpen, To: statusForAdjudication(caseResult.Satisfied, evidenceObservable), Coordinate: claimCoordinate}}}
	return caseResult, nil
}

func adjudicate(kind string, item Case, evidenceObservable bool, satisfiedReason string) (bool, string, string) {
	if !evidenceObservable {
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
		return item.RawSourceDigestChanged && item.ReceiptChanged && !item.SemanticProjectionEqual &&
			!item.DecisionEqual && !item.ResolutionEqual && !item.ReasonEqual && !item.ClaimTransitionsEqual &&
			item.RepositoryWritesZero && item.BaselineJudgment.Decision == model.DecisionAllowed &&
			item.MutatedJudgment.Decision == model.DecisionRefuted && item.MutatedJudgment.Reason == "SEMANTIC_POSTCONDITION_REFUTED"
	case "NON_SEMANTIC":
		return item.RawSourceDigestChanged && item.ReceiptChanged && item.SemanticProjectionEqual &&
			item.DecisionEqual && item.ResolutionEqual && item.ReasonEqual && item.ClaimTransitionsEqual &&
			item.EffectsEqual && item.ReplayObservationEqual &&
			item.RepositoryWritesZero && item.BaselineJudgment.Decision == model.DecisionAllowed &&
			item.MutatedJudgment.Decision == model.DecisionAllowed
	default:
		return false
	}
}

func statusForAdjudication(satisfied, evidenceObservable bool) string {
	if satisfied {
		return model.StatusDischarged
	}
	if evidenceObservable {
		return model.StatusRefuted
	}
	return model.StatusOpen
}

func project(source []byte) (FixtureProjection, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return FixtureProjection{}, fmt.Errorf("parse intervention source: %s", diagnostics.Error())
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
			return FixtureProjection{}, fmt.Errorf("source input is not int64: %w", err)
		}
		expected, err := strconv.ParseInt(fields["expected"], 10, 64)
		if err != nil {
			return FixtureProjection{}, fmt.Errorf("source expected is not int64: %w", err)
		}
		candidateResult, err := evaluateAdd(fields["candidate"], input)
		if err != nil {
			return FixtureProjection{}, err
		}
		if fields["case"] != PreservedCaseID || fields["invariant"] != "candidate-output-equals-expected" {
			return FixtureProjection{}, fmt.Errorf("preserved fixture declaration is not the contracted projection")
		}
		semanticDigest, err := canonicalSemanticDigest(source)
		if err != nil {
			return FixtureProjection{}, err
		}
		return FixtureProjection{Activity: activity.Name, CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult,
			Expected: expected, Invariant: fields["invariant"], ReplayRecipe: fields["replay"], SemanticSourceDigest: semanticDigest, ApprovedArtifact: fields["effect"] == "approved-artifact"}, nil
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

func mutateSemantic(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;input=2;candidate=add:1;" + OriginalSemanticMutation)
	newValue := []byte("case=preserved-translation;input=2;candidate=add:1;" + ExpectedSemanticMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("semantic intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}

func mutateOperation(source []byte) ([]byte, error) {
	old := []byte("case=preserved-translation;input=2;candidate=" + OriginalOperationMutation + ";expected=3;invariant=candidate-output-equals-expected;replay=" + OriginalOperationMutation)
	newValue := []byte("case=preserved-translation;input=2;candidate=" + ExpectedOperationMutation + ";expected=3;invariant=candidate-output-equals-expected;replay=" + ExpectedOperationMutation)
	if bytes.Count(source, old) != 1 {
		return nil, fmt.Errorf("operation intervention target count is not 1")
	}
	return bytes.Replace(source, old, newValue, 1), nil
}

func mutateNonSemantic(source []byte) ([]byte, error) {
	return append(append([]byte{}, source...), []byte("\n\n// non-semantic intervention: comment and whitespace only\n")...), nil
}

func transitions(receipt model.Receipt) []model.Transition {
	result := make([]model.Transition, 0, len(receipt.Claims))
	for _, claim := range receipt.Claims {
		result = append(result, claim.Transitions...)
	}
	return result
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
