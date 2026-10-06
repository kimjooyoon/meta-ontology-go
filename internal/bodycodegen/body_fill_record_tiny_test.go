package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSourceRecordIRBodyFillUsesTinyGoForUniqueGeneratedOperations(t *testing.T) {
	result, plan, generation, provider, modelLoadMS := runTinyGoRecordBodyFill(t)
	assertTinyGoRecordBodyFill(t, result, plan, generation, provider, modelLoadMS)
}

func runTinyGoRecordBodyFill(t *testing.T) (Result, IRBodyFillPlan, *IRBodyFillCandidateGenerationReceipt,
	*recordingTinyGoBodyFillProvider, float64) {
	t.Helper()
	source, activity := readTinyGoRecordBodyFillFixture(t)
	plan, generation, err := sourceIRBodyFillPlan("record-fill-tiny.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec)
	if err != nil {
		t.Fatal(err)
	}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAnd}
	modelLoadMS := 1.25
	result, err := generateWithIRBodyFillOptions(context.Background(), "record-fill-tiny.gooo", source,
		"ReviewCandidate", plan, "", "", IRBodyFillOptions{TinyModelLoadMS: &modelLoadMS}, provider)
	if err != nil {
		t.Fatal(err)
	}
	return result, plan, generation, provider, modelLoadMS
}

func readTinyGoRecordBodyFillFixture(t *testing.T) ([]byte, *syntax.ActivityDecl) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill-tiny.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity.Name == "ReviewCandidate" {
			return source, activity
		}
	}
	t.Fatal("TinyGo record body-fill activity not found")
	return nil, nil
}

func assertTinyGoRecordBodyFill(t *testing.T, result Result, plan IRBodyFillPlan,
	generation *IRBodyFillCandidateGenerationReceipt, provider *recordingTinyGoBodyFillProvider, modelLoadMS float64) {
	t.Helper()
	receipt := result.Report.BodyFill
	if receipt == nil || provider.calls != 1 || provider.request.Intent != plan.Intent ||
		receipt.Decision.Provider != decisionroute.ProviderTinyGo || receipt.TinyModelFocusHole != "condition" ||
		receipt.ProposedCandidateID == "" || receipt.SelectedCandidateID != receipt.ProposedCandidateID ||
		receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 2 || receipt.TestCasesTotal != 2 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		receipt.LocalModelPredictions != 1 || receipt.ExternalProviderCalls != 0 || !receipt.ExternalProviderCallsKnown ||
		receipt.Timing.TinyDecisionMS != receipt.Timing.ProviderDecisionMS || receipt.Timing.LayaDecisionMS != 0 ||
		receipt.Timing.TinyModelLoadMS == nil || *receipt.Timing.TinyModelLoadMS != modelLoadMS {
		t.Fatalf("TinyGo record selection was not independently scored and accounted: calls=%d receipt=%+v", provider.calls, receipt)
	}
	if got := []string{provider.request.Question.Options[0].Operation, provider.request.Question.Options[1].Operation}; got[0] != decisionroute.TinyGoOperationAnd || got[1] != decisionroute.TinyGoOperationOr {
		t.Fatalf("record candidates were not mapped to their unique condition operations: %v", got)
	}
	condition := ""
	for _, fill := range receipt.HoleFills {
		if fill.HoleID == "condition" {
			condition = fill.Expression
		}
	}
	// This low-level entry point keeps source_fill; the CLI test verifies that the source wrapper removes it.
	if !strings.Contains(condition, "&&") || !strings.Contains(result.GoooSource, `input.state == \"ready\"`) ||
		strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") ||
		strings.Contains(provider.request.State, `"reviewed":false`) {
		t.Fatalf("TinyGo assignment, emitted Gooo, or case boundary was wrong: condition=%q source=%s", condition, result.GoooSource)
	}
	if generation == nil || generation.AssignmentSpaceSize != 4 || generation.AssignmentsRetained != 2 ||
		generation.AssignmentsOmitted != 2 || generation.AssignmentCoveragePercent != 50 {
		t.Fatalf("record source grammar completeness was not preserved: %+v", generation)
	}
	receipt.CandidateGeneration = generation
	populateCompletenessReceipt(&result.Report, "")
	if dim := bodyFillDimension(result, "external_network_boundary"); dim.Status != "PASS" || dim.Numerator != 1 {
		t.Fatalf("local model was misclassified as an external provider: %+v", dim)
	}
}

func TestTinyGoRecordFillGroupsCandidatesByOperationAndScoresWithinGroup(t *testing.T) {
	result, provider := runGroupedTinyGoRecordBodyFill(t)
	receipt := result.Report.BodyFill
	if provider.calls != 1 || receipt == nil || receipt.TinyModelFocusHole != "condition" ||
		receipt.ProposedCandidateID != "accept_correct" || receipt.SelectedCandidateID != "accept_correct" ||
		receipt.TestCasesPassed != 2 || receipt.TestCasesTotal != 2 || receipt.HoldoutCasesPassed != 2 {
		t.Fatalf("operation class did not resolve to the best compatible complete assignment: calls=%d receipt=%+v", provider.calls, receipt)
	}
	if len(provider.request.Question.Options) != 2 ||
		provider.request.Question.Options[0].ID != decisionroute.TinyGoOperationAnd ||
		provider.request.Question.Options[1].ID != decisionroute.TinyGoOperationOr {
		t.Fatalf("shared operation classes were not collapsed into distinct choices: %+v", provider.request.Question.Options)
	}
	if len(receipt.CandidateScores) != 3 || receipt.CandidateScores[0].TestCasesPassed != 2 ||
		receipt.CandidateScores[1].TestCasesPassed != 1 || receipt.CandidateScores[2].TestCasesPassed != 1 {
		t.Fatalf("complete assignments were not scored before grouped selection: %+v", receipt.CandidateScores)
	}
}

func runGroupedTinyGoRecordBodyFill(t *testing.T) (Result, *recordingTinyGoBodyFillProvider) {
	t.Helper()
	source, activity := readTinyGoRecordBodyFillFixture(t)
	plan, _, err := sourceIRBodyFillPlan("record-fill-tiny.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec)
	if err != nil {
		t.Fatal(err)
	}
	plan.Candidates = []IRBodyFillCandidate{
		{ID: "accept_correct", Fills: map[string]string{"condition": `input.reviewed && input.state == "ready"`, "accepted": `"accepted"`}},
		{ID: "accept_wrong_label", Fills: map[string]string{"condition": `input.reviewed && input.state == "ready"`, "accepted": `"approved"`}},
		{ID: "or_condition", Fills: map[string]string{"condition": `input.reviewed || input.state == "ready"`, "accepted": `"accepted"`}},
	}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationAnd}
	result, err := generateWithIRBodyFillOptions(context.Background(), "record-fill-tiny.gooo", source,
		"ReviewCandidate", plan, "", "", IRBodyFillOptions{}, provider)
	if err != nil {
		t.Fatal(err)
	}
	return result, provider
}

func TestTinyGoRecordFillRejectsAssignmentsWithoutOperationClassBeforeInference(t *testing.T) {
	source, activity := readRecordBodyFillFixture(t)
	spec := activity.Assembly.Spec.Clone()
	for candidateIndex := range spec.FillPlan.Candidates {
		for fillIndex := range spec.FillPlan.Candidates[candidateIndex].Fills {
			if spec.FillPlan.Candidates[candidateIndex].Fills[fillIndex].HoleID == "condition" {
				value := `input.candidate_id == 1`
				if candidateIndex == 1 {
					value = `input.candidate_id == 2`
				}
				spec.FillPlan.Candidates[candidateIndex].Fills[fillIndex].Expression = value
			}
		}
	}
	plan, _, err := sourceIRBodyFillPlan("record-fill.gooo", source, "ReviewCandidate", spec)
	if err != nil {
		t.Fatal(err)
	}
	provider := &recordingTinyGoBodyFillProvider{operation: decisionroute.TinyGoOperationEqual}
	_, err = generateWithIRBodyFillOptions(context.Background(), "record-fill.gooo", source, "ReviewCandidate",
		plan, "", "", IRBodyFillOptions{}, provider)
	if err == nil || !strings.Contains(err.Error(), "at least two distinct supported root operations") || provider.calls != 0 {
		t.Fatalf("TinyGo ran without distinct typed operation classes: calls=%d err=%v", provider.calls, err)
	}
}
