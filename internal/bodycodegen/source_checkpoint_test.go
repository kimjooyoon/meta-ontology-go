package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func checkpointFixture(t *testing.T, body string, choices []assemblyspec.Choice) []byte {
	t.Helper()
	spec := assemblyspec.Spec{Choices: choices, Cases: []assemblyspec.Case{{Input: 0, Expected: 2}}, MaxAttempts: 64}
	formatted, err := syntax.FormatAssembly(&syntax.AssemblyDecl{Spec: spec})
	if err != nil {
		t.Fatal(err)
	}
	return []byte(strings.TrimSuffix(string(recipeSource(body)), "\n") + " " + formatted + "\n")
}

func assertCheckpointPreservesDocument(t *testing.T, source []byte, labels map[string]string) []byte {
	t.Helper()
	ctx := context.Background()
	document, err := DecodeSourcePathDocument(ctx, "source.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := document.Prepare()
	selected, err := prepared.Compile(labels)
	if err != nil {
		t.Fatal(err)
	}
	file, _ := syntax.Parse(string(source))
	activity := file.Declarations[1].(*syntax.ActivityDecl)
	checkpoint, err := selectedTypedPathSource(source, activity, selected.GoooBody(), labels, true)
	if err != nil {
		t.Fatal(err)
	}
	again, err := DecodeSourcePathDocument(ctx, "checkpoint.gooo", checkpoint, "Assemble", nil)
	if err != nil || !reflect.DeepEqual(document, again) {
		t.Fatal("checkpoint changed the complete alternative palette", err)
	}
	return checkpoint
}

func TestCheckpointPreservesAllFiveStructuralKinds(t *testing.T) {
	source := checkpointFixture(t, "let first = input + 2; let second = input * 2; let chosen = first; "+
		"if input < 0 { chosen = chosen + 3 } else { chosen = chosen - 3 }; return chosen + (first + second)",
		[]assemblyspec.Choice{
			{ID: "operands", Kind: "operand_order", Intent: "choose operands"},
			{ID: "reference", Kind: "local_reference", Intent: "choose local", Alternative: "second"},
			{ID: "assignment", Kind: "assignment_target", Intent: "choose assignment", Alternative: "first"},
			{ID: "branch", Kind: "branch_layout", Intent: "choose branch"},
			{ID: "order", Kind: "root_order", Intent: "choose order"},
		})
	assertCheckpointPreservesDocument(t, source, map[string]string{
		"operands": "layout_reverse", "reference": "reference_second", "assignment": "assign_second",
		"branch": "layout_reverse", "order": "schedule_reverse",
	})
}

func TestCheckpointRetainsInterdependentAlternatives(t *testing.T) {
	source := checkpointFixture(t, "let x = input; let y = input + 1; let z = x; return z + (y * 0) + (x * 0)", []assemblyspec.Choice{
		{ID: "reference", Kind: "local_reference", Intent: "choose local", Alternative: "y"},
		{ID: "order", Kind: "root_order", Occurrence: 1, Intent: "choose order"},
	})
	assertCheckpointPreservesDocument(t, source,
		map[string]string{"reference": "reference_first", "order": "schedule_reverse"})
	document, _ := DecodeSourcePathDocument(context.Background(), "source.gooo", source, "Assemble", nil)
	prepared, _ := document.Prepare()
	if _, err := prepared.Compile(map[string]string{"reference": "reference_second", "order": "schedule_reverse"}); err == nil {
		t.Fatal("interdependent local incorrectly available before its declaration")
	}
}

func TestCheckpointPreservesPartialFiniteScore(t *testing.T) {
	source := checkpointFixture(t, "return input - 2", []assemblyspec.Choice{
		{ID: "offset", Kind: "operand_order", Intent: "retain incomplete finite behavior"},
	})
	source = []byte(strings.Replace(string(source), `case "0" -> "2"`, `case "0" -> "999"`, 1))
	ctx := context.Background()
	document, err := DecodeSourcePathDocument(ctx, "source.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "source.gooo", source, "Assemble", document, "")
	if err != nil {
		t.Fatal(err)
	}
	realized, err := RealizeSourceAssembly(ctx, "source.gooo", source, result)
	if err != nil || realized.FinitePassed != 0 || realized.FiniteTotal != 1 {
		t.Fatal("partial functional completeness was discarded", err)
	}
}
