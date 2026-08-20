package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestProofBundleValidatesAndPreservesReceiptSchema(t *testing.T) {
	bundle := validProof()
	if err := validateProof(bundle); err != nil {
		t.Fatal(err)
	}
	receipt := makeReceipt(bundle, contextInput{})
	if receipt.Schema != receiptSchema || receipt.Relation != "conformance" || receipt.Repository != bundle.Repository {
		t.Fatal("receipt schema or relation changed")
	}
}

func TestOldProofAndReceiptSchemasFailClosed(t *testing.T) {
	bundle := validProof()
	bundle.Schema = "gooo/ci-proof/v2"
	if err := validateProof(bundle); err == nil {
		t.Fatal("old proof schema was accepted after GuardianEvidence contract migration")
	}
	bundle = validProof()
	receipt := makeReceipt(bundle, contextInput{})
	receipt.Schema = "gooo/provenance-receipt/v2"
	filename := writeReceiptFixture(t, receipt)
	if err := verifyReceipt(filename, bundle); err == nil {
		t.Fatal("old receipt schema was accepted after GuardianEvidence contract migration")
	}
}

func TestProofUnknownFieldsFailClosed(t *testing.T) {
	bundle := validProof()
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	document["legacy_guardian_evidence"] = map[string]any{"decision": "PASS"}
	data, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/proof.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readStrictJSON[proofBundle](filename); err == nil {
		t.Fatal("proof unknown field was accepted")
	}
}

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
		jobs[index] = jobInput{ID: 1, Name: name, Status: stringPointer("completed"), Conclusion: stringPointer("success"), HeadSHA: head, RunID: 1, RunAttempt: 1, ObservationState: apiTerminalSuccess}
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

func TestCIBranchProtectionUnavailableIsNotReady(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.Exists = false
	bundle.BranchProtection.ReadStatus = "unavailable"
	bundle.BranchProtection.MissingReason = "branch_protection_token_unavailable"
	if branchProtectionReady(bundle.BranchProtection) {
		t.Fatal("unavailable branch protection snapshot was promotion-ready")
	}
}

func TestCIMissingReasonIsRequiredForUnavailableEvidence(t *testing.T) {
	bundle := validProof()
	bundle.MissingReasons.Protection = ""
	bundle.DomainEvidence.MissingReasons.Protection = ""
	if err := validateProof(bundle); err == nil {
		t.Fatal("unavailable protection without a reason was accepted")
	}
}

func TestCIDomainEvidenceOutputTamperingFailsClosed(t *testing.T) {
	bundle := validProof()
	bundle.DomainEvidence.CLI.Output += "tampered"
	if err := validateProof(bundle); err == nil {
		t.Fatal("tampered CLI domain evidence was accepted")
	}
}

func TestCIDomainEvidenceCanonicalDigestOmitsDeferredFixture(t *testing.T) {
	bundle := validProof()
	data, err := json.Marshal(bundle.DomainEvidence)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"graph":{"command":"go run ./cmd/gooo graph-dump examples/billing/main.gooo","fixture"`) {
		t.Fatal("deferred graph fixture was serialized despite being unavailable")
	}
	if err := validateProof(bundle); err != nil {
		t.Fatalf("canonical deferred graph evidence was rejected: %v", err)
	}
}

func TestCIRefSeparationRejectsCheckoutMismatch(t *testing.T) {
	bundle := validProof()
	bundle.CheckoutRef = strings.Repeat("b", 40)
	if err := validateProof(bundle); err == nil {
		t.Fatal("checkout ref mismatch was accepted")
	}
}

func TestCIRefSeparationRejectsEventRefMismatch(t *testing.T) {
	bundle := validProof()
	bundle.EventRef = "refs/pull/2/merge"
	if err := validateProof(bundle); err == nil {
		t.Fatal("event ref mismatch was accepted")
	}
}

func TestCITerminalJobSnapshotRejectsInProgress(t *testing.T) {
	jobs := make([]jobInput, len(proofJobs))
	head := strings.Repeat("a", 40)
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Status: stringPointer("completed"), Conclusion: stringPointer("success"), HeadSHA: head}
	}
	jobs[len(jobs)-1].Status = stringPointer("in_progress")
	data, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/jobs.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readJobs(filename); err == nil {
		t.Fatal("in-progress canonical job was accepted")
	}
}
