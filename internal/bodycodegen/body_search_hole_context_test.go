package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func holeContextSource(body, cases string) []byte {
	return []byte(fmt.Sprintf("package holes\nnamespace holes\nentity Integer id \"holes://integer\"\nactivity Build(Integer) -> Integer computes `%s` assembling {\n    search hole \"value\" grammar \"integer-hole-residual/v1\" intent \"Fill the declared computation.\" max_candidates \"16\"\n%s\n    attempts \"16\"\n}\n", body, cases))
}

func TestHoleContextSearchSolvesNestedConstant(t *testing.T) {
	source := holeContextSource("return input + __GOOO_BODY_HOLE_value__", "case \"2\" -> \"3\"\ncase \"10\" -> \"11\"\nholdout_case \"100\" -> \"101\"")
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "nested.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRSearch(ctx, "nested.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodySearch
	if r.SelectedExpression != "1" || r.TrainingPassed != 2 || r.HoldoutPassed != 1 {
		t.Fatal("nested hole did not receive its required constant", r.SelectedExpression, r.TrainingPassed, r.HoldoutPassed)
	}
	probes := r.CandidateGeneration.HoleContext
	if probes == nil || probes.EvaluationCalls != 4 || len(probes.Probes) != 2 || *probes.Probes[0].Derived != 1 {
		t.Fatal("context observation is missing")
	}
	if err := VerifyIRBodySearchProjection(ctx, "nested.gooo", source, result); err != nil {
		t.Fatal("context-derived candidate failed replay", err)
	}
	old := []byte(strings.Replace(string(source), "integer-hole-residual/v1", "integer-offset-constant/v1", 1))
	oldSpec, err := SourceAssembly(ctx, "nested.gooo", old, "Build")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := GenerateWithSourceIRSearch(ctx, "nested.gooo", old, "Build", oldSpec, "", "")
	if err != nil || baseline.Report.BodySearch.TrainingPassed != 0 {
		t.Fatal("legacy grammar observation changed", err)
	}
}

func TestHoleContextSearchSupportsNestedArithmeticLocalsAndBranches(t *testing.T) {
	for _, test := range []struct{ name, body, cases, selected string }{
		{"weighted", "return (input + __GOOO_BODY_HOLE_value__) * 2", "case \"2\" -> \"10\"\ncase \"10\" -> \"26\"", "3"},
		{"subtract", "let saved = input\nreturn saved - __GOOO_BODY_HOLE_value__", "case \"2\" -> \"-1\"\ncase \"10\" -> \"7\"", "3"},
		{"multiply", "return input * __GOOO_BODY_HOLE_value__", "case \"2\" -> \"6\"\ncase \"10\" -> \"30\"", "3"},
		{"affine", "let saved = input\nlet fill = __GOOO_BODY_HOLE_value__\nreturn saved + fill", "case \"2\" -> \"13\"\ncase \"10\" -> \"53\"", "input * 4 + 3"},
		{"branch", "if input < 0 { return input }; return input + __GOOO_BODY_HOLE_value__", "case \"-4\" -> \"-4\"\ncase \"2\" -> \"3\"", "1"},
		{"large", "return input + __GOOO_BODY_HOLE_value__", "case \"0\" -> \"9007199254740993\"\ncase \"1\" -> \"9007199254740994\"", "9007199254740993"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := holeContextSource(test.body, test.cases)
			spec, err := SourceAssembly(context.Background(), "case.gooo", source, "Build")
			if err != nil {
				t.Fatal(err)
			}
			result, err := GenerateWithSourceIRSearch(context.Background(), "case.gooo", source, "Build", spec, "", "")
			if err != nil {
				t.Fatal(err)
			}
			if result.Report.BodySearch.SelectedExpression != test.selected || result.Report.BodySearch.TrainingPassed != 2 {
				t.Fatal("contextual construction differs", result.Report.BodySearch.SelectedExpression)
			}
		})
	}
}

func TestHoleContextSearchKeepsNonlinearGuessesPartial(t *testing.T) {
	source := holeContextSource("let factor = __GOOO_BODY_HOLE_value__\nreturn input + factor * factor", "case \"2\" -> \"7\"\ncase \"10\" -> \"15\"")
	spec, _ := SourceAssembly(context.Background(), "nonlinear.gooo", source, "Build")
	result, err := GenerateWithSourceIRSearch(context.Background(), "nonlinear.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodySearch
	if *r.CandidateGeneration.HoleContext.Probes[0].Derived != 5 || r.TrainingPassed != 0 || r.TrainingTotal != 2 {
		t.Fatal("two-point inference bypassed whole-body scoring")
	}
}

func TestHoleContextReceiptReplaysAndExcludesHoldouts(t *testing.T) {
	ctx := context.Background()
	source := holeContextSource("return input + __GOOO_BODY_HOLE_value__", "case \"2\" -> \"3\"\nholdout_case \"100\" -> \"101\"")
	spec, _ := SourceAssembly(ctx, "case.gooo", source, "Build")
	result, err := GenerateWithSourceIRSearch(ctx, "case.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(string(source), "\"100\" -> \"101\"", "\"111\" -> \"999\"", 1))
	changedSpec, _ := SourceAssembly(ctx, "case.gooo", changed, "Build")
	again, err := GenerateWithSourceIRSearch(ctx, "case.gooo", changed, "Build", changedSpec, "", "")
	if err != nil || !reflect.DeepEqual(result.Report.BodySearch.CandidateGeneration, again.Report.BodySearch.CandidateGeneration) {
		t.Fatal("holdout affected candidate construction", err)
	}
	value := int64(999)
	result.Report.BodySearch.CandidateGeneration.HoleContext.Probes[0].Zero = &value
	if err := VerifyIRBodySearchProjection(ctx, "case.gooo", source, result); err == nil {
		t.Fatal("altered probe evidence replayed")
	}
}
