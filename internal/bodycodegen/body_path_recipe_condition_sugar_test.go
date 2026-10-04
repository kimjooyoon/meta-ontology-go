package bodycodegen

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestSourceRecipeCommonConditions(t *testing.T) {
	for _, tc := range []struct {
		condition string
		want      func(int64) bool
	}{
		{"input > 0", func(x int64) bool { return x > 0 }},
		{"input >= 0", func(x int64) bool { return x >= 0 }},
		{"input != 0", func(x int64) bool { return x != 0 }},
		{"!(input <= 0)", func(x int64) bool { return !(x <= 0) }},
		{"!(input > 0) && input != -1", func(x int64) bool { return !(x > 0) && x != -1 }},
	} {
		t.Run(tc.condition, func(t *testing.T) {
			body := "let accepted = " + tc.condition + "; if accepted { return 1 } else { return 0 }"
			source := recipeSource(body)
			raw := recipeBytes(`{"id":"rule","kind":"branch_layout","occurrence":0,"intent":"조건을 충족하면 1을 반환한다. Return one when the condition holds."}`)
			doc, err := DecodeSourcePathDocument(context.Background(), "rule.gooo", source, "Assemble", raw)
			if err != nil {
				t.Fatal(err)
			}
			cases := make([]pathplan.TestCase, 0, 7)
			for _, x := range []int64{math.MinInt64, -2, -1, 0, 1, 2, math.MaxInt64} {
				expected := int64(0)
				if tc.want(x) {
					expected = 1
				}
				cases = append(cases, pathplan.TestCase{Input: x, Expected: expected})
			}
			doc.TestCases = cases
			result, err := GenerateWithTypedPaths(context.Background(), "rule.gooo", source, "Assemble", doc, "")
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.FunctionalCompleteness != 100 || p.Search.Selection.ModelCalls != 0 ||
				p.SourceBinding.Method != "normalized_condition_typed_body_tree/v1" || !p.SourceBinding.Equivalent {
				t.Fatal("condition assembly or source binding changed", p)
			}
		})
	}
}

func TestConditionNormalizationPreservesMeaning(t *testing.T) {
	for _, tc := range []struct {
		left, right string
		equal       bool
	}{
		{"input > 0", "0 < input", true}, {"input >= 0", "0 <= input", true},
		{"input != 0", "(input == 0) == false", true}, {"!(input < 0)", "(input < 0) == false", true},
		{"input > 0", "input < 0", false}, {"input >= 0", "input > 0", false},
		{"input != 0", "input == 0", false}, {"!(input < 0)", "input < 0", false},
	} {
		original := "if " + tc.left + " { return 1 } else { return 0 }"
		generated := []byte("package sample\nfunc Assemble(input int64) int64 { if " + tc.right + " { return 1 } else { return 0 } }")
		r, err := typedBodyTreeEquivalence(context.Background(), "Assemble", original, generated)
		if err != nil || r.Equivalent != tc.equal || r.Method != "normalized_condition_typed_body_tree/v1" {
			t.Fatal(tc, r, err)
		}
	}
}

func TestConditionNormalizationKeepsTypeAndNestingRules(t *testing.T) {
	raw := recipeBytes(`{"id":"rule","kind":"branch_layout","occurrence":0,"intent":"Keep the condition."}`)
	for _, condition := range []string{"!input", "input != true", "true > false", strings.Repeat("!", 20) + "true"} {
		source := recipeSource("if " + condition + " { return 1 } else { return 0 }")
		if _, err := DecodeSourcePathDocument(context.Background(), "invalid.gooo", source, "Assemble", raw); err == nil {
			t.Fatal("invalid condition assembled", condition)
		}
	}
}
