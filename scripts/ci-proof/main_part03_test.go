package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCIArtifactInventoryRejectsZeroArtifacts(t *testing.T) {
	if err := validateArtifacts(nil, 1, 1); err == nil {
		t.Fatal("zero artifact inventory was accepted")
	}
}
func TestCIGateRejectionsExposeMissingArtifactAndCompleteSet(t *testing.T) {
	context := contextInput{ArtifactsStatus: "missing", FixturePaths: []string{"examples/billing/main.gooo"}}
	rejections := gateRejections(proofInputs{Context: context})
	joined := strings.Join(rejections, ",")
	for _, expected := range []string{"artifact_evidence_not_verified", "artifact_inventory_missing"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("missing artifact rejection %q was not preserved: %v", expected, rejections)
		}
	}
	for index := 1; index < len(rejections); index++ {
		if rejections[index-1] >= rejections[index] {
			t.Fatalf("rejection set is not sorted and unique: %v", rejections)
		}
	}
}
func TestCIProofJobsRejectDuplicateID(t *testing.T) {
	jobs := make([]jobInput, len(proofJobs))
	head := strings.Repeat("a", 40)
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: 1, Name: name, Status: "completed", Conclusion: "success", HeadSHA: head, RunID: 1, RunAttempt: 1}
	}
	data, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/jobs.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readJobs(filename); err == nil {
		t.Fatal("duplicate canonical job ID was accepted")
	}
}
func TestCIBranchProtectionSnapshotMismatchFailsClosed(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.Digest = "mismatch"
	if err := validateProof(bundle); err == nil {
		t.Fatal("unbound branch protection snapshot was accepted")
	}
}
func TestCIBranchProtectionSnapshotRefTamperingFailsClosed(t *testing.T) {
	for _, mutate := range []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.EventRef = "refs/pull/2/merge" },
		func(snapshot *branchProtection) { snapshot.CheckoutRef = strings.Repeat("b", 40) },
	} {
		bundle := validProof()
		mutate(&bundle.BranchProtection)
		if err := validateProof(bundle); err == nil {
			t.Fatal("tampered branch protection ref was accepted")
		}
	}
}
