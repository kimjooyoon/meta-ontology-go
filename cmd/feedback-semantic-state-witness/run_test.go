package main

import (
	"path/filepath"
	"testing"
)

func TestRunWritesReadySemanticReceiptOutsideRoot(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "FIXED_POINT")
	output := filepath.Join(t.TempDir(), "semantic.json")
	failClosed, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if failClosed {
		t.Fatal("ready semantic state was rejected")
	}
	receipt, _, err := readJSON[semanticReceipt](output)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Report.Decision != "READY" || receipt.Report.Snapshot == nil ||
		receipt.Report.Snapshot.Decision != "FIXED_POINT" || !receipt.ReplayVerified ||
		receipt.RepositoryWrites != 0 || receipt.ReceiptDigest == "" {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func TestRunCheckRejectsUnknownSemanticDecision(t *testing.T) {
	root := t.TempDir()
	input, predecessor := writeSemanticFixture(t, root, "UNKNOWN")
	output := filepath.Join(t.TempDir(), "semantic.json")
	failClosed, err := run(config{root: root, input: input,
		predecessorReceipt: predecessor, report: output, check: true})
	if err != nil {
		t.Fatal(err)
	}
	if !failClosed {
		t.Fatal("unknown semantic decision was accepted")
	}
}
