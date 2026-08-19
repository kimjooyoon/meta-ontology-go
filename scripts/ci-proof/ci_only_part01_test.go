package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCIMachineProofDoesNotRequireProtectionOrReviews(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.ReadStatus = "unavailable"
	bundle.BranchProtection.Exists = false
	bundle.BranchProtection.Strict = false
	bundle.BranchProtection.RequiredChecks = nil
	bundle.BranchProtection.EnforceAdmins = false
	bundle.BranchProtection.MissingReason = "trusted_guardian_required"
	bundle.BranchProtection.Digest = digestBranchProtection(bundle.BranchProtection)
	bundle.Digests.Bundle = ""
	payload, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Digests.Bundle = digestBytes(payload)
	if err := validateProof(bundle); err != nil {
		t.Fatalf("machine proof incorrectly required human review or observable protection: %v", err)
	}
}
func TestCIBranchProtectionRequiresCIOnlySnapshot(t *testing.T) {
	bundle := validProof()
	if branchProtectionReady(bundle.BranchProtection) {
		t.Fatal("feature proof treated unobserved protection as promotion-ready")
	}
	bundle.BaseRef = "main"
	snapshot := validBranchProtection(bundle)
	snapshot.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	snapshot.RequiredCheckBindings = requiredCheckBindingsFor(snapshot.RequiredChecks)
	snapshot.Digest = digestBranchProtection(snapshot)
	if !branchProtectionReadyFor(snapshot, "main") {
		t.Fatal("trusted main branch protection snapshot was not promotion-ready")
	}
	for _, mutate := range []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.RequiredReviews = 1 },
		func(snapshot *branchProtection) { snapshot.DismissStaleReviews = true },
		func(snapshot *branchProtection) { snapshot.RequireLastPushApproval = true },
	} {
		mutated := snapshot
		mutate(&mutated)
		if branchProtectionReadyFor(mutated, "main") {
			t.Fatal("human-review branch protection predicate was accepted by CI-only gate")
		}
	}
}
func TestFeatureBranchProtectionRejectsObservedFreshness(t *testing.T) {
	bundle := validProof()
	baseContext := contextInput{BaseRef: "dev", EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef}
	evidence := evidenceInput{Repository: bundle.Repository, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, Attempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, Digests: evidenceDigests{Policy: bundle.Digests.Policy}}
	for _, test := range []struct {
		name   string
		mutate func(*branchProtection)
	}{
		{name: "observed_at", mutate: func(snapshot *branchProtection) { snapshot.ObservedAt = stringPointer("2026-08-14T00:00:00Z") }},
		{name: "valid_until", mutate: func(snapshot *branchProtection) { snapshot.ValidUntil = stringPointer("2026-08-14T00:10:00Z") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			snapshot := bundle.BranchProtection
			test.mutate(&snapshot)
			snapshot.Digest = digestBranchProtection(snapshot)
			if err := validateBranchProtectionAt(snapshot, evidence, baseContext, time.Date(2026, time.August, 14, 0, 5, 0, 0, time.UTC)); err == nil {
				t.Fatal("feature proof accepted freshness on an unobserved branch-protection snapshot")
			}
		})
	}
}
