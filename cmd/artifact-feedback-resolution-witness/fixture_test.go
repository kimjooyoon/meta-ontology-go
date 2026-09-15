package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func writeResolutionFixture(t *testing.T, root string) string {
	t.Helper()
	head := strings.Repeat("a", 40)
	coverage := artifactcoverage.Report{
		Schema: artifactcoverage.ReportSchema, CommitSHA: head,
		Repository: "kimjooyoon/meta-ontology-go", Decision: "UNKNOWN",
		Summary: artifactcoverage.Summary{
			RequiredOperations: 5, CanonicalOperations: 5,
		},
	}
	coverage.ReportDigest = mustDigest(t, coverage)
	input := artifactfeedback.ResolutionInput{
		Feedback: artifactfeedback.Input{
			Coverage: coverage, CoverageReplayDigest: coverage.ReportDigest,
			Cycle: artifactfeedback.CycleObservation{
				Schema: artifactfeedback.CycleSchema, HeadSHA: head, Status: "BOUND",
				CIConclusion: "success", EnvelopeDigest: strings.Repeat("1", 64),
				ReplayDigest: strings.Repeat("2", 64),
			},
		},
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "input.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func mustDigest(t *testing.T, value any) string {
	t.Helper()
	digest, err := digestJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}
