package bodycodegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGenerateWithSourceIRBodyFillUsesDeclaredCandidatesDeterministically(t *testing.T) {
	source := []byte("package sample\nnamespace sample\nentity Integer id \"sample://integer\"\n" +
		"activity Lift(Integer) -> Integer computes `let base = __GOOO_BODY_HOLE_seed__\n" +
		"let increment = __GOOO_BODY_HOLE_step__\nreturn base + increment` assembling {\n" +
		"source_fill intent \"Compose base and increment\" {\n" +
		"hole \"seed\"\nhole \"step\"\n" +
		"candidate \"add_one\" { fill \"seed\" \"input + 0\" fill \"step\" \"1\" }\n" +
		"candidate \"double\" { fill \"seed\" \"input * 2\" fill \"step\" \"0\" }\n" +
		"}\ncase \"0\" -> \"1\"\ncase \"2\" -> \"3\"\n}\n")
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "source-fill.gooo", source, "Lift", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.SelectedCandidateID != "add_one" || result.Report.BodyFill.FunctionalAccuracyPct != 100 {
		t.Fatalf("source fill did not produce a finite measured result: %+v", result.Report.BodyFill)
	}
	if strings.Contains(result.GoooSource, "source_fill") || strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatalf("generated Gooo source retained assembly instructions or holes:\n%s", result.GoooSource)
	}
	if !strings.Contains(result.Source, "base = (input + 0)") || !strings.Contains(result.Source, "increment int64 = 1") || !result.Report.TypecheckPassed {
		t.Fatalf("selected assignments were not compiled into the generated body: %s", result.Source)
	}
	replay, replayDiagnostics := syntax.Parse(result.GoooSource)
	if replayDiagnostics.HasErrors() {
		t.Fatalf("generated Gooo source cannot be replayed: %v\n%s", replayDiagnostics, result.GoooSource)
	}
	if replay.Declarations[1].(*syntax.ActivityDecl).Assembly != nil {
		t.Fatal("replay source still requires a model choice")
	}
}

func TestSourceDerivedAssignmentsReachLayaAsCompleteChoices(t *testing.T) {
	source := `package sample
namespace sample
entity Integer id "sample://integer"
activity Lift(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + ` assembling {
    source_fill intent "Compose input plus one from two derived parts." {
        hole "seed"
        hole "step"
        derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"
    }
    case "0" -> "1"
    case "2" -> "3"
}`
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	candidates, _, err := generateSourceFillCandidates(spec.FillPlan, spec.Cases)
	if err != nil {
		t.Fatal(err)
	}
	wanted := ""
	for _, candidate := range candidates {
		if candidate.Fills["seed"] == "input" && candidate.Fills["step"] == "1" {
			wanted = candidate.ID
			break
		}
	}
	if wanted == "" {
		t.Fatal("the declared grammar did not derive the input-plus-one assignment")
	}
	var observed struct {
		Candidates []IRBodyFillCandidateScore `json:"candidate_scores"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observed); err != nil {
			t.Errorf("decode complete Gooo assignments: %v", err)
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "source-fill-test-model", "routing": map[string]any{"model": "source-fill-test-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": wanted, "probabilities": map[string]float64{wanted: 1},
			}},
		})
	}))
	defer server.Close()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "derived-fill.gooo", []byte(source), "Lift", spec,
		server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.SelectedCandidateID != wanted ||
		result.Report.BodyFill.Decision.Provider != "laya" || result.Report.BodyFill.FunctionalAccuracyPct != 100 {
		t.Fatalf("Laya did not select and verify the source-derived complete assignment: %+v", result.Report.BodyFill)
	}
	if len(observed.Candidates) != len(candidates) || observed.Candidates[0].ID == "" {
		t.Fatalf("Laya did not receive the compiler-enumerated assignment set: got=%d want=%d", len(observed.Candidates), len(candidates))
	}
}

func TestSourceIRBodyFillMeasuresHoldoutAfterLayaSelection(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-holdout.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	candidates, _, err := generateSourceFillCandidates(spec.FillPlan, spec.Cases)
	if err != nil {
		t.Fatal(err)
	}
	wanted := ""
	for _, candidate := range candidates {
		if candidate.Fills["seed"] == "input" && candidate.Fills["step"] == "1" {
			wanted = candidate.ID
			break
		}
	}
	if wanted == "" {
		t.Fatal("training-only grammar did not derive the input-plus-one assignment")
	}
	observedTrainingCases := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			return
		}
		var state struct {
			TestCaseCount int `json:"test_case_count"`
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &state); err != nil {
			t.Errorf("decode Laya source-fill state: %v", err)
			return
		}
		observedTrainingCases = state.TestCaseCount
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "source-fill-holdout-test", "routing": map[string]any{"model": "source-fill-holdout-test"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": wanted, "probabilities": map[string]float64{wanted: 1},
			}},
		})
	}))
	defer server.Close()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "holdout-fill.gooo", source, "Lift", spec,
		server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.TestCasesTotal != 2 || fill.TestCasesPassed != 2 ||
		fill.HoldoutCasesTotal != 2 || fill.HoldoutCasesPassed != 2 || fill.HoldoutAccuracyPercent == nil || *fill.HoldoutAccuracyPercent != 100 {
		t.Fatalf("training and held-out accuracy were not measured independently: %+v", fill)
	}
	if observedTrainingCases != 2 || fill.HoldoutSuiteSHA256 == "" {
		t.Fatalf("Laya received holdout evidence or the report did not bind it: training_cases=%d fill=%+v", observedTrainingCases, fill)
	}
	overlap := spec.Clone()
	overlap.HoldoutCases[0].Input = overlap.Cases[0].Input
	if err := overlap.Validate(); err == nil || !strings.Contains(err.Error(), "also appears in training cases") {
		t.Fatalf("overlapping source training and holdout inputs were accepted: %v", err)
	}
	dimension := bodyFillDimension(result, "body_fill_holdout_accuracy")
	if dimension.Status != "PASS" || !containsString(result.Report.CompletenessReceipt.CoreDimensions, dimension.ID) {
		t.Fatalf("held-out accuracy is missing from completeness metrics: %+v", dimension)
	}
}

func TestSourceIRBodyFillHoldoutMismatchLowersCompleteness(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-holdout.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	originalFile, originalDiagnostics := syntax.Parse(string(source))
	if originalDiagnostics.HasErrors() {
		t.Fatal(originalDiagnostics)
	}
	originalSpec := originalFile.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	original, err := GenerateWithSourceIRBodyFill(context.Background(), "holdout-original.gooo", source, "Lift", originalSpec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	modified, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-holdout-partial.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(modified))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "holdout-mismatch.gooo", modified, "Lift", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.TestCasesPassed != 2 || fill.TestCasesTotal != 2 ||
		fill.HoldoutCasesPassed != 1 || fill.HoldoutCasesTotal != 2 || fill.HoldoutAccuracyPercent == nil || *fill.HoldoutAccuracyPercent != 50 {
		t.Fatalf("held-out mismatch was not isolated from training accuracy: %+v", fill)
	}
	originalFill := original.Report.BodyFill
	if originalFill == nil || fill.TestSuiteSHA256 != originalFill.TestSuiteSHA256 ||
		fill.IRPlanSHA256 == originalFill.IRPlanSHA256 || fill.HoldoutSuiteSHA256 == originalFill.HoldoutSuiteSHA256 ||
		fill.SelectedCandidateID != originalFill.SelectedCandidateID || !slices.Equal(fill.CandidateScores, originalFill.CandidateScores) {
		t.Fatalf("changing holdout evidence altered training selection or failed to change bound identities: original=%+v modified=%+v", originalFill, fill)
	}
	dimension := bodyFillDimension(result, "body_fill_holdout_accuracy")
	if dimension.Status != "PROGRESS" {
		t.Fatalf("partial holdout accuracy should remain visible as progress: %+v", dimension)
	}
}

func TestGenerateWithSourceIRBodyFillDerivesBoundedCompleteAssignments(t *testing.T) {
	source := `package sample
namespace sample
entity Integer id "sample://integer"
activity Lift(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + ` assembling {
    source_fill intent "Compose input plus one from two derived parts." {
        hole "seed"
        hole "step"
        derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"
    }
    case "0" -> "1"
    case "2" -> "3"
}`
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "derived-fill.gooo", []byte(source), "Lift", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.CandidateGeneration == nil || receipt.SelectedCandidateID == "" || receipt.FunctionalAccuracyPct != 100 {
		t.Fatalf("derived assignment was not generated, selected and measured: %+v", receipt)
	}
	generated := receipt.CandidateGeneration
	if generated.Grammar != "integer-offset-constant/v1" || !generated.GrammarComplete || generated.GrammarCoveragePercent != 100 ||
		generated.AssignmentSpaceSize <= uint64(generated.AssignmentsRetained) || generated.AssignmentsRetained != 16 ||
		generated.AssignmentsOmitted == 0 || generated.CandidateSetSHA256 == "" {
		t.Fatalf("bounded grammar or assignment truncation was not made explicit: %+v", generated)
	}
	grammarDimension := bodyFillDimension(result, "body_fill_candidate_grammar_coverage")
	assignmentDimension := bodyFillDimension(result, "body_fill_assignment_space_coverage")
	if grammarDimension.Status != "PASS" || assignmentDimension.Status != "PROGRESS" ||
		!containsString(result.Report.CompletenessReceipt.CoreDimensions, assignmentDimension.ID) {
		t.Fatalf("completeness receipt hid source-derived search bounds: grammar=%+v assignments=%+v", grammarDimension, assignmentDimension)
	}
	if strings.Contains(result.GoooSource, "derive grammar") || strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatalf("generated Gooo source retained unresolved generation or hole declarations:\n%s", result.GoooSource)
	}
}

func TestSourceDerivedPerHoleGrammarsComposeConditionalBody(t *testing.T) {
	source := `package sample
namespace sample
entity Integer id "sample://integer"
activity Lift(Integer) -> Integer computes ` + "`" + `if __GOOO_BODY_HOLE_condition__ { return __GOOO_BODY_HOLE_yes__ } else { return 0 }` + "`" + ` assembling {
    source_fill intent "Return one for nonpositive input; return zero otherwise." {
        hole "condition"
        hole "yes"
        derive assignments max_candidates "16" {
            hole "condition" grammar "integer-predicate/v1" max_expressions "8"
            hole "yes" grammar "integer-offset-constant/v1" max_expressions "2"
        }
    }
    case "-1" -> "1"
    case "0" -> "1"
    case "2" -> "0"
}`
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "conditional-fill.gooo", []byte(source), "Lift", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || !result.Report.TypecheckPassed {
		t.Fatalf("typed per-hole generation failed finite suite or Go typecheck: %+v", receipt)
	}
	generation := receipt.CandidateGeneration
	if generation == nil || generation.Grammar != "per-hole" || len(generation.HoleGrammars) != 2 ||
		generation.HoleGrammars[0].Grammar != "integer-predicate/v1" ||
		generation.HoleGrammars[1].Grammar != "integer-offset-constant/v1" ||
		generation.GrammarComplete || generation.AssignmentSpaceSize != 16 || generation.AssignmentsRetained != 16 ||
		generation.AssignmentsOmitted != 0 || generation.HoleGrammars[0].GrammarCoveragePercent >= 100 {
		t.Fatalf("per-hole grammar and full assignment space were not reported: %+v", generation)
	}
	if !strings.Contains(result.Source, "if input < 2") || !strings.Contains(result.Source, "return 1") {
		t.Fatalf("generated body did not compose the selected predicate and value: %s", result.Source)
	}
	grammarDimension := bodyFillDimension(result, "body_fill_candidate_grammar_coverage")
	assignmentDimension := bodyFillDimension(result, "body_fill_assignment_space_coverage")
	if grammarDimension.Status != "PROGRESS" || assignmentDimension.Status != "PASS" {
		t.Fatalf("completeness receipt did not account for the bounded search: %+v %+v", grammarDimension, assignmentDimension)
	}
}

func TestSourceDerivedPredicateCompositionFindsDisjointCases(t *testing.T) {
	candidates, total, complete, err := generateIntegerPredicateExpressions(8, []assemblyspec.Case{
		{Input: -2}, {Input: 0}, {Input: 2},
	}, true)
	if err != nil || complete || total != 29 || !slices.Contains(candidates, "(input >= -2) && (input <= 0)") {
		t.Fatalf("bounded grammar omitted its closed-range condition: candidates=%v total=%d complete=%v err=%v", candidates, total, complete, err)
	}
	source := `package sample
namespace sample
entity Integer id "sample://integer"
activity Select(Integer) -> Integer computes ` + "`" + `if __GOOO_BODY_HOLE_condition__ { return __GOOO_BODY_HOLE_yes__ } else { return 0 }` + "`" + ` assembling {
    source_fill intent "Return one for either selected input." {
        hole "condition"
        hole "yes"
        derive assignments max_candidates "16" {
            hole "condition" grammar "integer-predicate-composition/v1" max_expressions "8"
            hole "yes" grammar "integer-offset-constant/v1" max_expressions "2"
        }
    }
    case "-2" -> "1"
    case "0" -> "0"
    case "2" -> "1"
}`
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "composed-condition.gooo", []byte(source), "Select", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyFill == nil || result.Report.BodyFill.FunctionalAccuracyPct != 100 || !result.Report.TypecheckPassed {
		t.Fatalf("composed predicate did not satisfy declared cases and typecheck: %+v", result.Report.BodyFill)
	}
	generation := result.Report.BodyFill.CandidateGeneration
	if generation == nil || generation.Grammar != "per-hole" || generation.GrammarComplete || generation.HoleGrammars[0].Grammar != "integer-predicate-composition/v1" || generation.HoleGrammars[0].ExpressionCandidatesTotal <= 8 || generation.HoleGrammars[0].ExpressionsRetained != 8 {
		t.Fatalf("bounded composition grammar coverage was not measured: %+v", generation)
	}
	if !strings.Contains(result.Source, "||") {
		t.Fatalf("expected a disjoint-input predicate, got generated source: %s", result.Source)
	}
}

func TestSourceDerivedOutsideRangeFindsConditionalWindow(t *testing.T) {
	inputs := []assemblyspec.Case{{Input: -2}, {Input: 0}, {Input: 10}, {Input: 12}}
	conditions, total, complete, err := generateIntegerOutsideRangeExpressions(8, inputs)
	if err != nil || total != 24 || complete || len(conditions) != 8 ||
		!slices.Contains(conditions, "(input < 0) || (input > 10)") {
		t.Fatalf("outside-range grammar did not retain its bounded interval candidate: conditions=%v total=%d complete=%v err=%v", conditions, total, complete, err)
	}
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-outside-range.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "outside-range.gooo", source, "Select", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 4 || receipt.HoldoutCasesPassed != 3 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 || !result.Report.TypecheckPassed ||
		!strings.Contains(result.Source, "(input < 0) || (input > 10)") {
		t.Fatalf("outside-range composition did not typecheck and match train/holdout cases: report=%+v source=%s", receipt, result.Source)
	}
	if dimension := bodyFillDimension(result, "body_fill_holdout_accuracy"); dimension.Status != "PASS" {
		t.Fatalf("outside-range holdout completeness was not reported: %+v", dimension)
	}
}

func TestSourceDerivedPredicateCutpointsIncludeUnobservedIntegerBoundaries(t *testing.T) {
	conditions, total, complete, err := generateIntegerPredicateCutpointExpressions(8,
		[]assemblyspec.Case{{Input: -10}, {Input: 10}})
	if err != nil || total != 14 || complete || len(conditions) != 8 ||
		!slices.Contains(conditions, "input < 0") || !slices.Contains(conditions, "input <= 0") {
		t.Fatalf("cutpoint grammar did not derive both interior boundaries with honest coverage: conditions=%v total=%d complete=%v err=%v", conditions, total, complete, err)
	}
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-cutpoint.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "cutpoint.gooo", source, "NonPositive", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 2 || receipt.HoldoutCasesTotal != 2 ||
		receipt.HoldoutAccuracyPercent == nil || !result.Report.TypecheckPassed {
		t.Fatalf("cutpoint candidate composition did not preserve separate train and holdout evidence: report=%+v source=%s", receipt, result.Source)
	}
	holdout := bodyFillDimension(result, "body_fill_holdout_accuracy")
	if holdout.Denominator != 2 || (receipt.HoldoutCasesPassed == 2 && holdout.Status != "PASS") ||
		(receipt.HoldoutCasesPassed < 2 && holdout.Status != "PROGRESS") {
		t.Fatalf("held-out boundary score did not map to its four-state metric: receipt=%+v dimension=%+v", receipt, holdout)
	}
	probes := receipt.BehavioralProbes
	if probes == nil || probes.Schema != irBodyFillBehavioralProbeSchema ||
		!slices.Contains(probes.ProbeInputs, 0) || slices.Contains(probes.ProbeInputs, -10) || slices.Contains(probes.ProbeInputs, 10) ||
		probes.CandidateRunsCompleted != probes.CandidateCount || probes.CandidateRunsFailed != 0 ||
		probes.CandidatePairsEvaluated != probes.CandidatePairsTotal || probes.CandidatePairsDistinguished == 0 ||
		probes.ProbeInputsWithDisagreement == 0 || probes.ProbeInputsSHA256 == "" {
		t.Fatalf("automatic probes did not reveal and measure bounded candidate distinctions: %+v", probes)
	}
	wantDistinguishability := float64(probes.CandidatePairsDistinguished) * 100 / float64(probes.CandidatePairsEvaluated)
	if probes.CandidatePairDistinguishabilityPercent != wantDistinguishability {
		t.Fatalf("candidate pair distinguishability=%v, want %v", probes.CandidatePairDistinguishabilityPercent, wantDistinguishability)
	}
	probeCoverage := bodyFillDimension(result, "body_fill_candidate_probe_coverage")
	if probeCoverage.Status != "PASS" || probeCoverage.Numerator != probes.CandidateCount*len(probes.ProbeInputs) ||
		probeCoverage.Denominator != probes.CandidateCount*probes.ProbeInputsTotal ||
		!containsString(result.Report.CompletenessReceipt.CoreDimensions, probeCoverage.ID) {
		t.Fatalf("candidate probe evaluation coverage was not exposed as a completeness dimension: %+v", probeCoverage)
	}
}

func TestSourceDerivedPredicateCutpointsHandleInt64Edges(t *testing.T) {
	minInt64 := int64(-1 << 63)
	maxInt64 := int64(1<<63 - 1)
	tests := []struct {
		name             string
		inputs           []int64
		midpointLessThan []string
	}{
		{name: "full int64 span", inputs: []int64{minInt64, maxInt64}, midpointLessThan: []string{"input < -1", "input < 0"}},
		{name: "negative odd gap", inputs: []int64{-5, -2}, midpointLessThan: []string{"input < -4", "input < -3"}},
		{name: "upper edge", inputs: []int64{maxInt64 - 2, maxInt64}, midpointLessThan: []string{"input < 9223372036854775806"}},
		{name: "adjacent lower edge", inputs: []int64{minInt64, minInt64 + 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cases := []assemblyspec.Case{{Input: test.inputs[0]}, {Input: test.inputs[1]}}
			conditions, total, complete, err := generateIntegerPredicateCutpointExpressions(128, cases)
			if err != nil || !complete || len(conditions) != total {
				t.Fatalf("edge cutpoint grammar did not fully enumerate its finite expression set: count=%d total=%d complete=%v err=%v", len(conditions), total, complete, err)
			}
			seen := make(map[string]struct{}, len(conditions))
			for _, condition := range conditions {
				if _, duplicate := seen[condition]; duplicate {
					t.Fatalf("edge inputs produced duplicate expression %q", condition)
				}
				seen[condition] = struct{}{}
			}
			for _, expression := range test.midpointLessThan {
				if !slices.Contains(conditions, expression) {
					t.Errorf("missing safe midpoint expression %q in %v", expression, conditions)
				}
			}
			if wantTotal := (2+len(test.midpointLessThan))*4 + 2; total != wantTotal {
				t.Fatalf("finite grammar count=%d, want %d for endpoints plus midpoint thresholds", total, wantTotal)
			}
		})
	}
}

func TestLayaCanResolveTrainingTieWhileHoldoutRemainsWithheld(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-cutpoint.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	candidates, _, err := generateSourceFillCandidates(spec.FillPlan, spec.Cases)
	if err != nil {
		t.Fatal(err)
	}
	wanted := ""
	for _, candidate := range candidates {
		if candidate.Fills["condition"] == "input <= 0" && candidate.Fills["yes"] == "1" {
			wanted = candidate.ID
			break
		}
	}
	if wanted == "" {
		t.Fatal("the declared cutpoint grammar did not produce the intended nonpositive path")
	}
	var observed irBodyFillState
	var observedFields map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/systemone" {
			http.NotFound(writer, request)
			return
		}
		var payload struct {
			State map[string]string `json:"state"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode Laya request: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observed); err != nil {
			t.Errorf("decode training-only Gooo chooser state: %v", err)
			return
		}
		if err := json.Unmarshal([]byte(payload.State["request"]), &observedFields); err != nil {
			t.Errorf("decode chooser state fields: %v", err)
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"model": "cutpoint-test-model", "routing": map[string]any{"model": "cutpoint-test-model"},
			"answers": map[string]any{"body_ir_fill": map[string]any{
				"choice": wanted, "probabilities": map[string]float64{wanted: 1},
			}},
		})
	}))
	defer server.Close()

	result, err := GenerateWithSourceIRBodyFill(context.Background(), "cutpoint.gooo", source, "NonPositive", spec,
		server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	fill := result.Report.BodyFill
	if fill == nil || fill.Decision.Provider != "laya" || fill.SelectedCandidateID != wanted ||
		fill.TestCasesPassed != 2 || fill.HoldoutCasesPassed != 2 || fill.HoldoutCasesTotal != 2 ||
		fill.HoldoutAccuracyPercent == nil || *fill.HoldoutAccuracyPercent != 100 || !result.Report.TypecheckPassed {
		t.Fatalf("Laya-selected boundary did not pass typed generation and separate holdout evaluation: %+v", fill)
	}
	if observed.TestCaseCount != 2 || observed.TestSuiteSHA256 != fill.TestSuiteSHA256 ||
		len(observed.Candidates) != len(candidates) || !slices.ContainsFunc(observed.Candidates, func(candidate IRBodyFillCandidateScore) bool {
		return candidate.ID == wanted && candidate.TestCasesPassed == 2 && candidate.TestCasesTotal == 2
	}) {
		t.Fatalf("Laya did not receive the full tied candidate set with training-only evidence: observed=%+v", observed)
	}
	if _, hasHoldoutCases := observedFields["holdout_cases"]; hasHoldoutCases {
		t.Fatalf("Laya request included withheld cases: %s", observedFields["holdout_cases"])
	}
	if _, hasBehavioralProbes := observedFields["behavioral_probes"]; hasBehavioralProbes {
		t.Fatalf("post-selection behavioral probes were sent to Laya: %s", observedFields["behavioral_probes"])
	}
	if dimension := bodyFillDimension(result, "body_fill_holdout_accuracy"); dimension.Status != "PASS" {
		t.Fatalf("the model-selected generalization evidence was not reflected in completeness: %+v", dimension)
	}
}
