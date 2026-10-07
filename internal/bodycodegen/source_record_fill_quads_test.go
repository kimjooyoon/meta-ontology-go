package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRecordRelationCompositionV3CountsFiniteGrammarExactly(t *testing.T) {
	for _, test := range []struct {
		atomCount int
		want      int
	}{
		{atomCount: 2, want: 4},
		{atomCount: 3, want: 17},
		{atomCount: 4, want: 88},
		{atomCount: 6, want: 796},
		{atomCount: 12, want: 21704},
	} {
		got, err := recordFieldRelationCompositionV3Count(test.atomCount)
		if err != nil || got != test.want {
			t.Fatalf("v3 grammar count for %d atoms = %d, %v; want %d", test.atomCount, got, err, test.want)
		}
	}
	if _, err := recordFieldRelationCompositionV3Count(int(^uint(0) >> 1)); err == nil {
		t.Fatal("overflowing four-relation grammar count was accepted")
	}
}

func TestRecordRelationCompositionV3EnumeratesAllFourAtomShapesAndFallbacks(t *testing.T) {
	source := []byte(`package relationquadgrammar
namespace relationquadgrammar
entity Left id "relationquadgrammar://left" fields {
    field key id "relationquadgrammar://left/key" type string required one
    field active id "relationquadgrammar://left/active" type boolean required one
}
entity Right id "relationquadgrammar://right" fields {
    field key id "relationquadgrammar://right/key" type string required one
    field active id "relationquadgrammar://right/active" type boolean required one
}
entity Decision id "relationquadgrammar://decision"
activity Compare(Left, Right) -> Decision computes "return Decision{}"
`)
	file, diagnostics := ParseBodyFile("relationquadgrammar.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	_, records, err := resolveBodyModel(file)
	if err != nil {
		t.Fatal(err)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == "Compare" {
			activity = candidate
			break
		}
	}
	if activity == nil {
		t.Fatal("record relation activity not found")
	}
	context := recordFillGrammarContext{activity: activity, records: records, output: recordTypeByName(records, activity.Output)}
	grammar := assemblyspec.FillHoleGrammar{Grammar: recordFieldRelationCompositionV3Grammar, MaxExpressions: 88}
	expressions, total, complete, err := context.expressions(grammar)
	if err != nil {
		t.Fatal(err)
	}
	if total != 88 || len(expressions) != total || !complete {
		t.Fatalf("complete four-atom grammar = %d/%d complete=%t", len(expressions), total, complete)
	}
	quadruples := make(map[string]bool, 40)
	for _, expression := range expressions[:40] {
		relations := strings.Count(expression, " == ") + strings.Count(expression, " != ")
		operators := strings.Count(expression, " && ") + strings.Count(expression, " || ")
		if relations != 4 || operators != 3 {
			t.Fatalf("four-atom candidate has %d relations and %d Boolean operators: %s", relations, operators, expression)
		}
		quadruples[expression] = true
	}
	if len(quadruples) != 40 {
		t.Fatalf("four-atom portion contains %d unique expressions, want 5 tree shapes × 8 operator combinations", len(quadruples))
	}

	truncated, truncatedTotal, truncatedComplete, err := context.expressions(
		assemblyspec.FillHoleGrammar{Grammar: recordFieldRelationCompositionV3Grammar, MaxExpressions: 8})
	if err != nil {
		t.Fatal(err)
	}
	replay, replayTotal, replayComplete, err := context.expressions(
		assemblyspec.FillHoleGrammar{Grammar: recordFieldRelationCompositionV3Grammar, MaxExpressions: 8})
	if err != nil {
		t.Fatal(err)
	}
	if truncatedTotal != 88 || replayTotal != truncatedTotal || truncatedComplete || replayComplete ||
		len(truncated) != 8 || !equalRecordExpressionPrefix(truncated, replay) {
		t.Fatalf("bounded v3 prefix is not exact and deterministic: first=%d/%d complete=%t replay=%d/%d complete=%t",
			len(truncated), truncatedTotal, truncatedComplete, len(replay), replayTotal, replayComplete)
	}
}

func TestSourceRecordIRBodyFillComposesFourRelationsAcrossRecordInputs(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-relation-quads.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("record-relation-quads.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == "MatchAll" {
			activity = candidate
			break
		}
	}
	if activity == nil {
		t.Fatal("four-relation body-fill activity not found")
	}
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "record-relation-quads.gooo", source,
		"MatchAll", &activity.Assembly.Spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyFill
	if receipt == nil || receipt.FunctionalAccuracyPct != 100 || receipt.TestCasesPassed != 4 || receipt.TestCasesTotal != 4 ||
		receipt.HoldoutAccuracyPercent == nil || *receipt.HoldoutAccuracyPercent != 100 ||
		receipt.HoldoutCasesPassed != 2 || receipt.HoldoutCasesTotal != 2 ||
		!result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("four-relation record fill did not generate and replay: receipt=%+v report=%+v", receipt, result.Report)
	}
	generation := receipt.CandidateGeneration
	if generation == nil || len(generation.HoleGrammars) != 2 ||
		generation.HoleGrammars[0].Grammar != recordFieldRelationCompositionV3Grammar ||
		generation.HoleGrammars[0].ExpressionCandidatesTotal != 21704 ||
		generation.HoleGrammars[0].ExpressionsRetained != 8 || generation.HoleGrammars[0].GrammarComplete ||
		generation.AssignmentSpaceSize != 16 || generation.AssignmentsRetained != 16 ||
		generation.AssignmentsOmitted != 0 || generation.AssignmentCoveragePercent != 100 {
		t.Fatalf("four-relation candidate and assignment completeness was not exact: %+v", generation)
	}
	var selectedCondition string
	for _, fill := range receipt.HoleFills {
		if fill.HoleID == "condition" {
			selectedCondition = fill.Expression
		}
	}
	if strings.Count(selectedCondition, " == ")+strings.Count(selectedCondition, " != ") != 4 ||
		!strings.Contains(result.GoooSource, selectedCondition) ||
		strings.Contains(result.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(result.GoooSource, "source_fill") {
		t.Fatalf("generated four-relation Gooo source is incomplete: condition=%q source=%s", selectedCondition, result.GoooSource)
	}
}

func equalRecordExpressionPrefix(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
