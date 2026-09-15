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
	if report.EntityCount != 20 || report.ActivityCount != 9 {
		t.Fatalf("model = %d entities/%d activities",
			report.EntityCount, report.ActivityCount)
	}
	if len(report.ExecutorCoverage) != 5 {
		t.Fatalf("executor coverage = %d, want 5", len(report.ExecutorCoverage))
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

func TestRegistrationExecutorRequiresItsExactExecuteInput(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "self-improvement", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, broken := range []string{
		strings.Replace(string(source), ", SyntaxRegistrationWorker) -> ExecutionManifest", ") -> ExecutionManifest", 1),
		strings.Replace(string(source), "executor://scripts/meta-execution:registration-worker",
			"executor://unbound-registration-worker", 1),
	} {
		report := buildReport(path, []byte(broken), strings.Repeat("d", 40))
		if report.Status != "FAIL" || len(report.ExecutorCoverage) != 5 {
			t.Fatalf("missing native input was hidden: %+v", report)
		}
		found := false
		for _, coverage := range report.ExecutorCoverage {
			if coverage.Operation == "register-syntax-capability" {
				found = true
				if coverage.Covered {
					t.Fatal("unbound native registration executor was covered")
				}
			}
		}
		if !found {
			t.Fatal("native registration obligation disappeared from denominator")
		}
	}
}
