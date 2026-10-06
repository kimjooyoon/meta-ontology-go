package bodycodegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func readRecordBodyFillFixture(t *testing.T) ([]byte, *syntax.ActivityDecl) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity.Name == "ReviewCandidate" {
			return source, activity
		}
	}
	t.Fatal("record body-fill activity not found")
	return nil, nil
}

func readDerivedRecordBodyFillFixture(t *testing.T) ([]byte, *syntax.ActivityDecl) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-derived.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill-derived.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity.Name == "ReviewCandidate" {
			return source, activity
		}
	}
	t.Fatal("derived record body-fill activity not found")
	return nil, nil
}

func TestSourceRecordIRBodyFillLetsLayaSelectTypedDomainLogic(t *testing.T) {
	source, activity := readRecordBodyFillFixture(t)
	var observed struct {
		InputType  string                     `json:"input_type"`
		OutputType string                     `json:"output_type"`
		BodyIR     string                     `json:"body_ir"`
		TestCases  int                        `json:"test_case_count"`
		Candidates []IRBodyFillCandidateScore `json:"candidate_scores"`
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		calls++
		var payload struct {
			Questions map[string]struct {
				Options []struct {
					ID string `json:"id"`
				} `json:"options"`
			} `json:"questions"`
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			http.Error(writer, "bad request", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observed); err != nil {
			t.Errorf("decode Gooo record-fill state: %v", err)
			http.Error(writer, "bad state", http.StatusBadRequest)
			return
		}
		if !strings.Contains(observed.BodyIR, "__GOOO_BODY_HOLE_condition__") ||
			observed.InputType != "Candidate" || observed.OutputType != "Review" || observed.TestCases != 2 ||
			len(observed.Candidates) != 2 || strings.Contains(payload.State["request"], `"candidate_id":1`) {
			t.Errorf("Laya did not receive typed source state with scores but without raw value cases: %#v", observed)
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "record-fill-model", "routing": map[string]any{"model": "record-fill-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": "ready_is_accepted", "probabilities": map[string]float64{"ready_is_accepted": 1},
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if calls != 1 || receipt == nil || receipt.Decision.Provider != "laya" ||
		receipt.SelectedCandidateID != "ready_is_accepted" || receipt.ProposedAccuracyPct != 100 ||
		receipt.TestCasesPassed != 2 || receipt.TestCasesTotal != 2 || receipt.FunctionalAccuracyPct != 100 ||
		len(receipt.SelectedValueCaseResults) != 2 || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("Laya-selected record body was not independently checked: calls=%d receipt=%+v report=%+v", calls, receipt, result.Report)
	}
	if receipt.CandidateScores[1].AccuracyPercent != 0 || receipt.SelectedValueCaseResults[0].Actual == nil ||
		strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(result.GoooSource, "source_fill") {
		t.Fatalf("record choices, outcomes or resulting source were incomplete: source=%s receipt=%+v", result.GoooSource, receipt)
	}
}

func TestSourceRecordIRBodyFillUsesDeterministicBestCaseFallback(t *testing.T) {
	source, activity := readRecordBodyFillFixture(t)
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.Decision.Provider != "deterministic" || receipt.ExternalProviderCalls != 0 ||
		receipt.SelectedCandidateID != "ready_is_accepted" || receipt.FunctionalAccuracyPct != 100 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("record fill fallback did not choose the best finite candidate: %+v", receipt)
	}
}

func TestSourceRecordIRBodyFillRejectsIllTypedCandidateBeforeLaya(t *testing.T) {
	source, activity := readRecordBodyFillFixture(t)
	spec := activity.Assembly.Spec.Clone()
	for index := range spec.FillPlan.Candidates[0].Fills {
		if spec.FillPlan.Candidates[0].Fills[index].HoleID == "condition" {
			spec.FillPlan.Candidates[0].Fills[index].Expression = `input.candidate_id`
		}
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		http.Error(writer, "model must not see an ill-typed candidate", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill.gooo", source,
		"ReviewCandidate", spec, server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err == nil || !strings.Contains(err.Error(), "candidate \"ready_is_accepted\"") || calls != 0 {
		t.Fatalf("ill-typed candidate was not rejected before model selection: calls=%d err=%v", calls, err)
	}
}

func TestSourceRecordIRBodyFillDerivesTypedCandidatesFromValueCases(t *testing.T) {
	source, activity := readDerivedRecordBodyFillFixture(t)
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill-derived.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.CandidateGeneration == nil || receipt.FunctionalAccuracyPct != 100 ||
		receipt.TestCasesPassed != 2 || receipt.TestCasesTotal != 2 || receipt.HoldoutCasesPassed != 2 ||
		receipt.HoldoutCasesTotal != 2 || receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		len(receipt.SelectedValueHoldoutCaseResults) != 2 || len(receipt.CandidateScores) != 16 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("derived record candidates did not compose and replay: %+v report=%+v", receipt, result.Report)
	}
	generation := receipt.CandidateGeneration
	if generation.Grammar != "per-hole" || generation.AssignmentSpaceSize != 16 || generation.AssignmentsRetained != 16 ||
		generation.AssignmentsOmitted != 0 || generation.AssignmentCoveragePercent != 100 || generation.GrammarComplete ||
		len(generation.HoleGrammars) != 3 || generation.HoleGrammars[0].Grammar != recordFieldPredicateGrammar ||
		generation.HoleGrammars[0].ExpressionCandidatesTotal != 8 || generation.HoleGrammars[0].ExpressionsRetained != 4 ||
		generation.HoleGrammars[0].GrammarCoveragePercent != 50 ||
		generation.HoleGrammars[1].Grammar != recordStringLiteralGrammar || generation.HoleGrammars[2].Grammar != recordStringLiteralGrammar {
		t.Fatalf("record-derived grammar coverage was not reported exactly: %+v", generation)
	}
	if !strings.Contains(result.GoooSource, `input.state == \"ready\"`) {
		t.Fatalf("derived Gooo body did not use the observed record field: %s", result.GoooSource)
	}
	if strings.Contains(result.GoooSource, "derive assignments") || strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") ||
		strings.Contains(result.GoooSource, "source_fill") {
		t.Fatalf("generated Gooo source retained meta instructions: %s", result.GoooSource)
	}
}

func TestSourceRecordIRBodyFillDerivesIntegerBoundaryPredicates(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-integer-boundary.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill-integer-boundary.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == "ReviewCandidate" {
			activity = candidate
			break
		}
	}
	if activity == nil {
		t.Fatal("integer-boundary record body-fill activity not found")
	}

	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill-integer-boundary.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 5 || receipt.TestCasesTotal != 5 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		receipt.HoldoutCasesPassed != 2 || receipt.HoldoutCasesTotal != 2 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("integer-boundary record predicate did not generate and replay: receipt=%+v report=%+v", receipt, result.Report)
	}
	generation := receipt.CandidateGeneration
	if generation == nil || len(generation.HoleGrammars) != 2 ||
		generation.HoleGrammars[1].Grammar != recordFieldPredicateV2Grammar ||
		generation.HoleGrammars[1].ExpressionCandidatesTotal != 30 ||
		generation.HoleGrammars[1].ExpressionsRetained != 16 || generation.HoleGrammars[1].GrammarComplete ||
		generation.AssignmentSpaceSize != 32 || generation.AssignmentsRetained != 16 ||
		generation.AssignmentsOmitted != 16 || generation.AssignmentCoveragePercent != 50 {
		t.Fatalf("ordered predicate candidate coverage was not reported: %+v", generation)
	}
	var selectedCondition string
	for _, fill := range receipt.HoleFills {
		if fill.HoleID == "condition" {
			selectedCondition = fill.Expression
		}
	}
	if selectedCondition != "input.score > 0" {
		t.Fatalf("selected record boundary = %q, want input.score > 0; receipt=%+v", selectedCondition, receipt)
	}
	if strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(result.GoooSource, "source_fill") {
		t.Fatalf("generated source retained body-fill instructions: %s", result.GoooSource)
	}
}

func TestSourceRecordIRBodyFillDerivesPairwiseComposedConditions(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-composed.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill-composed.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == "ReviewCandidate" {
			activity = candidate
			break
		}
	}
	if activity == nil {
		t.Fatal("composed record body-fill activity not found")
	}
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill-composed.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 2 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("composed record predicate did not generate and replay: receipt=%+v report=%+v", receipt, result.Report)
	}
	generation := receipt.CandidateGeneration
	if generation == nil || generation.HoleGrammars[0].Grammar != recordPredicateCompositionGrammar ||
		generation.HoleGrammars[0].ExpressionCandidatesTotal != 100 || generation.HoleGrammars[0].ExpressionsRetained != 8 ||
		generation.AssignmentSpaceSize != 32 || generation.AssignmentsRetained != 16 || generation.AssignmentCoveragePercent != 50 {
		t.Fatalf("pairwise grammar and assignment completeness were not reported: %+v", generation)
	}
	if !strings.Contains(result.GoooSource, `input.state == \"ready\"`) ||
		!strings.Contains(result.GoooSource, "input.reviewed == true") || !strings.Contains(result.GoooSource, "&&") {
		t.Fatalf("generated Gooo body did not compose the two typed record fields: %s", result.GoooSource)
	}
}

func TestSourceRecordIRBodyFillDerivesIntegerRangeComposition(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-integer-range.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-fill-integer-range.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == "WithinPreferredRange" {
			activity = candidate
			break
		}
	}
	if activity == nil {
		t.Fatal("integer-range record body-fill activity not found")
	}
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill-integer-range.gooo", source,
		"WithinPreferredRange", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 4 || receipt.TestCasesTotal != 4 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		receipt.HoldoutCasesPassed != 2 || receipt.HoldoutCasesTotal != 2 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("integer-range record predicate did not generate and replay: receipt=%+v report=%+v", receipt, result.Report)
	}
	generation := receipt.CandidateGeneration
	if generation == nil || len(generation.HoleGrammars) != 3 ||
		generation.HoleGrammars[0].Grammar != recordPredicateCompositionV2Grammar ||
		generation.HoleGrammars[0].ExpressionCandidatesTotal != 576 ||
		generation.HoleGrammars[0].ExpressionsRetained != 16 || generation.HoleGrammars[0].GrammarComplete ||
		generation.HoleGrammars[1].Grammar != recordStringLiteralGrammar ||
		generation.HoleGrammars[1].ExpressionCandidatesTotal != 1 ||
		generation.HoleGrammars[1].ExpressionsRetained != 1 || !generation.HoleGrammars[1].GrammarComplete ||
		generation.HoleGrammars[2].Grammar != recordStringLiteralGrammar ||
		generation.HoleGrammars[2].ExpressionCandidatesTotal != 1 ||
		generation.HoleGrammars[2].ExpressionsRetained != 1 || !generation.HoleGrammars[2].GrammarComplete ||
		generation.AssignmentSpaceSize != 16 || generation.AssignmentsRetained != 16 || generation.AssignmentsOmitted != 0 {
		t.Fatalf("bounded range grammar completeness was not reported: %+v", generation)
	}
	var selectedCondition string
	for _, fill := range receipt.HoleFills {
		if fill.HoleID == "condition" {
			selectedCondition = fill.Expression
		}
	}
	if !strings.Contains(selectedCondition, " && ") {
		t.Fatalf("generated body did not select a bounded integer interval: %q; receipt=%+v", selectedCondition, receipt)
	}
	if strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(result.GoooSource, "source_fill") {
		t.Fatalf("generated Gooo source retained body-fill instructions: %s", result.GoooSource)
	}
}

func TestSourceRecordIRBodyFillLetsLayaRankDerivedCandidates(t *testing.T) {
	source, activity := readDerivedRecordBodyFillFixture(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && request.URL.Path == "/health" {
			_ = json.NewEncoder(writer).Encode(map[string]any{"revisions": map[string]string{"record-grammar-model": "0123456789abcdef0123456789abcdef"}})
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.Error(writer, "invalid derived candidate request", http.StatusNotFound)
			return
		}
		calls++
		if json.NewDecoder(request.Body).Decode(&payload) != nil {
			http.Error(writer, "invalid derived candidate request", http.StatusBadRequest)
			return
		}
		var state struct {
			TestCaseCount int                        `json:"test_case_count"`
			Candidates    []IRBodyFillCandidateScore `json:"candidate_scores"`
		}
		if json.Unmarshal([]byte(payload.State["request"]), &state) != nil ||
			state.TestCaseCount != 2 || len(state.Candidates) != 16 || strings.Contains(payload.State["request"], `"candidate_id":41`) ||
			strings.Contains(payload.State["request"], `"candidate_id":43`) || strings.Contains(payload.State["request"], `"candidate_id":44`) {
			http.Error(writer, "model request omitted derived candidates or included raw cases", http.StatusBadRequest)
			return
		}
		best := state.Candidates[0]
		for _, candidate := range state.Candidates[1:] {
			if candidate.AccuracyPercent > best.AccuracyPercent {
				best = candidate
			}
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "record-grammar-model", "routing": map[string]any{"model": "record-grammar-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": best.ID, "probabilities": map[string]float64{best.ID: 1},
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-fill-derived.gooo", source,
		"ReviewCandidate", &activity.Assembly.Spec, server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || result.Report.BodyFill == nil || result.Report.BodyFill.CandidateGeneration == nil ||
		result.Report.BodyFill.Decision.Provider != "laya" || result.Report.BodyFill.FunctionalAccuracyPct != 100 ||
		result.Report.BodyFill.HoldoutCasesPassed != 2 || result.Report.BodyFill.HoldoutCasesTotal != 2 ||
		result.Report.BodyFill.HoldoutAccuracyPercent == nil || *result.Report.BodyFill.HoldoutAccuracyPercent != 100 ||
		result.Report.BodyFill.ProposedCandidateID == "" || result.Report.BodyFill.SelectedCandidateID != result.Report.BodyFill.ProposedCandidateID {
		t.Fatalf("Laya did not select from the derived record grammar: calls=%d receipt=%+v", calls, result.Report.BodyFill)
	}
}
