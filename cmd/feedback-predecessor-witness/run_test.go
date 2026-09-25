package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func TestRunWritesReplayVerifiedReceiptOutsideRoot(t *testing.T) {
	root := t.TempDir()
	input := writeFixture(t, root, false)
	report := filepath.Join(t.TempDir(), "receipt.json")
	failClosed, err := run(config{root: root, input: input, report: report, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if failClosed {
		t.Fatal("selected predecessor was rejected")
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var output receipt
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}
	if !output.ReplayVerified ||
		output.Report.Decision != feedbackpredecessor.DecisionSelected ||
		output.RepositoryWrites != 0 || output.ReceiptDigest == "" {
		t.Fatalf("receipt = %#v", output)
	}
}

func TestRunCheckRejectsAmbiguousPredecessor(t *testing.T) {
	root := t.TempDir()
	input := writeFixture(t, root, true)
	report := filepath.Join(t.TempDir(), "receipt.json")
	failClosed, err := run(config{root: root, input: input, report: report, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if !failClosed {
		t.Fatal("ambiguous predecessor was accepted")
	}
}

func TestRunCheckAllowsNonPromotingBaseline(t *testing.T) {
	root := t.TempDir()
	input := writeUnsuccessfulFixture(t, root)
	report := filepath.Join(t.TempDir(), "receipt.json")
	rejected, err := run(config{root: root, input: input, report: report, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if rejected {
		t.Fatal("lower-resolution baseline was rejected")
	}
}
