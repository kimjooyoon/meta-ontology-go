package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"
)

func TestRunBuildsBoundResolutionInput(t *testing.T) {
	root := t.TempDir()
	coveragePath, provenancePath := writeSourceFixtures(t, root)
	output := filepath.Join(t.TempDir(), "input.json")
	if err := run(config{
		root: root, coverage: coveragePath, provenance: provenancePath,
		output: output, ciConclusion: "success",
	}); err != nil {
		t.Fatal(err)
	}
	var input artifactfeedback.ResolutionInput
	if err := readJSON(output, &input); err != nil {
		t.Fatal(err)
	}
	if input.Feedback.Cycle.Status != "BOUND" ||
		input.Feedback.Coverage.CommitSHA != input.Feedback.Cycle.HeadSHA ||
		input.Feedback.RepositoryWrites != 0 {
		t.Fatalf("resolution input = %#v", input)
	}
}

func TestRunRejectsHeadMismatch(t *testing.T) {
	root := t.TempDir()
	coveragePath, provenancePath := writeSourceFixtures(t, root)
	var provenance provenanceEnvelope
	if err := readJSON(provenancePath, &provenance); err != nil {
		t.Fatal(err)
	}
	provenance.HeadSHA = strings.Repeat("b", 40)
	writeFixtureJSON(t, provenancePath, provenance)
	err := run(config{
		root: root, coverage: coveragePath, provenance: provenancePath,
		output: filepath.Join(t.TempDir(), "input.json"), ciConclusion: "success",
	})
	if err == nil {
		t.Fatal("mismatched source heads were accepted")
	}
}
