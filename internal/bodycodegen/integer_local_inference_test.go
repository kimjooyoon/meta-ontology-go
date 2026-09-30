package bodycodegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateInfersIntegerLiteralLocalsAsInt64(t *testing.T) {
	const source = `package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity ConstantLocal(Integer) -> Integer computes "let result = 5\nreturn result"
`
	result, err := Generate("constant-local.gooo", []byte(source), "ConstantLocal")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func ConstantLocal(input int64) int64", "var result int64 = 5"} {
		if !strings.Contains(result.Source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, result.Source)
		}
	}
	if !result.Report.TypecheckPassed || result.Report.RouteEquivalence.Decision != "PASS" ||
		result.Report.SourceSemanticUnits != result.Report.LoweredSemanticUnits || result.Report.CompletenessPercent != 100 {
		t.Fatalf("integer local normalization changed source accounting or failed equivalence: %#v", result.Report)
	}
}

func TestIntegerLocalInferenceContinuesAfterTypeErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
		want []string
	}{
		{
			name: "dependent sequential locals",
			body: "var a = 5\nvar b = input + a\nvar c = 7\nreturn b + c",
			want: []string{"var a int64 = 5", "var b = input + a", "var c int64 = 7"},
		},
		{
			name: "nested branch local after bad condition",
			body: "var a = 5\nif input > a { var later = 7; return later } else { return 0 }",
			want: []string{"var a int64 = 5", "if input > a", "var later int64 = 7"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			generated, err := generateRoute("sample", "ContinueInference", "sample://activity/continue-inference", "int64", "int64", test.body, preserveRoute)
			if err != nil {
				t.Fatalf("generate after initial type errors: %v", err)
			}
			for _, want := range test.want {
				if !strings.Contains(string(generated.source), want) {
					t.Fatalf("generated source missing %q:\n%s", want, generated.source)
				}
			}
		})
	}
}

func TestInferredIntegerLocalEvaluatorMatchesCompiledGo(t *testing.T) {
	const body = `var minValue = -9223372036854775808
var maxValue = 9223372036854775807
var delta = 20 + 22
var result = minValue
result = result + 1
if input == maxValue { result = maxValue } else { result = result + delta }
return result`
	generated, err := generateRoute("sample", "InferredConstants", "sample://activity/inferred-constants", "int64", "int64", body, preserveRoute)
	if err != nil {
		t.Fatalf("generate inferred integer locals: %v", err)
	}
	cases := []IRBodyFillTestCase{
		{Input: -9223372036854775808, Expected: -9223372036854775765},
		{Input: 0, Expected: -9223372036854775765},
		{Input: 9223372036854775807, Expected: 9223372036854775807},
	}
	results, passed, err := evaluateIntegerCases(generated.source, "InferredConstants", cases)
	if err != nil || passed != len(cases) {
		t.Fatalf("evaluator results=%#v passed=%d err=%v", results, passed, err)
	}
	for index, result := range results {
		if !result.Passed || result.Actual != cases[index].Expected {
			t.Fatalf("case %d evaluator result=%#v want=%d", index, result, cases[index].Expected)
		}
	}

	const reassignmentBody = `var result = 5
result = input
result = result + 1
return result`
	reassigned, err := generateRoute("sample", "Reassigned", "sample://activity/reassigned", "int64", "int64", reassignmentBody, preserveRoute)
	if err != nil {
		t.Fatalf("generate reassigned integer local: %v", err)
	}
	reassignmentCases := []IRBodyFillTestCase{
		{Input: -9223372036854775808, Expected: -9223372036854775807},
		{Input: 0, Expected: 1},
		{Input: 9223372036854775807, Expected: -9223372036854775808},
	}
	results, passed, err = evaluateIntegerCases(reassigned.source, "Reassigned", reassignmentCases)
	if err != nil || passed != len(reassignmentCases) {
		t.Fatalf("reassignment evaluator results=%#v passed=%d err=%v", results, passed, err)
	}

	compiled := "package main\nimport \"fmt\"\n" + withoutPackage(t, string(generated.source)) + "\n" +
		withoutPackage(t, string(reassigned.source)) + "\nfunc main() {\n"
	want := make([]string, 0, len(cases)+len(reassignmentCases))
	for _, testCase := range cases {
		compiled += fmt.Sprintf("fmt.Println(InferredConstants(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	for _, testCase := range reassignmentCases {
		compiled += fmt.Sprintf("fmt.Println(Reassigned(%d))\n", testCase.Input)
		want = append(want, fmt.Sprint(testCase.Expected))
	}
	compiled += "}\n"
	if got := strings.TrimSpace(runBodyFillGoOracle(t, compiled)); got != strings.Join(want, "\n") {
		t.Fatalf("compiled Go output=%q, want=%q", got, strings.Join(want, "\n"))
	}
}

func TestIntegerLocalInferenceRejectsOutOfRangeCompileTimeConstants(t *testing.T) {
	for _, body := range []string{
		"var value = 9223372036854775808\nreturn 0",
		"var value = 9223372036854775807 + 1\nreturn 0",
	} {
		if _, err := generateRoute("sample", "OutOfRange", "sample://activity/out-of-range", "int64", "int64", body, preserveRoute); err == nil {
			t.Fatalf("compile-time out-of-range integer was accepted: %s", body)
		}
	}
}

func TestIntegerLocalInferencePreservesBoolAndTextDefaults(t *testing.T) {
	for _, test := range []struct {
		name       string
		inputType  string
		outputType string
		body       string
		want       string
	}{
		{name: "boolean", inputType: "bool", outputType: "bool", body: "var value = true\nreturn value", want: "var value = true"},
		{name: "text", inputType: "string", outputType: "string", body: "var value = \"text\"\nreturn value", want: `var value = "text"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			generated, err := generateRoute("sample", "Local", "sample://activity/local", test.inputType, test.outputType, test.body, preserveRoute)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(generated.source), test.want) || strings.Contains(string(generated.source), "var value bool") || strings.Contains(string(generated.source), "var value string") {
				t.Fatalf("non-integer local inference changed unexpectedly:\n%s", generated.source)
			}
		})
	}
}
