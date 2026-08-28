package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestClaimDependencyExampleReconstructsTypedEdges(t *testing.T) {
	source, err := os.ReadFile("testdata/claim-dependency/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(source, []byte(claimDependencyCandidateID)) {
		t.Fatal("candidate id is not directly mapped in Gooo source")
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	report := resolveClaimDependencies("main.gooo", source, file)
	if report.Decision != claimDependencyObserved || report.Resolution.State != claimStateClosed || report.Summary.ActivitiesObserved != 6 || report.Summary.RecoverableRoots != 1 || report.Summary.TypedDeclarations != 5 || report.Summary.DependencyInputs != 8 || report.Summary.TypedEdges != 8 || report.Summary.EdgeKindsObserved != 4 || report.Summary.UnresolvedInputs != 0 || report.Summary.CyclicActivities != 0 {
		t.Fatalf("claim dependency denominator changed: %#v", report)
	}
	if report.KindCounts.Requires != 3 || report.KindCounts.Supports != 2 || report.KindCounts.Contradicts != 2 || report.KindCounts.FailureEntailment != 1 || len(report.Indicators) != 8 {
		t.Fatalf("typed edge reconstruction changed: %#v", report)
	}
	declared := make(map[string]bool)
	for _, node := range report.Nodes {
		declared[node.Activity] = true
	}
	for _, indicator := range report.Indicators {
		for _, activity := range indicator.Activities {
			if !declared[activity] {
				t.Fatalf("indicator %q detached from Gooo activity %q", indicator.ID, activity)
			}
		}
	}
}
