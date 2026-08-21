package linecaps

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

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
