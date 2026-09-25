package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCIReceiptRejectsTamperedBindingEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*provenanceReceipt)
	}{
		{name: "repository", mutate: func(receipt *provenanceReceipt) { receipt.Repository = "other/repo" }},
		{name: "job head", mutate: func(receipt *provenanceReceipt) { receipt.Jobs[0].HeadSHA = strings.Repeat("b", 40) }},
		{name: "branch protection", mutate: func(receipt *provenanceReceipt) { receipt.BranchProtection.MissingReason = "tampered" }},
		{name: "domain evidence", mutate: func(receipt *provenanceReceipt) { receipt.DomainEvidence.ObserverStatus = "verified" }},
		{name: "proof digest", mutate: func(receipt *provenanceReceipt) { receipt.Digests.Bundle = strings.Repeat("d", 64) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := validProof()
			receipt := makeReceipt(bundle, contextInput{})
			receipt = roundTripReceipt(t, receipt)
			test.mutate(&receipt)
			filename := writeReceiptFixture(t, receipt)
			if err := verifyReceipt(filename, bundle); err == nil {
				t.Fatal("tampered receipt evidence was accepted")
			}
		})
	}
}
func roundTripReceipt(t *testing.T, receipt provenanceReceipt) provenanceReceipt {
	t.Helper()
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var independent provenanceReceipt
	if err := json.Unmarshal(data, &independent); err != nil {
		t.Fatal(err)
	}
	return independent
}
func writeReceiptFixture(t *testing.T, receipt provenanceReceipt) string {
	t.Helper()
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/receipt.jsonl"
	if err := os.WriteFile(filename, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return filename
}
func TestCICacheC5MissingArtifactFailsClosed(t *testing.T) {
	if err := validateArtifacts([]artifactInput{{ID: 1, Name: "ci-evidence-1-1", Size: 0, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 1, RunAttempt: 1}}, 1, 1); err == nil {
		t.Fatal("zero-sized artifact was accepted")
	}
}
func TestCIArtifactDigestMissingFailsClosed(t *testing.T) {
	if err := validateArtifacts([]artifactInput{{ID: 1, Name: "ci-evidence-1-1", Size: 1, RunID: 1, RunAttempt: 1}}, 1, 1); err == nil {
		t.Fatal("artifact without digest was accepted")
	}
}
func TestCIArtifactDigestAcceptsGitHubSHA256Form(t *testing.T) {
	artifact := artifactInput{ID: 1, Name: "ci-evidence-1-1", Size: 1, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 1, RunAttempt: 1}
	if err := validateArtifacts([]artifactInput{artifact}, 1, 1); err != nil {
		t.Fatalf("GitHub SHA-256 artifact digest was rejected: %v", err)
	}
}
