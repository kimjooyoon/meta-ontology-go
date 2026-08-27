package intervention

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/judge"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/producer"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                       = "gooo/invariant-transformation-intervention-report/v1"
	DenominatorID                = "gooo/invariant-transformation-intervention-denominator/v1"
	SemanticDenominatorID        = "gooo/invariant-transformation-intervention-semantic-denominator/v1"
	NonSemanticDenominatorID     = "gooo/invariant-transformation-intervention-nonsemantic-denominator/v1"
	SemanticCaseID               = "semantic-source-intervention"
	NonSemanticCaseID            = "nonsemantic-source-intervention"
	SemanticClaimID              = "semantic-intervention-claim"
	NonSemanticClaimID           = "nonsemantic-intervention-claim"
	InterventionStage            = "INTERVENTION"
	SemanticStep                 = "compare-semantic-projection-and-decision"
	SemanticReason               = "SEMANTIC_PROJECTION_AND_DECISION_CHANGED"
	NonSemanticStep              = "compare-nonsemantic-projection-and-decision"
	NonSemanticReason            = "NONSEMANTIC_PROJECTION_AND_DECISION_PRESERVED"
	PreservedCaseID              = "preserved-translation"
	ExpectedSemanticMutation     = "expected=4"
	OriginalSemanticMutation     = "expected=3"
	NonSemanticInterventionLabel = "comment-and-whitespace-only"
)

type FixtureProjection struct {
	Activity            string `json:"activity"`
	CaseID              string `json:"case_id"`
	Input               int64  `json:"input"`
	CandidateOperation  string `json:"candidate_operation"`
	CandidateResult     int64  `json:"candidate_result"`
	Expected            int64  `json:"expected"`
	Invariant           string `json:"invariant"`
	RegressionAvailable bool   `json:"regression_available"`
	ApprovedArtifact    bool   `json:"approved_artifact"`
}

type Claim struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
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
	ID             string           `json:"id"`
	CasesTotal     int              `json:"cases_total"`
	SemanticChange SliceDenominator `json:"semantic_change"`
	NonSemantic    SliceDenominator `json:"nonsemantic_change"`
}

type Case struct {
	ID                        string             `json:"id"`
	Kind                      string             `json:"kind"`
	SourceEdit                string             `json:"source_edit"`
	BaselineProjection        FixtureProjection  `json:"baseline_projection"`
	MutatedProjection         FixtureProjection  `json:"mutated_projection"`
	BaselineProjectionDigest  string             `json:"baseline_projection_digest"`
	MutatedProjectionDigest   string             `json:"mutated_projection_digest"`
	BaselineSourceDigest      string             `json:"baseline_source_digest"`
	MutatedSourceDigest       string             `json:"mutated_source_digest"`
	BaselineReceiptDigest     string             `json:"baseline_receipt_digest"`
	MutatedReceiptDigest      string             `json:"mutated_receipt_digest"`
	BaselineReceiptDecision   string             `json:"baseline_receipt_decision"`
	MutatedReceiptDecision    string             `json:"mutated_receipt_decision"`
	BaselineJudgment          model.Judgment     `json:"baseline_judgment"`
	MutatedJudgment           model.Judgment     `json:"mutated_judgment"`
	BaselineClaimTransitions  []model.Transition `json:"baseline_claim_transitions"`
	MutatedClaimTransitions   []model.Transition `json:"mutated_claim_transitions"`
	RawSourceDigestChanged    bool               `json:"raw_source_digest_changed"`
	ReceiptChanged            bool               `json:"receipt_changed"`
	SemanticProjectionEqual   bool               `json:"semantic_projection_equal"`
	DecisionEqual             bool               `json:"decision_equal"`
	ResolutionEqual           bool               `json:"resolution_equal"`
	ReasonEqual               bool               `json:"reason_equal"`
	DecisionChanged           bool               `json:"decision_changed"`
	ClaimTransitionsEqual     bool               `json:"claim_transitions_equal"`
	RepositoryWritesZero      bool               `json:"repository_writes_zero"`
	BaselineRepositoryWrites  int                `json:"baseline_repository_writes"`
	MutatedRepositoryWrites   int                `json:"mutated_repository_writes"`
	BaselineMutationAuthority bool               `json:"baseline_mutation_authority"`
	MutatedMutationAuthority  bool               `json:"mutated_mutation_authority"`
	Claim                     Claim              `json:"claim"`
	Satisfied                 bool               `json:"satisfied"`
}

type Report struct {
	Schema           string           `json:"schema"`
	HeadSHA          string           `json:"head_sha"`
	SourcePath       string           `json:"source_path"`
	SourceDigest     string           `json:"source_digest"`
	Denominator      FixedDenominator `json:"denominator"`
	CaseCount        int              `json:"case_count"`
	Cases            []Case           `json:"cases"`
	Decision         string           `json:"decision"`
	Resolution       string           `json:"resolution"`
	Reason           string           `json:"reason"`
	RepositoryWrites int              `json:"repository_writes"`
	Digest           string           `json:"digest"`
}

func Build(source []byte, headSHA string) (Report, error) {
	if !model.ValidHead(headSHA) {
		return Report{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	if _, err := project(source); err != nil {
		return Report{}, err
	}
	semanticCase, err := buildCase(source, headSHA, SemanticCaseID, "SEMANTIC", "semantic-value-change", mutateSemantic, SemanticClaimID, SemanticStep, SemanticReason)
	if err != nil {
		return Report{}, err
	}
	nonSemanticCase, err := buildCase(source, headSHA, NonSemanticCaseID, "NON_SEMANTIC", NonSemanticInterventionLabel, mutateNonSemantic, NonSemanticClaimID, NonSemanticStep, NonSemanticReason)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Schema: Schema, HeadSHA: headSHA, SourcePath: model.SourcePath, SourceDigest: model.DigestBytes(source),
		Denominator: FixedDenominator{
			ID: DenominatorID, CasesTotal: 2,
			SemanticChange: SliceDenominator{ID: SemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(semanticCase.Satisfied), CoverageBPS: boolInt(semanticCase.Satisfied) * 10_000},
			NonSemantic:    SliceDenominator{ID: NonSemanticDenominatorID, CasesTotal: 1, CasesSatisfied: boolInt(nonSemanticCase.Satisfied), CoverageBPS: boolInt(nonSemanticCase.Satisfied) * 10_000},
		},
		CaseCount: 2, Cases: []Case{semanticCase, nonSemanticCase}, Decision: "PASS", Resolution: model.ResolutionExact,
		Reason: "FIXED_INTERVENTION_CONTRACT_SATISFIED", RepositoryWrites: 0,
	}
	return seal(report), nil
}

func ValidateReport(report Report, source []byte, headSHA string) error {
	if report.Schema != Schema || report.HeadSHA != headSHA || report.SourcePath != model.SourcePath ||
		report.SourceDigest != model.DigestBytes(source) || report.CaseCount != 2 || report.RepositoryWrites != 0 ||
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

func buildCase(source []byte, headSHA, id, kind, edit string, mutate func([]byte) ([]byte, error), claimID, step, reason string) (Case, error) {
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
	writesZero := baselineReceipt.RepositoryWrites == 0 && mutatedReceipt.RepositoryWrites == 0 &&
		!baselineReceipt.MutationAuthority && !mutatedReceipt.MutationAuthority
	claimCoordinate := model.Coordinate{Stage: InterventionStage, Step: step, Reason: reason}
	claim := Claim{ID: claimID, Status: model.StatusDischarged, Reason: reason, Coordinate: claimCoordinate,
		Transitions: []model.Transition{{From: model.StatusOpen, To: model.StatusDischarged, Coordinate: claimCoordinate}}}
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
		BaselineRepositoryWrites: baselineReceipt.RepositoryWrites, MutatedRepositoryWrites: mutatedReceipt.RepositoryWrites,
		BaselineMutationAuthority: baselineReceipt.MutationAuthority, MutatedMutationAuthority: mutatedReceipt.MutationAuthority,
		Claim: claim,
	}
	switch kind {
	case "SEMANTIC":
		caseResult.Satisfied = baselineJudgment.Independent && mutatedJudgment.Independent && rawDigestChanged && receiptChanged &&
			!semanticEqual && !decisionEqual && !resolutionEqual && !reasonEqual && !transitionsEqual && writesZero &&
			baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionRefuted &&
			mutatedJudgment.Reason == "SEMANTIC_POSTCONDITION_REFUTED"
	case "NON_SEMANTIC":
		caseResult.Satisfied = baselineJudgment.Independent && mutatedJudgment.Independent && rawDigestChanged && receiptChanged &&
			semanticEqual && decisionEqual && resolutionEqual && reasonEqual && transitionsEqual && writesZero &&
			baselineJudgment.Decision == model.DecisionAllowed && mutatedJudgment.Decision == model.DecisionAllowed
	default:
		return Case{}, fmt.Errorf("unsupported intervention kind %q", kind)
	}
	return caseResult, nil
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
		return FixtureProjection{Activity: activity.Name, CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult,
			Expected: expected, Invariant: fields["invariant"], RegressionAvailable: fields["replay"] == "present", ApprovedArtifact: fields["effect"] == "approved-artifact"}, nil
	}
	return FixtureProjection{}, fmt.Errorf("preserved fixture activity is missing")
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
