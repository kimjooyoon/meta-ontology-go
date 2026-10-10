package bodycodegen

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestSourceRecipeUnaryArithmetic(t *testing.T) {
	for _, tc := range []struct {
		body string
		want func(int64) int64
	}{
		{"return -input", func(x int64) int64 { return -x }},
		{"return -(input + 1)", func(x int64) int64 { return -(x + 1) }},
		{"return -(-input)", func(x int64) int64 { return -(-x) }},
		{"let value = -input; value = -value; return value", func(x int64) int64 { return x }},
		{"if -input > 0 { return -input } else { return input }", func(x int64) int64 {
			if -x > 0 {
				return -x
			}
			return x
		}},
	} {
		t.Run(tc.body, func(t *testing.T) {
			ctx := context.Background()
			source := recipeSource(tc.body)
			doc, err := DecodeSourcePathDocument(ctx, "unary.gooo", source, "Assemble", recipeBytes(recipeOperand))
			if err != nil {
				t.Fatal(err)
			}
			for _, input := range []int64{math.MinInt64, -9007199254740993, -1, 0, 1, 9007199254740993, math.MaxInt64} {
				doc.TestCases = append(doc.TestCases, pathplan.TestCase{Input: input, Expected: tc.want(input)})
			}
			doc.TestCases = doc.TestCases[1:]
			encoded, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			full, err := DecodeSourcePathDocument(ctx, "unary.gooo", source, "Assemble", encoded)
			if err != nil || !reflect.DeepEqual(doc, full) {
				t.Fatal("document changed", err)
			}
			result, err := GenerateWithTypedPaths(ctx, "unary.gooo", source, "Assemble", doc, "")
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.FunctionalCompleteness != 100 || p.Search.Selection.ModelCalls != 0 || !p.SourceBinding.Equivalent ||
				!strings.Contains(p.SourceBinding.Method, "normalized_arithmetic") {
				t.Fatal("arithmetic normalization or finite score missing", p)
			}
		})
	}
}

func TestUnaryTreeBindingMethods(t *testing.T) {
	for _, tc := range []struct {
		left, right, method string
		equal               bool
	}{
		{"return -input", "return 0-input", "normalized_arithmetic_typed_body_tree/v1", true},
		{"return -input", "return input-0", "normalized_arithmetic_typed_body_tree/v1", false},
		{"return -(input+1)", "return 0-(input+1)", "normalized_arithmetic_typed_body_tree/v1", true},
		{"if -input > 0 { return -input } else { return input }", "if 0 < (0-input) { return 0-input } else { return input }", "normalized_arithmetic_condition_typed_body_tree/v1", true},
		{"return input + -2", "return input + -2", "canonical_typed_body_tree/v1", true},
		{"return input + -9223372036854775808", "return input + -9223372036854775808", "canonical_typed_body_tree/v1", true},
	} {
		generated := []byte("package sample\nfunc Assemble(input int64) int64 {" + tc.right + "}")
		r, err := typedBodyTreeEquivalence(context.Background(), "Assemble", tc.left, generated)
		if err != nil || r.Equivalent != tc.equal || r.Method != tc.method {
			t.Fatal(tc, r, err)
		}
	}
}

func TestUnaryNegativeLiteralsKeepExistingArena(t *testing.T) {
	for _, n := range []int64{-2, math.MinInt64} {
		source := recipeSource("return input + " + strconv.FormatInt(n, 10))
		doc, err := DecodeSourcePathDocument(context.Background(), "literal.gooo", source, "Assemble", recipeBytes(recipeOperand))
		if err != nil {
			t.Fatal(err)
		}
		want := []bodyplan.Expr{{Kind: bodyplan.ExprInput, Name: "input"}, {Kind: bodyplan.ExprInt, Int: n},
			{Kind: bodyplan.ExprBinary, Operation: "add", Left: 0, Right: 1}}
		if !reflect.DeepEqual(doc.Plan.Base.Expressions, want) || doc.Plan.Decisions[0].Target != 2 {
			t.Fatal("old indices changed", doc)
		}
	}
	doc, err := DecodeSourcePathDocument(context.Background(), "zero.gooo", recipeSource("return input + -0"), "Assemble", recipeBytes(recipeOperand))
	if err != nil || len(doc.Plan.Base.Expressions) != 3 || doc.Plan.Decisions[0].Target != 2 || doc.Plan.Base.Expressions[1].Int != 0 {
		t.Fatal("negative zero changed the accepted arena", err)
	}
}

func TestUnaryArithmeticKeepsTypeAndBounds(t *testing.T) {
	for _, body := range []string{"return -true", "let value = false; return -value", "return " + strings.Repeat("-(", 20) + "input" + strings.Repeat(")", 20)} {
		if _, err := DecodeSourcePathDocument(context.Background(), "invalid.gooo", recipeSource(body), "Assemble", recipeBytes(recipeOperand)); err == nil {
			t.Fatal("invalid unary recipe accepted", body)
		}
	}
}

func TestUnaryCheckpointPreservesOriginalPalette(t *testing.T) {
	source := checkpointFixture(t, "let value = -input; value = -value; return value", []assemblyspec.Choice{
		{ID: "sign", Kind: "operand_order", Intent: "Choose the integer sign. 정수의 부호를 고른다."},
	})
	assertCheckpointPreservesDocument(t, source, map[string]string{"sign": "layout_forward"})
	assertCheckpointPreservesDocument(t, source, map[string]string{"sign": "layout_reverse"})
}

func TestUnaryRecipeSelectorsFollowSourceOrder(t *testing.T) {
	doc, err := DecodeSourcePathDocument(context.Background(), "nested.gooo", recipeSource("return -(input + 1)"), "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	outer := doc.Plan.Base.Expressions[doc.Plan.Decisions[0].Target]
	if outer.Operation != "subtract" || doc.Plan.Base.Expressions[outer.Left].Int != 0 || doc.Plan.Base.Expressions[outer.Right].Operation != "add" {
		t.Fatal("synthetic unary subtraction is not the outer source site", doc.Plan.Base.Expressions)
	}
	innerBytes := []byte(strings.Replace(string(recipeBytes(recipeOperand)), `"occurrence":0`, `"occurrence":1`, 1))
	innerDoc, err := DecodeSourcePathDocument(context.Background(), "nested.gooo", recipeSource("return -(input + 1)"), "Assemble", innerBytes)
	if err != nil || innerDoc.Plan.Base.Expressions[innerDoc.Plan.Decisions[0].Target].Operation != "add" {
		t.Fatal("nested source site changed", err)
	}
}
