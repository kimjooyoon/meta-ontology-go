package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
)

func TestFormatAndFixRemainReadOnly(t *testing.T) {
	reader := fixtureReader{source: "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\n"}
	var stdout, stderr bytes.Buffer
	if code := runFormat([]string{"--json", "input.gooo"}, reader, &stdout, &stderr); code != exitOK {
		t.Fatalf("format = %d stderr=%q", code, stderr.String())
	}
	report := formatCommandReport{}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.Changed || report.DirectWrites != 0 || !strings.Contains(report.Source, "\n\nentity") {
		t.Fatalf("report = %#v", report)
	}
	stdout.Reset()
	if code := runFix([]string{"--json", "input.gooo"}, reader, &stdout, &stderr); code != exitOK {
		t.Fatalf("fix = %d stderr=%q", code, stderr.String())
	}
	plan := formatfix.Plan{}
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	if err := formatfix.Validate(plan); err != nil {
		t.Fatal(err)
	}
	if plan.Decision != formatfix.DecisionChangePlanned || plan.DirectWrites != 0 {
		t.Fatalf("plan = %#v", plan)
	}
}

func TestFixRejectsWriteAuthority(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runFix([]string{"--write", "input.gooo"}, fixtureReader{}, &stdout, &stderr); code != exitUsage ||
		stderr.String() != fixUsage+"\n" {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}
