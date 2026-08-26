package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContractClosesLanguageSelfImprovementLoop(t *testing.T) {
	path := filepath.Join(
		"..", "..", "examples", "self-improvement", "main.gooo",
	)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	report := buildReport(path, source, strings.Repeat("a", 40))
	if report.Status != "PASS" {
		t.Fatalf("status = %s, errors = %v", report.Status, report.Errors)
	}
	if report.EntityCount != 15 || report.ActivityCount != 8 {
		t.Fatalf("model = %d entities/%d activities",
			report.EntityCount, report.ActivityCount)
	}
	if len(report.ExecutorCoverage) != 3 {
		t.Fatalf("executor coverage = %d, want 3", len(report.ExecutorCoverage))
	}
	for _, indicator := range report.Indicators {
		if indicator.Verdict != "PASS" {
			t.Fatalf("indicator %s = %s", indicator.ID, indicator.Verdict)
		}
	}
}

func TestContractRejectsObservationConsumption(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-improvement", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	open := strings.Replace(string(source), "activity Improve(Evidence) -> SourceTree", "activity Improve(Evidence, ReadOnlyImprovementInput) -> SourceTree", 1)
	report := buildReport(path, []byte(open), strings.Repeat("c", 40))
	if report.Status != "FAIL" {
		t.Fatalf("status = %s, want FAIL", report.Status)
	}
}

func TestContractRejectsAnOpenImprovementLoop(t *testing.T) {
	path := filepath.Join(
		"..", "..", "examples", "self-improvement", "main.gooo",
	)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	open := strings.Replace(
		string(source),
		"activity Improve(Evidence) -> SourceTree",
		"activity Improve(Evidence) -> Evidence",
		1,
	)
	report := buildReport(path, []byte(open), strings.Repeat("b", 40))
	if report.Status != "FAIL" {
		t.Fatalf("status = %s, want FAIL", report.Status)
	}
	for _, indicator := range report.Indicators {
		if indicator.ID == "coherence.closed-self-improvement-loop" &&
			indicator.Verdict == "FAIL" {
			return
		}
	}
	t.Fatal("closed-loop indicator did not fail")
}
