package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestRunChecksAndAppliesExactEvidence(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "fixture.go")
	source := "package fixture\n\nfunc value() int {\n\tresult := 1\n\treturn result\n}\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	subject := (sourcepolicy.SourceSubject{Path: "fixture.go", Line: 3, Name: "value"}).String()
	report := linecaps.LineMetricsReport{CommitSHA: "exact-head", Meta: sourcepolicy.Report{Indicators: []sourcepolicy.Indicator{{
		MetricID: sourcepolicy.DimensionRefactorAssign, Family: sourcepolicy.FamilyRefactor, Subject: subject,
		Value: 2, Relation: sourcepolicy.RelationEqual, Satisfied: false, Proof: sourcepolicy.ProofRegression,
		Producer: "linecaps.Analyze", Consumer: "refactor-planner", Operation: sourcepolicy.OperationCollapseAssign,
	}}}}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	metrics := filepath.Join(root, "metrics.json")
	if err := os.WriteFile(metrics, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(options{root: root, metrics: metrics, sha: "exact-head", check: true}); err != nil {
		t.Fatal(err)
	}
	if err := run(options{root: root, metrics: metrics, sha: "exact-head", subject: subject}); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(updated), "result :=") || !strings.Contains(string(updated), "return 1") {
		t.Fatalf("unexpected transformation:\n%s", updated)
	}
	if err := run(options{root: root, metrics: metrics, sha: "stale", check: true}); err == nil {
		t.Fatal("stale metrics evidence was accepted")
	}
}
