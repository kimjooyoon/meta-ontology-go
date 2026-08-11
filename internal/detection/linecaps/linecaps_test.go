package linecaps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultLimitsAndExactBoundariesPass(t *testing.T) {
	limits := DefaultLimits()
	source := sourceWithFunctionLines(limits.MaxFunctionLines)
	source = padSourceToLines(source, limits.MaxFileLines)
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

func TestAnalyzeDiscoversSortedFilesAndSkipsVendor(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, filepath.Join("z", "last.go"), "package z\n")
	writeGoFile(t, root, filepath.Join("a", "first.go"), "package a\n")
	writeGoFile(t, root, filepath.Join("vendor", "ignored.go"), "package ignored\n"+strings.Repeat("\n", 10))
	files, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a/first.go", "z/last.go"}
	if strings.Join(files, ",") != strings.Join(want, ",") {
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

func TestReportTextJSONAndErrorAreDeterministic(t *testing.T) {
	report := Report{Findings: []Finding{
		{Path: "z.go", Rule: RuleFileLines, Actual: 301, Limit: 300},
		{Path: "a.go", Rule: RuleFunctionLines, Name: "F", StartLine: 2, EndLine: 77, Actual: 76, Limit: 75},
	}}
	wantText := "linecaps: violations=2\na.go:2-77: function-lines F: got 76, limit 75\nz.go: file-lines: got 301, limit 300\n"
	if got := report.Text(); got != wantText {
		t.Fatalf("unexpected text output:\n%s", got)
	}
	data, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Findings) != 2 || decoded.Findings[0].Path != "a.go" {
		t.Fatalf("JSON output was not ordered: %s", data)
	}
	if report.Err() == nil || report.Err().Error() != strings.TrimSuffix(wantText, "\n") {
		t.Fatalf("error output did not match text format: %v", report.Err())
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

func TestInvalidLimitsAndEscapingPathsFail(t *testing.T) {
	root := t.TempDir()
	if _, err := Analyze(root, nil, Limits{}); err == nil {
		t.Fatal("invalid limits were accepted")
	}
	if _, err := Analyze(root, []string{"../outside.go"}, DefaultLimits()); err == nil {
		t.Fatal("path escaping root was accepted")
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
