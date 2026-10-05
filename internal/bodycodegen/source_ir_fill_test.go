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
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-composed-condition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*syntax.ActivityDecl).Assembly.Spec.Clone()
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "composed-condition.gooo", source, "Select", spec, "", "", IRBodyFillOptions{})
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
