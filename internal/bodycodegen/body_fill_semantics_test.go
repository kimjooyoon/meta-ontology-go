package bodycodegen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBodyFillEvaluatorAgreesWithCompiledEmittedGo(t *testing.T) {
	const semanticsBody = `var result = input
var true = input == 3
var false = input != 3
if true && !false { result = result + 10 } else { result = result - 10 }
if input == 3 || false { result = result } else { result = result + 100 }
var label = "go"
if label + "oo" == "gooo" { result = result } else { result = result + 100 }
if (9223372036854775807 + 1) < 0 { result = result + 100 } else { result = result + 1 }
if input == -9223372036854775808 { result = result + 100 } else { result = result }
var overflow = input + 1
if overflow < input { result = result + 1 } else { result = result + overflow }
if input > 0 { var branch = input + 1; result = result + branch } else { var branch = input - 1; result = result + branch }
return result`
	semantics, err := generateRoute("sample", "Semantics", "sample://activity/semantics", "int64", "int64", semanticsBody, preserveRoute)
	if err != nil {
		t.Fatalf("emit semantic regression body: %v", err)
	}

	const fixture = `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Precedence(Integer) -> Integer computes "if input == 3 { return 2*__GOOO_BODY_HOLE_value__ } else { return input }"
`
	plan := IRBodyFillPlan{
		Schema: bodyFillPlanSchema, Intent: "Preserve expression precedence when filling one expression hole.",
		HoleID: "value", Candidates: []IRBodyFillCandidate{
			{ID: "increment", Expression: "input + 1 // candidate tail"},
			{ID: "identity", Expression: "input"},
		},
		TestCases: []IRBodyFillTestCase{{Input: 3, Expected: 8}, {Input: 0, Expected: 0}, {Input: 1, Expected: 1}},
	}
	filled, err := GenerateWithIRBodyFill(context.Background(), "semantics.gooo", []byte(fixture), "Precedence", plan, "", "")
	if err != nil {
		t.Fatalf("fill precedence regression body: %v", err)
	}
	if filled.Report.BodyFill == nil || filled.Report.BodyFill.SelectedCandidateID != "increment" {
		t.Fatalf("precedence candidate was not selected: %#v", filled.Report.BodyFill)
	}
	const unaryFixture = `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity UnaryPrecedence(Integer) -> Integer computes "if input == 3 { return -__GOOO_BODY_HOLE_value__ } else { return input }"
`
	unaryPlan := IRBodyFillPlan{
		Schema: bodyFillPlanSchema, Intent: "Keep adjacent unary candidate tokens grouped at the expression hole.",
		HoleID: "value", Candidates: []IRBodyFillCandidate{
			{ID: "negated_input", Expression: "-input"},
			{ID: "identity", Expression: "input"},
		},
		TestCases: []IRBodyFillTestCase{{Input: 3, Expected: 3}, {Input: 0, Expected: 0}, {Input: 1, Expected: 1}},
	}
	unaryFilled, err := GenerateWithIRBodyFill(context.Background(), "unary-semantics.gooo", []byte(unaryFixture), "UnaryPrecedence", unaryPlan, "", "")
	if err != nil {
		t.Fatalf("fill unary precedence regression body: %v", err)
	}
	if unaryFilled.Report.BodyFill == nil || unaryFilled.Report.BodyFill.SelectedCandidateID != "negated_input" {
		t.Fatalf("unary precedence candidate was not selected: %#v", unaryFilled.Report.BodyFill)
	}

	const runeBody = `func RuneOverflow(input int64) int64 {
	_ = input
	var value rune = 2147483647
	value = value + 1
	if value == -2147483648 { return input + 1 } else { return input }
}`
	semanticGo := withoutPackage(t, string(semantics.source))
	filledGo := withoutPackage(t, filled.Source)
	unaryGo := withoutPackage(t, unaryFilled.Source)
	source := "package sample\n" + semanticGo + "\n" + filledGo + "\n" + unaryGo + "\n" + runeBody + "\n"
	semanticsCases := []IRBodyFillTestCase{
		{Input: 3, Expected: 22},
		{Input: -1, Expected: -12},
		{Input: -9223372036854775808, Expected: -9223372036854775717},
		{Input: 9223372036854775807, Expected: -9},
	}
	if results, passed, err := evaluateIntegerCases([]byte(source), "Semantics", semanticsCases); err != nil || passed != len(semanticsCases) {
		t.Fatalf("semantic evaluator results=%#v passed=%d err=%v", results, passed, err)
	} else {
		for index, result := range results {
			if result.Actual != semanticsCases[index].Expected {
				t.Fatalf("semantic case %d = %d, want %d", index, result.Actual, semanticsCases[index].Expected)
			}
		}
	}
	if results, passed, err := evaluateIntegerCases([]byte(source), "Precedence", plan.TestCases); err != nil || passed != len(plan.TestCases) {
		t.Fatalf("filled-body evaluator results=%#v passed=%d err=%v", results, passed, err)
	}
	if results, passed, err := evaluateIntegerCases([]byte(source), "UnaryPrecedence", unaryPlan.TestCases); err != nil || passed != len(unaryPlan.TestCases) {
		t.Fatalf("unary filled-body evaluator results=%#v passed=%d err=%v", results, passed, err)
	}
	runeCases := []IRBodyFillTestCase{{Input: 7, Expected: 8}, {Input: 9223372036854775807, Expected: -9223372036854775808}}
	if results, passed, err := evaluateIntegerCases([]byte(source), "RuneOverflow", runeCases); err != nil || passed != len(runeCases) {
		t.Fatalf("rune evaluator results=%#v passed=%d err=%v", results, passed, err)
	}

	// Run the emitted functions in one independent Go process. The cases are
	// hard-coded above, and this oracle uses the installed compiler/runtime.
	oracleSource := "package main\nimport \"fmt\"\n" + semanticGo + "\n" + filledGo + "\n" + unaryGo + "\n" + runeBody + "\nfunc main() {\n"
	want := make([]string, 0, len(semanticsCases)+len(plan.TestCases)+len(unaryPlan.TestCases)+len(runeCases))
	for _, testCase := range semanticsCases {
		oracleSource += fmt.Sprintf("fmt.Println(Semantics(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	for _, testCase := range plan.TestCases {
		oracleSource += fmt.Sprintf("fmt.Println(Precedence(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	for _, testCase := range unaryPlan.TestCases {
		oracleSource += fmt.Sprintf("fmt.Println(UnaryPrecedence(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	for _, testCase := range runeCases {
		oracleSource += fmt.Sprintf("fmt.Println(RuneOverflow(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	oracleSource += "}\n"
	if got := runBodyFillGoOracle(t, oracleSource); strings.TrimSpace(got) != strings.Join(want, "\n") {
		t.Fatalf("compiled Go oracle output = %q, want %q", got, strings.Join(want, "\n"))
	}
}

func TestBodyFillEvaluatorRejectsNonIntegralFloatWithoutPanicking(t *testing.T) {
	const source = `package sample
func FloatValue(input int64) int64 {
	var value float64 = 1.5
	value = value + 1.0
	if input == 0 { return int64(value) } else { return input }
}`
	if _, _, err := evaluateIntegerCases([]byte(source), "FloatValue", []IRBodyFillTestCase{{Input: 0, Expected: 1}}); err == nil || !strings.Contains(err.Error(), "outside the integer evaluator profile") {
		t.Fatalf("non-integral float should be rejected as outside the profile without a panic, got %v", err)
	}
}

func withoutPackage(t *testing.T, source string) string {
	t.Helper()
	_, body, ok := strings.Cut(source, "\n")
	if !ok || !strings.HasPrefix(source, "package sample\n") {
		t.Fatalf("generated source has unexpected package header: %q", source)
	}
	return body
}

func runBodyFillGoOracle(t *testing.T, source string) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.go")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", path)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("compiled Go oracle exceeded its bounded context: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("compiled Go oracle failed: %v\n%s", err, output)
	}
	return string(output)
}
