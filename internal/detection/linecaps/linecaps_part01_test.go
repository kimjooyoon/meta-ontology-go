package linecaps

import (
	"testing"
)

func TestDefaultLimitsAndExactBoundariesPass(t *testing.T) {
	limits := DefaultLimits()
	source := padSourceToLines(sourceWithFunctionLines(limits.MaxFunctionLines), limits.MaxFileLines)
	findings, err := AnalyzeSource("fixture.go", []byte(source), limits)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("exact caps should pass: %#v", findings)
	}
}
func TestFileCapRejectsOneLineOverAndHandlesEOF(t *testing.T) {
	limits := Limits{MaxFileLines: 3, MaxFunctionLines: 75}
	for name, source := range map[string]string{
		"trailing newline":    "package p\n\n\n\n",
		"no trailing newline": "package p\n\n\nfunc F() {}",
	} {
		t.Run(name, func(t *testing.T) {
			findings, err := AnalyzeSource("fixture.go", []byte(source), limits)
			if err != nil {
				t.Fatal(err)
			}
			if !hasRule(findings, RuleFileLines) {
				t.Fatalf("expected file cap finding, got %#v", findings)
			}
		})
	}
}
func TestFunctionCapRejectsOneLineOverAndIncludesLiterals(t *testing.T) {
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	source := "package p\n\nfunc TooLong() {\n\t_ = 1\n}\n\nvar _ = func() {\n\t_ = 2\n}\n"
	findings, err := AnalyzeSource("fixture.go", []byte(source), limits)
	if err != nil {
		t.Fatal(err)
	}
	if countRule(findings, RuleFunctionLines) != 2 {
		t.Fatalf("expected declaration and literal findings, got %#v", findings)
	}
	if findings[0].Actual != limits.MaxFunctionLines+1 {
		t.Fatalf("expected one-line-over finding, got %#v", findings[0])
	}
}
func TestMethodsAndNestedFunctionsHaveStableNamesAndRanges(t *testing.T) {
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	source := "package p\n\ntype T struct{}\n\nfunc (T) Method() {\n\tfunc() {\n\t\t_ = 1\n\t}()\n}\n"
	findings, err := AnalyzeSource("fixture.go", []byte(source), limits)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected method and literal findings, got %#v", findings)
	}
	if findings[0].Name != "method Method" || findings[0].StartLine != 5 || findings[0].EndLine != 9 {
		t.Fatalf("unexpected method finding: %#v", findings[0])
	}
	if findings[1].Name != "function literal" || findings[1].StartLine != 6 || findings[1].EndLine != 8 {
		t.Fatalf("unexpected literal finding: %#v", findings[1])
	}
}
