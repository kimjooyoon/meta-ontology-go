package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestCanonicalProcessReplayIgnoresWorkspaceAndRawDigest(t *testing.T) {
	first := []byte("failure at /tmp/work-one/src\n")
	second := []byte("failure at /tmp/work-two/src\n")
	canonicalFirst := canonicalProcessBytes("/tmp/work-one", first)
	canonicalSecond := canonicalProcessBytes("/tmp/work-two", second)
	if !bytes.Equal(canonicalFirst, canonicalSecond) {
		t.Fatal("workspace-only process evidence did not replay canonically")
	}

	different := canonicalProcessBytes("/tmp/work-two", []byte("different failure\n"))
	if bytes.Equal(canonicalFirst, different) {
		t.Fatal("canonical process evidence drift was accepted")
	}
}

func TestOperationReplayProjectionExcludesRawDigests(t *testing.T) {
	process := generation.ProcessObservation{
		Command:      []string{"go", "run", "<workspace>", "--plan", "meta-execution-function-plan.json"},
		ExitCode: 1, StdoutBytes: 2, StdoutDigest: "sha256:" + strings.Repeat("1", 64),
		StderrBytes: 3, StderrDigest: "sha256:" + strings.Repeat("2", 64),
		RawStdoutDigest: "sha256:" + strings.Repeat("3", 64),
		RawStderrDigest: "sha256:" + strings.Repeat("4", 64),
	}
	projection := operationReplayEvidenceFrom(process, process, process)
	encoded, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(strings.Repeat("3", 64))) ||
		bytes.Contains(encoded, []byte(strings.Repeat("4", 64))) {
		t.Fatal("raw process digests entered replay projection")
	}
	if !bytes.Contains(encoded, []byte(strings.Repeat("1", 64))) ||
		!bytes.Contains(encoded, []byte(strings.Repeat("2", 64))) {
		t.Fatal("canonical process digests were omitted from replay projection")
	}
	if !bytes.Contains(encoded, []byte("<workspace>")) ||
		bytes.Contains(encoded, []byte("/tmp/actual-workspace")) {
		t.Fatal("replay projection did not retain only the canonical command descriptor")
	}
}

func TestOperationReplayProjectionBindsCanonicalCommand(t *testing.T) {
	first := generation.ProcessObservation{Command: []string{"go", "run", "<workspace>"},
		StdoutDigest: "sha256:" + strings.Repeat("1", 64), StderrDigest: "sha256:" + strings.Repeat("2", 64)}
	second := first
	second.Command = []string{"go", "run", "<workspace>", "--different-option"}
	firstEncoded, err := json.Marshal(operationReplayEvidenceFrom(first, first, first))
	if err != nil {
		t.Fatal(err)
	}
	secondEncoded, err := json.Marshal(operationReplayEvidenceFrom(second, second, second))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(firstEncoded, secondEncoded) {
		t.Fatal("command descriptor drift was omitted from replay projection")
	}
}

func TestStructuredExtractorFailureCanonicalTupleBindsDerivedBlocker(t *testing.T) {
	failure := &operationError{
		stage: "derive-recipe", step: "select-declaration", reason: "NO_SAFE_DECLARATION_CAPACITY",
		class: "KNOWN_CONTRADICTION", derivedRelations: []generation.CounterexampleRelation{{
			Counterexample: "fixture.go#func:SelectedExtractedSuffix08",
			DerivedFrom: "fixture.go#func:Selected", Relation: "DERIVED_FROM",
		}},
	}
	process := generation.ProcessObservation{ExitCode: 1, StdoutBytes: 7, StderrBytes: 9}
	first := canonicalStructuredExtractorFailure("fixture.go:10:Selected", failure, process)
	second := canonicalStructuredExtractorFailure("fixture.go:10:Selected", failure, process)
	if len(first) == 0 || !bytes.Equal(first, second) {
		t.Fatal("structured extractor failure tuple was not deterministic")
	}
	failure.derivedRelations[0].Counterexample = "fixture.go#func:SelectedExtractedSuffix09"
	third := canonicalStructuredExtractorFailure("fixture.go:10:Selected", failure, process)
	if bytes.Equal(first, third) {
		t.Fatal("derived blocker drift was omitted from canonical failure evidence")
	}
}
