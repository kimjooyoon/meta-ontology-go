package linecaps

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestAnalyzeDiscoversSortedFilesAndSkipsExcludedDirectories(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, filepath.Join("z", "last.go"), "package z\n")
	writeGoFile(t, root, filepath.Join("a", "first.go"), "package a\n")
	writeGoFile(t, root, filepath.Join("vendor", "ignored.go"), "package ignored\n"+strings.Repeat("\n", 10))
	writeGoFile(t, root, filepath.Join(".git", "ignored.go"), "package ignored\n"+strings.Repeat("\n", 10))
	files, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a/first.go", "z/last.go"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected discovery order: got %v want %v", files, want)
	}
	report, err := Analyze(root, nil, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("discovered files should pass: %s", report.Text())
	}
}

func TestAnalyzeIsPermutationInvariantAndDoesNotMutatePaths(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "a.go", "package p\n\nfunc A() {\n\t_ = 1\n}\n")
	writeGoFile(t, root, "b.go", "package p\n\nfunc B() {\n\t_ = 2\n}\n")
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	absoluteA, err := filepath.Abs(filepath.Join(root, "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	firstPaths := []string{"b.go", absoluteA, "./a.go", "a.go"}
	secondPaths := []string{"a.go", "b.go"}
	firstBefore := append([]string(nil), firstPaths...)
	secondBefore := append([]string(nil), secondPaths...)
	first, err := Analyze(root, firstPaths, limits)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Analyze(root, secondPaths, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || first.Text() != second.Text() {
		t.Fatalf("path permutation changed result:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if !reflect.DeepEqual(firstPaths, firstBefore) || !reflect.DeepEqual(secondPaths, secondBefore) {
		t.Fatal("Analyze mutated its path inputs")
	}
}

func TestReportTextJSONAndErrorAreDeterministic(t *testing.T) {
	findings := []Finding{
		{Path: "z.go", Rule: RuleFileLines, Actual: 301, Limit: 300},
		{Path: "a.go", Rule: RuleFunctionLines, Name: "F", StartLine: 2, EndLine: 77, Actual: 76, Limit: 75},
	}
	first := Report{Findings: findings}
	second := Report{Findings: []Finding{findings[1], findings[0]}}
	wantText := "linecaps: violations=2\na.go:2-77: function-lines F: got 76, limit 75\nz.go: file-lines: got 301, limit 300\n"
	if first.Text() != wantText || second.Text() != wantText {
		t.Fatalf("unexpected text output:\nfirst=%ssecond=%s", first.Text(), second.Text())
	}
	firstJSON, err := first.JSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := second.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("JSON output changed with finding order:\nfirst=%ssecond=%s", firstJSON, secondJSON)
	}
	var decoded Report
	if err := json.Unmarshal(firstJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Findings) != 2 || decoded.Findings[0].Path != "a.go" {
		t.Fatalf("JSON output was not ordered: %s", firstJSON)
	}
	if first.Err() == nil || first.Err().Error() != strings.TrimSuffix(wantText, "\n") {
		t.Fatalf("error output did not match text format: %v", first.Err())
	}
}

func TestParseAndReadFailuresRemainVisible(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "broken.go", "package p\nfunc broken(\n")
	report, err := Analyze(root, []string{"broken.go", "missing.go"}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(report.Findings, RuleParseFile) || !hasRule(report.Findings, RuleReadFile) {
		t.Fatalf("verification failures were hidden: %#v", report.Findings)
	}
}

func TestInvalidInputsFailClosed(t *testing.T) {
	root := t.TempDir()
	for name, run := range map[string]func() error{
		"invalid limits": func() error {
			_, err := Analyze(root, nil, Limits{})
			return err
		},
		"empty root": func() error {
			_, err := Analyze("", nil, DefaultLimits())
			return err
		},
		"relative escape": func() error {
			_, err := Analyze(root, []string{"../outside.go"}, DefaultLimits())
			return err
		},
		"absolute escape": func() error {
			_, err := Analyze(root, []string{filepath.Join(root, "..", "outside.go")}, DefaultLimits())
			return err
		},
		"empty path": func() error {
			_, err := Analyze(root, []string{""}, DefaultLimits())
			return err
		},
		"empty source path": func() error {
			_, err := AnalyzeSource("", []byte("package p\n"), DefaultLimits())
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil {
				t.Fatal("malformed input was accepted")
			}
		})
	}
}

func TestAnalyzeSourceDoesNotMutateSource(t *testing.T) {
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	source := []byte("package p\n\nfunc TooLong() {\n\t_ = 1\n}\n")
	want := append([]byte(nil), source...)
	findings, err := AnalyzeSource("fixture.go", source, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, RuleFunctionLines) {
		t.Fatalf("expected function finding, got %#v", findings)
	}
	if !bytes.Equal(source, want) {
		t.Fatal("AnalyzeSource mutated its source input")
	}
}

func TestRefactorCandidateFindingForSimpleReturns(t *testing.T) {
	source := `
package p

func ReturnDirect(v int) int {
	return v
}

func ReturnWrapped(v int) int {
	return len(v)
}

func ReturnVariable(v int) int {
	value := v
	return value
}

func NotRefactorable(v int) int {
	if v > 0 {
		return v
	}
	return 0
}
`
	findings, err := AnalyzeSource("fixture.go", []byte(source), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	refactorRules := filterRules(findings, []Rule{RuleRefactorReturn, RuleRefactorAssign})
	if len(refactorRules) != 3 {
		t.Fatalf("unexpected refactor findings: %#v", findings)
	}
	for _, finding := range refactorRules {
		if finding.Rule == RuleRefactorReturn && !strings.Contains(finding.Detail, "single return") {
			t.Fatalf("unexpected refactor-return detail: %#v", finding)
		}
	}
	if !hasFinding(findings, RuleRefactorAssign) {
		t.Fatalf("expected assignment refactor finding: %#v", findings)
	}
	if !hasFinding(findings, RuleRefactorReturn) {
		t.Fatalf("expected return refactor findings: %#v", findings)
	}
}

func TestRefactorCandidateFindingSkipsLargeFunctions(t *testing.T) {
	source := `
package p

func Long(v int) int {
	for i := 0; i < 3; i++ {
		v += i
	}
	return v
}
`
	findings, err := AnalyzeSource("fixture.go", []byte(source), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(findings, RuleRefactorReturn) || hasFinding(findings, RuleRefactorAssign) {
		t.Fatalf("non-trivial function should not be refactor candidate: %#v", findings)
	}
}

func sourceWithFunctionLines(lines int) string {
	return "package p\n\nfunc F() {\n" + strings.Repeat("\t_ = 1\n", lines-4) + "}\n"
}

func padSourceToLines(source string, lines int) string {
	current := strings.Count(source, "\n")
	if !strings.HasSuffix(source, "\n") {
		current++
	}
	return source + strings.Repeat("\n", lines-current)
}

func writeGoFile(t *testing.T, root, path, source string) {
	t.Helper()
	fullPath := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasRule(findings []Finding, rule Rule) bool {
	return countRule(findings, rule) > 0
}

func countRule(findings []Finding, rule Rule) int {
	count := 0
	for _, finding := range findings {
		if finding.Rule == rule {
			count++
		}
	}
	return count
}

func hasFinding(findings []Finding, rule Rule) bool {
	for _, finding := range findings {
		if finding.Rule == rule {
			return true
		}
	}
	return false
}

func filterRules(findings []Finding, rules []Rule) []Finding {
	result := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		for _, rule := range rules {
			if finding.Rule == rule {
				result = append(result, finding)
				break
			}
		}
	}
	return result
}
