package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func recipeSource(body string) []byte {
	return []byte("package sample\nnamespace sample\nentity Integer id \"sample://integer\"\n" +
		fmt.Sprintf("activity Assemble(Integer) -> Integer computes %q\n", body))
}

func recipeBytes(choices string) []byte {
	return []byte(`{"schema":"gooo/source-typed-path-recipe/v1","choices":[` + choices +
		`],"test_cases":[{"input":0,"expected":2}],"max_attempts":64}`)
}

const recipeOperand = `{"id":"operands","kind":"operand_order","occurrence":0,"intent":"2에서 입력을 뺀다. Subtract input from two."}`

func TestSourceRecipeComposesExistingSourceAndFullDocument(t *testing.T) {
	source := recipeSource("return input - 2")
	doc, err := DecodeSourcePathDocument(context.Background(), "source.gooo", source, "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(doc)
	full, err := DecodeSourcePathDocument(context.Background(), "source.gooo", source, "Assemble", raw)
	if err != nil || !reflect.DeepEqual(full, doc) {
		t.Fatal("full document changed", err)
	}
	result, err := GenerateWithTypedPaths(context.Background(), "source.gooo", source, "Assemble", doc, "")
	if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 ||
		result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || !strings.Contains(result.Source, "return (2 - input)") {
		t.Fatal("recipe did not reach existing construction", err, result)
	}
	if result.Report.BodyPaths.SourceBinding.Method != "canonical_typed_body_tree/v1" {
		t.Fatal("unparenthesized source requires explicit typed tree evidence")
	}
}

func TestSourceRecipeAllFiveStructuralKindsAndSourceOrder(t *testing.T) {
	source := recipeSource("let first = input + 2; let second = input * 2; let chosen = first; " +
		"if input < 0 { chosen = chosen + 3 } else { chosen = chosen - 3 }; return chosen + (first + second)")
	choices := recipeOperand + `,
	{"id":"reference","kind":"local_reference","occurrence":0,"intent":"choose local","alternative_name":"second"},
	{"id":"assignment","kind":"assignment_target","occurrence":0,"intent":"choose assignment","alternative_name":"first"},
	{"id":"branch","kind":"branch_layout","occurrence":0,"intent":"choose branch"},
	{"id":"order","kind":"root_order","occurrence":0,"intent":"choose order"}`
	doc, err := DecodeSourcePathDocument(context.Background(), "source.gooo", source, "Assemble", recipeBytes(choices))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Plan.Decisions) != 5 || doc.Plan.Base.Expressions[doc.Plan.Decisions[0].Target].Operation != "add" ||
		doc.Plan.Base.Expressions[doc.Plan.Decisions[1].Target].Name != "first" ||
		doc.Plan.Base.Statements[doc.Plan.Decisions[2].Target].Name != "chosen" {
		t.Fatal("source-relative selectors changed", doc.Plan.Decisions)
	}
	doc.TestCases = []pathplan.TestCase{{Input: -1, Expected: 3}, {Input: 0, Expected: 1}, {Input: 2, Expected: 9}}
	result, err := GenerateWithTypedPaths(context.Background(), "source.gooo", source, "Assemble", doc, "")
	if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 {
		t.Fatal("conditional source did not replay", err)
	}
}

func TestSourceRecipeNestedOperatorSourceOrder(t *testing.T) {
	doc, err := DecodeSourcePathDocument(context.Background(), "source.gooo",
		recipeSource("return input - (2 - 3)"), "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	e := doc.Plan.Base.Expressions[doc.Plan.Decisions[0].Target]
	if doc.Plan.Base.Expressions[e.Left].Kind != "input" || doc.Plan.Base.Expressions[e.Right].Kind != "binary" {
		t.Fatal("outer source operator was not first")
	}
}

func TestSourceRecipeStrictRejection(t *testing.T) {
	valid := string(recipeBytes(recipeOperand))
	for _, raw := range []string{valid + "{}", strings.Replace(valid, `"occurrence":0`, `"occurrence":null`, 1),
		strings.Replace(valid, `"occurrence":0,`, "", 1), strings.Replace(valid, `"occurrence":0`, `"occurrence":-1`, 1),
		strings.Replace(valid, `"occurrence":0`, `"occurrence":128`, 1), strings.Replace(valid, `"occurrence":0`, `"occurrence":3`, 1),
		strings.Replace(valid, `"occurrence":0`, `"Occurrence":0`, 1), strings.Replace(valid, `"occurrence":0`, `"occurrence":0,"occurrence":1`, 1),
		strings.Replace(valid, `"intent":`, `"unknown":`, 1), strings.Replace(valid, `"expected":2`, `"expected":null`, 1),
		strings.Replace(valid, `"max_attempts":64`, `"max_attempts":65`, 1),
		strings.Replace(valid, `"kind":"operand_order"`, `"kind":"unknown"`, 1),
		strings.Replace(valid, `"occurrence":0`, `"occurrence":0,"alternative_name":"input"`, 1)} {
		if _, err := DecodeSourcePathDocument(context.Background(), "source.gooo", recipeSource("return input - 2"), "Assemble", []byte(raw)); err == nil {
			t.Fatal("invalid recipe accepted", raw)
		}
	}
	for _, body := range []string{"return -input", "return input / 2", "var local int8 = 1; return input + (0 * 1)",
		"let value = input - 2; return value"} {
		raw := recipeBytes(recipeOperand)
		if strings.HasPrefix(body, "let") {
			raw = recipeBytes(`{"id":"reference","kind":"local_reference","occurrence":0,"intent":"select","alternative_name":"missing"}`)
		}
		if _, err := DecodeSourcePathDocument(context.Background(), "source.gooo", recipeSource(body), "Assemble", raw); err == nil {
			t.Fatal("unsupported source or missing local accepted", body)
		}
	}
}

func TestSourceRecipeCancellationAndOwnership(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DecodeSourcePathDocument(ctx, "source.gooo", recipeSource("return input - 2"), "Assemble", recipeBytes(recipeOperand)); err == nil {
		t.Fatal("cancellation ignored")
	}
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			doc, err := DecodeSourcePathDocument(context.Background(), "source.gooo", recipeSource("return input - 2"), "Assemble", recipeBytes(recipeOperand))
			if err != nil {
				t.Error(err)
				return
			}
			doc.Plan.Base.Expressions[0].Name = "request-owned"
		})
	}
	group.Wait()
}

func TestSourceRecipeExampleObservationResolution(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/path-recipe.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/path-recipe.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := DecodeSourcePathDocument(context.Background(), "recipe.gooo", source, "Compose", raw)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPathOptions(context.Background(), "recipe.gooo", source, "Compose", doc, "", TypedPathOptions{
		Observation: &PathObservationOptions{Inputs: []int64{10, 0, 3}, MaxCandidates: 8, MaxRounds: 2, OracleActivity: "Expected", ReuseProbeOutputs: true, ResolveUniqueCandidate: true},
	})
	if err != nil || result.Report.BodyPaths.Resolution == nil || result.Report.BodyPaths.Resolution.Status != "RESOLVED" {
		t.Fatal("recipe observation failed", err)
	}
}
