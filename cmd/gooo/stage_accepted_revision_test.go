package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestRunStageAcceptedRevision(t *testing.T) {
	root := t.TempDir()
	candidatePath := filepath.Join(root, "candidate.gooo")
	comparisonPath := filepath.Join(root, "comparison.json")
	out := filepath.Join(root, "stage")
	candidate := []byte("activity ObserveRepair\n")
	if err := os.WriteFile(candidatePath, candidate, 0o644); err != nil {
		t.Fatal(err)
	}
	comparison := valueexecution.AcceptedRevisionNextRunComparison{
		Schema:                       valueexecution.AcceptedRevisionNextRunComparisonSchema,
		State:                        valueexecution.ReplayClosed,
		Outcome:                      valueexecution.NextRunOutcomeImproved,
		Reason:                       "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN",
		SourceDigest:                 "source-digest",
		CandidateSourceDigest:        digestForStageTest(candidate),
		RevisionCandidateID:          "candidate-1",
		Activity:                     "ObserveRepair",
		InputDigest:                  "input-digest",
		BaselineFailureCode:          "EXPECTED_FAILURE",
		BaselineExecutionDigest:      "baseline-execution",
		AcceptedExecutionDigest:      "accepted-execution",
		NextCandidateExecutionDigest: "accepted-execution",
	}
	encoded, err := json.Marshal(comparison)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(comparisonPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if got := runStageAcceptedRevision([]string{candidatePath, "--comparison", comparisonPath, "--out", out}, OSFileReader{}, &stdout, &stderr); got != exitOK {
		t.Fatalf("runStageAcceptedRevision() = %d, stderr=%s", got, stderr.String())
	}
	var stage valueexecution.AcceptedRevisionNextRunStage
	if err := json.Unmarshal(stdout.Bytes(), &stage); err != nil {
		t.Fatal(err)
	}
	if stage.Decision != valueexecution.AcceptedRevisionNextRunStageDecision || stage.ComparisonDigest == "" {
		t.Fatalf("unexpected stage: %#v", stage)
	}
	gotCandidate, err := os.ReadFile(filepath.Join(out, "candidate.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotCandidate, candidate) {
		t.Fatal("staged candidate differs from source candidate")
	}
}

func TestRunStageAcceptedRevisionFailsClosedBeforeWriting(t *testing.T) {
	root := t.TempDir()
	candidatePath := filepath.Join(root, "candidate.gooo")
	comparisonPath := filepath.Join(root, "comparison.json")
	out := filepath.Join(root, "stage")
	if err := os.WriteFile(candidatePath, []byte("candidate"), 0o644); err != nil {
		t.Fatal(err)
	}
	comparison, _ := json.Marshal(valueexecution.AcceptedRevisionNextRunComparison{State: valueexecution.ReplayUnknown})
	if err := os.WriteFile(comparisonPath, comparison, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if got := runStageAcceptedRevision([]string{candidatePath, "--comparison", comparisonPath, "--out", out}, OSFileReader{}, &stdout, &stderr); got == exitOK {
		t.Fatal("expected fail-closed staging")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("staging output exists after rejection: %v", err)
	}
}

func digestForStageTest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
