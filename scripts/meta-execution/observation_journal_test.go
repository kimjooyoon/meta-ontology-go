package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestJournalPreservesEnteredBoundaryBeforeTerminalBundle(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	file, state, err := openObservationJournal(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	plan := generation.Plan{HeadSHA: "source-head", PlanDigest: "plan-binding"}
	manifest := generation.ExecutionManifest{ManifestDigest: "manifest-binding"}
	action := generation.Action{Activity: "FixtureActivity", Subject: "fixture.go",
		IndicatorID: "fixture-indicator", InputContractSourceDigest: "contract-binding"}
	trace := newMetaExecutionTrace(plan, manifest, action, 1, state)
	trace.emitActionEntered()
	data, err := os.ReadFile(manifestPath + ".trace.ndjson")
	if err != nil {
		t.Fatal(err)
	}
	var event metaExecutionTraceEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatal(err)
	}
	if event.Boundary != "ACTION_ENTERED" || event.Activity != action.Activity ||
		event.HeadSHA != plan.HeadSHA || event.ManifestDigest != manifest.ManifestDigest ||
		event.InvocationID == "" || event.InputContractSourceDigest != action.InputContractSourceDigest ||
		!event.DiagnosticOnly || event.Permission != "UNOBSERVED" {
		t.Fatalf("journal changed execution boundary: %+v", event)
	}
}

func TestJournalDoesNotMixSequentialAttemptEvents(t *testing.T) {
	manifest := filepath.Join(t.TempDir(), "manifest.json")
	first, firstState, err := openObservationJournal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.WriteString("previous attempt\n"); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, secondState, err := openObservationJournal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	data, err := os.ReadFile(manifest + ".trace.ndjson")
	if err != nil || len(data) != 0 || firstState.invocationID == secondState.invocationID {
		t.Fatalf("journal inherited previous attempt: %q %v", data, err)
	}
}
