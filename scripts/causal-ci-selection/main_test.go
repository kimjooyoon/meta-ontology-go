package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
)

func TestProcessObservationRejectsShellContradictions(t *testing.T) {
	value := causalci.ConsumerAdjudication{ExitCode: 0, Result: causalci.ExecutionPass, ConsumerIdentity: "gooo://consumer/causal-ci-selection/process"}
	if processObservationConformant(value, 1, nil) {
		t.Fatal("self-reported PASS with a nonzero OS exit was accepted")
	}
	if processObservationConformant(value, 0, []byte("process failure")) {
		t.Fatal("self-reported exit 0 with failure stderr was accepted")
	}
	if !processObservationConformant(value, 0, nil) {
		t.Fatal("matching PASS process observation was rejected")
	}
}

func TestSourceBoundGoVersion(t *testing.T) {
	root := t.TempDir()
	original := []byte("module example.org/causal\n\ngo 1.27.3\n")
	if err := os.WriteFile(filepath.Join(root, "go.mod"), original, 0o644); err != nil {
		t.Fatal(err)
	}
	expected, err := sourceBoundGoVersion(root, bytesDigest(original))
	if err != nil || expected != "go1.27.3" {
		t.Fatalf("source-bound expectation: %q, %v", expected, err)
	}
	if !exactGoVersion("go version go1.27.3 linux/amd64", expected) ||
		exactGoVersion("go version go1.27.4 linux/amd64", expected) {
		t.Fatal("runtime version comparison lost its independent observation")
	}
	if _, err := sourceBoundGoVersion(root, bytesDigest([]byte("other module"))); err == nil {
		t.Fatal("changed go.mod preimage accepted")
	}
}
