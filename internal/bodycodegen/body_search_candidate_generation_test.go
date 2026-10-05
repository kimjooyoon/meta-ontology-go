package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestIRBodySearchCandidateGenerationEnumeratesStableTypedGrammar(t *testing.T) {
	plan := generatedCandidateSearchPlan(16)
	first, err := generateIRBodySearchCandidates(&plan)
	if err != nil {
		t.Fatal(err)
	}
	secondPlan := generatedCandidateSearchPlan(16)
	second, err := generateIRBodySearchCandidates(&secondPlan)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Candidates, secondPlan.Candidates) || first.CandidateSetSHA256 != second.CandidateSetSHA256 {
		t.Fatalf("generated search space was not deterministic: first=%#v second=%#v", plan.Candidates, secondPlan.Candidates)
	}
	expressions := make([]string, len(plan.Candidates))
	seenIDs := map[string]bool{}
	for i, candidate := range plan.Candidates {
		expressions[i] = candidate.Expression
		if seenIDs[candidate.ID] || !strings.HasPrefix(candidate.ID, "generated_") {
			t.Fatalf("generated candidate ID is unstable or duplicated: %#v", candidate)
		}
		seenIDs[candidate.ID] = true
	}
	for _, want := range []string{"input", "0", "input + 2", "input + 1"} {
		if !containsString(expressions, want) {
			t.Fatalf("generated grammar omitted %q: %#v", want, expressions)
		}
	}
	if !first.GrammarComplete || first.CandidatesOmitted != 0 || first.CandidatesRetained != first.CandidatesEnumerated || first.GrammarCoveragePercent != 100 {
		t.Fatalf("complete bounded grammar coverage was misreported: %#v", first)
	}
}

func TestIRBodySearchCandidateGenerationReportsTruncationAsIncomplete(t *testing.T) {
	plan := generatedCandidateSearchPlan(2)
	receipt, err := generateIRBodySearchCandidates(&plan)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.GrammarComplete || receipt.CandidatesOmitted == 0 || receipt.CandidatesRetained != 2 ||
		receipt.GrammarCoveragePercent >= 100 || len(plan.Candidates) != 2 {
		t.Fatalf("candidate ceiling hid incomplete grammar coverage: plan=%#v receipt=%#v", plan, receipt)
	}
}

func TestGenerateWithIRBodySearchUsesGeneratedIRCandidatesAndScoresFinalBody(t *testing.T) {
	source := readIRBodySearchFixture(t)
	plan := generatedCandidateSearchPlan(16)
	plan.MaxAttempts = 16
	plan.HoldoutTestCases = []IRBodyFillTestCase{
		{Input: -9223372036854775808, Expected: 0},
		{Input: 9223372036854775807, Expected: 9223372036854775807},
	}
	var request searchRequestSnapshot
	server := newIRBodySearchServer(t, func(_ int, snapshot searchRequestSnapshot) string {
		request = snapshot
		return generatedSearchCandidateID("0")
	})
	defer server.Close()

	result, err := GenerateWithIRBodySearch(context.Background(), "generated-search.gooo", source,
		"ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodySearch
	if receipt == nil || receipt.CandidateGeneration == nil {
		t.Fatalf("automatic candidate generation evidence is missing: %#v", result.Report)
	}
	generated := receipt.CandidateGeneration
	if generated.CandidatesEnumerated < 2 || generated.CandidatesRetained != generated.CandidatesEnumerated ||
		generated.CandidateSetSHA256 == "" || generated.GrammarCoveragePercent != 100 || !generated.GrammarComplete {
		t.Fatalf("candidate grammar coverage was not fully and precisely reported: %#v", generated)
	}
	if receipt.SelectedExpression != "0" || receipt.TrainingAccuracyPercent == nil || *receipt.TrainingAccuracyPercent != 100 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 || !strings.Contains(result.Source, "output = 0") {
		t.Fatalf("generated candidate was not typed, filled, and measured: receipt=%#v source=%s", receipt, result.Source)
	}
	if request.State.Schema != irBodySearchStateSchema || !containsString(searchCandidateIDs(t, request.State.RemainingCandidates), generatedSearchCandidateID("0")) ||
		strings.Contains(request.Raw, `"expected"`) {
		t.Fatalf("Laya did not receive generated choices or holdout data leaked into choice context: %s", request.Raw)
	}
	grammar := searchCompletenessDimension(t, result.Report.CompletenessReceipt, "search_candidate_grammar_coverage")
	if grammar.Status != "PASS" || !containsString(result.Report.CompletenessReceipt.CoreDimensions, grammar.ID) {
		t.Fatalf("bounded grammar completeness was not represented as a distinct dimension: %#v", grammar)
	}
}

func TestIRBodySearchCandidateGenerationRejectsMixedOrInvalidSources(t *testing.T) {
	plan := generatedCandidateSearchPlan(8)
	plan.Candidates = []IRBodyFillCandidate{{ID: "manual", Expression: "input"}}
	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("candidate_generation accepted a simultaneous manual candidate list")
	}
	plan = generatedCandidateSearchPlan(1)
	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("candidate_generation accepted an unsearchable one-candidate ceiling")
	}
	plan = generatedCandidateSearchPlan(8)
	plan.CandidateGeneration.Grammar = "unbounded-go/v1"
	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("candidate_generation accepted an undeclared grammar")
	}
}

func generatedCandidateSearchPlan(maxCandidates int) IRBodySearchPlan {
	return IRBodySearchPlan{
		Schema: irBodySearchPlanSchema, Intent: "Return zero for negative integers and preserve zero or positive integers.",
		HoleID: "floor",
		CandidateGeneration: &IRBodySearchCandidateGeneration{
			Schema: bodySearchCandidateGenerationSchema, Grammar: bodySearchIntegerAffineGrammar,
			MaxCandidates: maxCandidates,
		},
		TestCases: []IRBodyFillTestCase{
			{Input: -2, Expected: 0}, {Input: -1, Expected: 0}, {Input: 0, Expected: 0},
			{Input: 1, Expected: 1}, {Input: 2, Expected: 2},
		},
		MaxAttempts: 5,
	}
}
