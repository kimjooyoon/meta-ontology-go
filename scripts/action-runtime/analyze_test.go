package main

import (
	"strings"
	"testing"
)

func TestBuildReportConnectsPolicyAndIndicators(t *testing.T) {
	source := []byte(`steps:
  - uses: actions/checkout@v5
  - uses: actions/setup-go@v6
  - uses: actions/upload-artifact@v6
  - uses: actions/download-artifact@v7
    with:
      run-id: 42
      name: evidence
  - uses: actions/github-script@v8
`)
	report := buildReport("ci.yml", source, strings.Repeat("a", 40))
	if report.Status != "PASS" {
		t.Fatalf("status = %s, observations = %#v", report.Status, report.Observations)
	}
	if report.ActionsTotal != 5 || report.ActionsCompliant != 5 {
		t.Fatalf("actions = %d/%d", report.ActionsCompliant, report.ActionsTotal)
	}
	if len(report.Indicators) != 5 {
		t.Fatalf("indicators = %d, want 5", len(report.Indicators))
	}
}

func TestBuildReportRejectsRuntimeAndInputDrift(t *testing.T) {
	source := []byte(`steps:
  - uses: actions/checkout@v4
  - uses: actions/download-artifact@v7
    with:
      if-no-files-found: error
`)
	report := buildReport("ci.yml", source, strings.Repeat("b", 40))
	if report.Status != "FAIL" {
		t.Fatalf("status = %s, want FAIL", report.Status)
	}
	if report.ActionsCompliant != 1 || report.InvalidInputsTotal != 1 {
		t.Fatalf("runtime = %d, invalid = %d",
			report.ActionsCompliant, report.InvalidInputsTotal)
	}
}
