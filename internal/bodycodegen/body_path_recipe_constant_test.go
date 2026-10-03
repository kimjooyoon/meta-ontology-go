package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
)

func TestSourceRecipeConstantBodies(t *testing.T) {
	for _, body := range []string{
		"return 2 - 3",
		"let value = 2 - 3; return value",
		"let flag = true; let value = 2 - 3; if flag { return value } else { return 1 }",
	} {
		t.Run(body, func(t *testing.T) {
			source := recipeSource(body)
			raw := []byte(strings.Replace(string(recipeBytes(recipeOperand)), `"expected":2`, `"expected":1`, 1))
			doc, err := DecodeSourcePathDocument(context.Background(), "constant.gooo", source, "Assemble", raw)
			if err != nil {
				t.Fatal(err)
			}
			if last := doc.Plan.Base.Expressions[len(doc.Plan.Base.Expressions)-1]; last.Kind != bodyplan.ExprInput || last.Name != "input" {
				t.Fatal("declared input missing from constant body", last)
			}
			full, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeSourcePathDocument(context.Background(), "constant.gooo", source, "Assemble", full)
			if err != nil || !reflect.DeepEqual(doc, decoded) {
				t.Fatal("constant full-document round trip", err)
			}
			result, err := GenerateWithTypedPaths(context.Background(), "constant.gooo", source, "Assemble", doc, "")
			if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 ||
				result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || !strings.Contains(result.Source, "(3 - 2)") {
				t.Fatal("constant path did not assemble", err, result)
			}
		})
	}
}

func TestSourceRecipeExistingInputRepresentationStaysStable(t *testing.T) {
	doc, err := DecodeSourcePathDocument(context.Background(), "source.gooo", recipeSource("return input - 2"), "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	want := []bodyplan.Expr{{Kind: bodyplan.ExprInput, Name: "input"}, {Kind: bodyplan.ExprInt, Int: 2},
		{Kind: bodyplan.ExprBinary, Operation: "subtract", Left: 0, Right: 1}}
	if !reflect.DeepEqual(doc.Plan.Base.Expressions, want) || doc.Plan.Decisions[0].Target != 2 {
		t.Fatal("previously accepted expression indices changed", doc.Plan.Base.Expressions)
	}
}

func TestSourceRecipeSignatureInputHonorsArenaBound(t *testing.T) {
	b := recipeArena{ctx: context.Background(), expressionCount: 128}
	if err := b.includeDeclaredInput(); err == nil || b.expressionCount != 128 {
		t.Fatal("signature input exceeded expression cap", err)
	}
}
