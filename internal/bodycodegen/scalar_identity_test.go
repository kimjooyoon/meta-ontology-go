package bodycodegen

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestScalarIdentitiesPreserveAuthoredTypeNames(t *testing.T) {
	for _, kind := range []struct{ id, english, korean, native string }{
		{"integer", "Count", "정수", "int64"},
		{"boolean", "Flag", "논리", "bool"},
		{"string", "Message", "문자열", "string"},
	} {
		for _, name := range []string{kind.english, kind.korean} {
			source := fmt.Sprintf("package p\nnamespace p\nentity %s id %q\nactivity Echo(%s) -> %s computes \"return input\"\n", name, "urn:gooo:type:"+kind.id, name, name)
			result, err := Generate("alias.gooo", []byte(source), "Echo")
			if err != nil || result.Report.InputType != kind.native || result.Report.OutputType != kind.native {
				t.Fatalf("%s: %v %+v", name, err, result.Report)
			}
			if !strings.Contains(result.Source, "input "+kind.native) || !result.Report.RouteEquivalence.Equivalent {
				t.Fatal("alias lost native signature or source equivalence", name)
			}
		}
	}
}

func TestScalarIdentityPureCallsAndIntegerFill(t *testing.T) {
	source := []byte("package p\nnamespace p\nentity 정수 id \"urn:gooo:type:integer\"\nactivity Add(정수) -> 정수 computes \"return input + 1\"\nactivity Main(정수) -> 정수 computes \"return Add(input)\"\n")
	if result, err := Generate("calls.gooo", source, "Main"); err != nil || !strings.Contains(result.Source, "input int64") {
		t.Fatal("pure call lost scalar identity", err)
	}
	fixture, plan := readIRBodyFillInputs(t)
	fixture = []byte(strings.ReplaceAll(strings.ReplaceAll(string(fixture), "bodycodegen://entity/integer", "urn:gooo:type:integer"), "Integer", "정수"))
	if result, err := GenerateWithIRBodyFill(context.Background(), "fill.gooo", fixture, "ClampNegativeToZero", plan, "", ""); err != nil || result.Report.BodyFill == nil {
		t.Fatal("integer fill lost scalar identity", err)
	}
}

func TestUnknownScalarIdentityAndRecordShapesRemainExplicit(t *testing.T) {
	for _, declaration := range []string{
		`entity 정수 id "example://unknown"`,
		`entity 정수 id "urn:gooo:type:integer" fields {}`,
		`entity 정수 id "urn:gooo:type:integer" fields { field n id "example://n" type integer required one }`,
	} {
		source := "package p\nnamespace p\n" + declaration + "\nactivity Echo(정수) -> 정수 computes \"return input\"\n"
		if _, err := Generate("alias.gooo", []byte(source), "Echo"); err == nil {
			t.Fatal("unknown nominal or record acquired a scalar representation", declaration)
		}
	}
}

func TestScalarLegacySpellingsAndIdentityAuthority(t *testing.T) {
	for _, declaration := range []string{`entity Integer id "example://legacy"`, `entity Integer id "urn:gooo:type:integer"`} {
		source := "package p\nnamespace p\n" + declaration + "\nactivity Echo(Integer) -> Integer computes \"return input\"\n"
		if result, err := Generate("legacy.gooo", []byte(source), "Echo"); err != nil || result.Report.InputType != "int64" {
			t.Fatal("legacy scalar spelling changed", declaration, err)
		}
	}
	source := "package p\nnamespace p\nentity Integer id \"URN:gooo:type:boolean\"\nactivity Echo(Integer) -> Integer computes \"return !input\"\n"
	if result, err := Generate("identity.gooo", []byte(source), "Echo"); err != nil || result.Report.InputType != "bool" {
		t.Fatal("recognized identity lost to its display spelling", err)
	}
}

func TestScalarIdentityIntegerSearchAndTypedPaths(t *testing.T) {
	source := readIRBodySearchFixture(t)
	source = []byte(strings.ReplaceAll(strings.ReplaceAll(string(source), "bodycodegen://entity/integer", "urn:gooo:type:integer"), "Integer", "정수"))
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "zero", Expression: "0"}, {ID: "identity", Expression: "input"}},
		[]IRBodyFillTestCase{{Input: -2, Expected: 0}, {Input: 3, Expected: 3}}, nil, 2)
	result, err := GenerateWithIRBodySearch(context.Background(), "alias.gooo", source, "ClampNegativeToZero", plan, "", "")
	if err != nil || result.Report.BodySearch == nil || result.Report.BodySearch.TrainingPassed != 2 {
		t.Fatal("integer search lost scalar identity", err)
	}
	source = recipeSource("return input - 2")
	source = []byte(strings.ReplaceAll(strings.ReplaceAll(string(source), "sample://integer", "urn:gooo:type:integer"), "Integer", "정수"))
	doc, err := DecodeSourcePathDocument(context.Background(), "alias.gooo", source, "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	result, err = GenerateWithTypedPaths(context.Background(), "alias.gooo", source, "Assemble", doc, "")
	if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 || result.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
		t.Fatal("typed path binding lost scalar identity", err)
	}
}

func TestScalarIdentityRecordFillGrammarPreservesNativePredicates(t *testing.T) {
	for _, names := range [][3]string{{"Integer", "Boolean", "Text"}, {"정수", "논리", "문자열"}} {
		source := fmt.Sprintf(`package p
namespace p
entity %s id "urn:gooo:type:integer"
entity %s id "urn:gooo:type:boolean"
entity %s id "urn:gooo:type:string"
activity Inspect(%s, %s, %s) -> %s computes "return input0"
`, names[0], names[1], names[2], names[0], names[1], names[2], names[0])
		file, diagnostics := ParseBodyFile("predicates.gooo", []byte(source))
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		activity := file.Declarations[len(file.Declarations)-1].(*syntax.ActivityDecl)
		grammar := recordFillGrammarContext{file: file, activity: activity, cases: []assemblyspec.ValueCase{
			{Inputs: `[9007199254740993,true,"한글"]`},
		}}
		atoms, err := grammar.predicateAtoms(true)
		if err != nil || len(atoms) != 10 || !slices.Contains(atoms, "input0 >= 9007199254740993") ||
			!slices.Contains(atoms, "input1 == true") || !slices.Contains(atoms, `input2 == "한글"`) {
			t.Fatalf("native predicates changed for %v: %v %#v", names, err, atoms)
		}
		selectors, err := grammar.fieldRelationSelectors()
		if err != nil || len(selectors) != 0 {
			t.Fatal("scalar inputs were treated as records", err, selectors)
		}
	}
}
