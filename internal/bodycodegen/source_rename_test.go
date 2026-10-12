package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const sourceRenameFixture = `package pricing
namespace pricing
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
// Calculate is the original term in this note.
activity Calculate(Integer) -> Integer computes "return input - 11"
activity Check(Integer,Integer) -> Boolean computes "return input1 == Calculate(input0)"
activity Select(Integer) -> Integer computes "return Calculate(input)" assembling {
 choice "value" field_value at "0" alternative "Calculate(input) + 1" intent "Calculate the fee."
 criteria {
  predicate "fee" activity "Check"
  input "[4]"
  holdout_input "[23]"
 }
 attempts "2"
}
`

func TestSourceRenameKeepsIdentityCallsAndObligations(t *testing.T) {
	result, err := RenameActivitySource(context.Background(), "rename.gooo", []byte(sourceRenameFixture), "Calculate", "요금")
	if err != nil {
		t.Fatal(err)
	}
	if result.ActivityID != "pricing://activity/calculate" || !result.IdentityPinned || result.CallSites != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, want := range []string{`activity 요금(Integer) -> Integer id "pricing://activity/calculate"`,
		`return input1 == 요금(input0)`, `alternative "요금(input) + 1"`,
		`// Calculate is the original term`, `intent "Calculate the fee."`, `holdout_input "[23]"`} {
		if !strings.Contains(result.Source, want) {
			t.Fatal("missing", want, result.Source)
		}
	}
	if result.InputSHA256 == result.OutputSHA256 || result.InputSHA256 != digest([]byte(sourceRenameFixture)) {
		t.Fatal("source identity did not change")
	}
	previous := 0
	var rebuilt strings.Builder
	for _, edit := range result.Edits {
		if edit.Start < previous || sourceRenameFixture[edit.Start:edit.End] != edit.Before {
			t.Fatal("invalid edit span", edit)
		}
		rebuilt.WriteString(sourceRenameFixture[previous:edit.Start])
		rebuilt.WriteString(edit.After)
		previous = edit.End
	}
	rebuilt.WriteString(sourceRenameFixture[previous:])
	if rebuilt.String() != result.Source {
		t.Fatal("receipt did not reconstruct source")
	}
	criteria, err := RenameActivitySource(context.Background(), "rename.gooo", []byte(result.Source), "Check", "조건")
	if err != nil || !strings.Contains(criteria.Source, `predicate "fee" activity "조건"`) {
		t.Fatal("criterion reference was not renamed", err, criteria.Source)
	}
	file, diags := ParseBodyFile("renamed.gooo", []byte(criteria.Source))
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	selectActivity := file.Declarations[4].(*syntax.ActivityDecl)
	if selectActivity.Assembly.Spec.Criteria.Inputs[0] != "[4]" || selectActivity.Assembly.Spec.Criteria.HoldoutInputs[0] != "[23]" {
		t.Fatal("criterion inputs changed")
	}
}

func TestSourceRenamePreservesRawStringsAndLexicalBindings(t *testing.T) {
	for _, tc := range []struct {
		body, want string
		count      int
	}{
		{"return (Calculate)(input)", "return (요금)(input)", 1},
		{"let Calculate = input; return Calculate", "let Calculate = input; return Calculate", 0},
		{"let text = \"Calculate(input)\"; _ = text; return Calculate(input)", "let text = \"Calculate(input)\"; _ = text; return 요금(input)", 1},
		{"// Calculate(input)\nreturn input.Calculate + Calculate(input.Fee)", "// Calculate(input)\nreturn input.Calculate + 요금(input.Fee)", 1},
		{"if input > 0 { let Calculate = input; _ = Calculate }; return Calculate(input)", "if input > 0 { let Calculate = input; _ = Calculate }; return 요금(input)", 1},
	} {
		got, n, err := renameBodyCalls(tc.body, 1, "Calculate", "요금")
		if err != nil || n != tc.count || got != tc.want {
			t.Fatal(tc.body, got, n, err)
		}
	}
	source := strings.Replace(sourceRenameFixture, `computes "return input1 == Calculate(input0)"`, "computes `\r\nreturn input1 == Calculate(input0)\r\n`", 1)
	result, err := RenameActivitySource(context.Background(), "raw.gooo", []byte(source), "Calculate", "요금")
	if err != nil || !strings.Contains(result.Source, "computes `\r\nreturn input1 == 요금(input0)\r\n`") {
		t.Fatal(err, result.Source)
	}
}

func TestSourceRenameRejectsCaptureAndAmbiguousUses(t *testing.T) {
	for _, tc := range []struct{ body, to string }{
		{"return Calculate(input)", "input"},
		{"let 요금 = input; return Calculate(input)", "요금"},
		{"return Calculate(input) + len(input)", "len"},
		{"let f = Calculate; return f(input)", "요금"},
		{"return (", "요금"},
	} {
		if _, _, err := renameBodyCalls(tc.body, 1, "Calculate", tc.to); err == nil {
			t.Fatal("accepted", tc)
		}
	}
	for _, name := range []string{"Integer", "Check", "activity", "for", "_", "x-y", "Calculate", ""} {
		result, err := RenameActivitySource(context.Background(), "bad.gooo", []byte(sourceRenameFixture), "Calculate", name)
		if err == nil || result.Source != "" {
			t.Fatal("partial or accepted result", name, err)
		}
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := RenameActivitySource(canceled, "cancel.gooo", []byte(sourceRenameFixture), "Calculate", "요금"); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestSourceRenameKeepsExplicitIDBindingsAndOracle(t *testing.T) {
	source := `package flow
namespace flow
entity Integer id "urn:gooo:type:integer"
activity Calculate(Integer) -> Integer id "urn:gooo:activity:fee" computes "return input - 11"
activity Receive(Integer) -> Integer computes "return input - 11" assembling {
 choice "order" operand_order at "0" intent "Calculate is a note."
 case "11" -> "0"
 observe { input "12" oracle "Calculate" max_candidates "2" max_rounds "1" }
 attempts "2"
}

bind Calculate.result -> Receive.input
`
	result, err := RenameActivitySource(context.Background(), "flow.gooo", []byte(source), "Calculate", "요금")
	if err != nil {
		t.Fatal(err)
	}
	if result.IdentityPinned || result.ActivityID != "urn:gooo:activity:fee" ||
		!strings.Contains(result.Source, `bind 요금.result -> Receive.input`) ||
		!strings.Contains(result.Source, `oracle "요금"`) {
		t.Fatal(result)
	}
	if strings.Count(result.Source, `id "urn:gooo:activity:fee"`) != 1 {
		t.Fatal("identity duplicated")
	}
}

func TestSourceRenameCheckpointAndScopeLimits(t *testing.T) {
	source := strings.Replace(sourceRenameFixture, ` attempts "2"`, ` baseline "return Calculate(input)"
 picked "value" -> "value_first"
 attempts "2"`, 1)
	result, err := RenameActivitySource(nil, "checkpoint.gooo", []byte(source), "Calculate", "요금")
	if err != nil || result.CallSites != 4 || !strings.Contains(result.Source, `baseline "return 요금(input)"`) ||
		!strings.Contains(result.Source, `picked "value" -> "value_first"`) {
		t.Fatal(err, result)
	}
	for _, name := range []string{"bool", "string", "int64", "nil", "true", "len", "init"} {
		if _, err := RenameActivitySource(nil, "reserved.gooo", []byte(sourceRenameFixture), "Calculate", name); err == nil {
			t.Fatal("accepted reserved projection name", name)
		}
	}
	if _, err := RenameActivitySource(nil, "large.gooo", make([]byte, 128<<10+1), "Calculate", "요금"); err == nil {
		t.Fatal("ignored source limit")
	}
	fragment := strings.Replace(sourceRenameFixture, `return Calculate(input)`, `let 요금 = input; return Calculate(input)`, 1)
	if _, err := RenameActivitySource(nil, "capture.gooo", []byte(fragment), "Calculate", "요금"); err == nil {
		t.Fatal("accepted enclosing capture")
	}
	search := `package p
namespace p
entity Integer id "urn:gooo:type:integer"
activity Calculate(Integer) -> Integer computes "return input" assembling {
 search hole "result" grammar "integer-offset-constant/v1" intent "return input" max_candidates "2"
 case "1" -> "1"
 attempts "2"
}
`
	file, diag := ParseBodyFile("search.gooo", []byte(search))
	if diag.HasErrors() || file == nil {
		t.Fatal("invalid scope fixture", diag)
	}
	if result, err := RenameActivitySource(nil, "search.gooo", []byte(search), "Calculate", "요금"); err == nil || result.Source != "" || !strings.Contains(err.Error(), "candidate-scope") {
		t.Fatal("search silently rewritten", err, result)
	}
}
