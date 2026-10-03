package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const chainBody = "if input < 0 { return 0 } else if input < 10 { return input } else { return 10 }"
const blockBody = "if input < 0 { return 0 } else { if input < 10 { return input } else { return 10 } }"

func chainRecipe() []byte {
	return recipeBytes(`{"id":"inner","kind":"branch_layout","occurrence":1,"intent":"중간 범위는 입력을 반환한다. Return input in the middle range."}`)
}

func TestSourceRecipeConditionChainMatchesExplicitBlocks(t *testing.T) {
	var docs []pathplan.Document
	for _, body := range []string{chainBody, blockBody} {
		doc, err := DecodeSourcePathDocument(context.Background(), "range.gooo", recipeSource(body), "Assemble", chainRecipe())
		if err != nil {
			t.Fatal(err)
		}
		docs = append(docs, doc)
	}
	if !reflect.DeepEqual(docs[0], docs[1]) {
		t.Fatal("else-if and explicit else block produced different typed plans")
	}
	doc := docs[0]
	target := doc.Plan.Base.Statements[doc.Plan.Decisions[0].Target]
	condition := doc.Plan.Base.Expressions[target.Expr]
	if target.Kind != bodyplan.StmtIf || doc.Plan.Base.Expressions[condition.Right].Int != 10 {
		t.Fatal("occurrence one did not select the inner condition")
	}
	doc.TestCases = []pathplan.TestCase{{Input: -1, Expected: 0}, {Input: 0, Expected: 0},
		{Input: 9, Expected: 9}, {Input: 10, Expected: 10}, {Input: 11, Expected: 10}}
	result, err := GenerateWithTypedPaths(context.Background(), "range.gooo", recipeSource(chainBody), "Assemble", doc, "")
	if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 ||
		result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || result.Report.BodyPaths.SourceBinding.Method != "canonical_typed_body_tree/v1" {
		t.Fatal("condition-chain source did not assemble", err, result)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSourcePathDocument(context.Background(), "range.gooo", recipeSource(chainBody), "Assemble", raw)
	if err != nil || !reflect.DeepEqual(doc, decoded) {
		t.Fatal("full document changed the chain", err)
	}
}

func TestSourceRecipeConditionChainPreservesSharedLocalAssignment(t *testing.T) {
	body := "let value = input; if input < 0 { value = 0 } else if input < 10 { value = input } else { value = 10 }; return value"
	doc, err := DecodeSourcePathDocument(context.Background(), "range.gooo", recipeSource(body), "Assemble", chainRecipe())
	if err != nil {
		t.Fatal(err)
	}
	doc.TestCases = []pathplan.TestCase{{Input: -3, Expected: 0}, {Input: 5, Expected: 5}, {Input: 12, Expected: 10}}
	if _, err = GenerateWithTypedPaths(context.Background(), "range.gooo", recipeSource(body), "Assemble", doc, ""); err != nil {
		t.Fatal(err)
	}
}

func TestSourceRecipeConditionChainRetainsScopeAndDepthBounds(t *testing.T) {
	for _, body := range []string{
		"if input < 0 { let hidden = input; return hidden } else if input < 10 { return hidden } else { return 10 }",
		strings.Repeat("if input < 0 { return 0 } else ", 20) + "{ return input }",
	} {
		if _, err := DecodeSourcePathDocument(context.Background(), "range.gooo", recipeSource(body), "Assemble", chainRecipe()); err == nil {
			t.Fatal("chain escaped scope or nesting bound")
		}
	}
}
