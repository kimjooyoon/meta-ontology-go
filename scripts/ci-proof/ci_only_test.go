package main

import (
	"encoding/json"
	"strings"
	"testing"
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

func TestCIBranchProtectionRequiresExactGitHubAppBindings(t *testing.T) {
	bundle := validProof()
	bundle.BaseRef = "main"
	mainSnapshot := validBranchProtection(bundle)
	mainSnapshot.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	mainSnapshot.RequiredCheckBindings = requiredCheckBindingsFor(mainSnapshot.RequiredChecks)
	mainSnapshot.Digest = digestBranchProtection(mainSnapshot)
	mutations := []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].AppID = 1 },
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].Context = "other" },
		func(snapshot *branchProtection) {
			snapshot.RequiredCheckBindings[1].Context = snapshot.RequiredCheckBindings[0].Context
		},
	}
	for index, mutate := range mutations {
		snapshot := mainSnapshot
		snapshot.RequiredCheckBindings = append([]requiredCheckBinding(nil), mainSnapshot.RequiredCheckBindings...)
		mutate(&snapshot)
		if branchProtectionReadyFor(snapshot, "main") {
			t.Fatalf("invalid app binding mutation %d was accepted", index)
		}
	}
}

func TestCIBranchProtectionBindingDigestRoundTripsStructuredChecks(t *testing.T) {
	bundle := validProof()
	bundle.BaseRef = "main"
	protection := validBranchProtection(bundle)
	protection.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	protection.RequiredCheckBindings = requiredCheckBindingsFor(protection.RequiredChecks)
	protection.Digest = digestBranchProtection(protection)
	data, err := json.Marshal(protection)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip branchProtection
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if len(roundTrip.RequiredCheckBindings) != len(proofJobs)+1 || digestBranchProtection(roundTrip) != protection.Digest {
		t.Fatalf("structured app bindings were not preserved in the protection digest: %+v", roundTrip.RequiredCheckBindings)
	}
}

func TestCIBranchProtectionMissingReasonCanonicalizesForBothStatuses(t *testing.T) {
	bundle := validProof()
	bundle.BaseRef = "main"
	verifiedProtection := validBranchProtection(bundle)
	verifiedProtection.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	verifiedProtection.RequiredCheckBindings = requiredCheckBindingsFor(verifiedProtection.RequiredChecks)
	verifiedProtection.Digest = digestBranchProtection(verifiedProtection)
	verified, err := json.Marshal(verifiedProtection)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(verified), `"missing_reason":""`) {
		t.Fatalf("verified protection JSON omitted the empty missing_reason key: %s", verified)
	}
	unavailable := validBranchProtection(validProof())
	unavailable.ReadStatus = "unavailable"
	unavailable.Exists = false
	unavailable.RequiredChecks = nil
	unavailable.RequiredCheckBindings = nil
	unavailable.MissingReason = "branch_protection_token_unavailable"
	unavailable.Digest = digestBranchProtection(unavailable)
	data, err := json.Marshal(unavailable)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip branchProtection
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if roundTrip.MissingReason == "" || digestBranchProtection(roundTrip) != unavailable.Digest {
		t.Fatalf("unavailable protection missing reason was not canonicalized: %+v", roundTrip)
	}
}

func TestCIGateRejectionsUseMachineEvidenceWithoutHumanReviews(t *testing.T) {
	bundle := validProof()
	context := contextInput{
		Actor: "builder", Builder: "builder", Gate: "CI policy", BranchProtection: bundle.BranchProtection,
		ScopeDecision: "passed", FixtureStatus: "verified", SourceStatus: "verified", SemanticStatus: "verified", ProvenanceStatus: "verified",
		ArtifactsStatus: "verified", WriteEffect: "none", NoWrite: true,
		FixturePaths: []string{"examples/billing/main.gooo"}, Artifacts: bundle.Artifacts,
		MissingReasons: missingReasons{Protection: "domain_protection_observer_unavailable", Provenance: "domain_provenance_observer_unavailable"},
		Cache:          validCache(),
	}
	rejections := gateRejections(proofInputs{
		Governance: governanceInput{Promotion: promotionInput{BranchProtectionRequired: true}},
		Evidence:   evidenceInput{HeadSHA: bundle.HeadSHA}, Context: context,
	})
	for _, rejection := range rejections {
		if strings.Contains(rejection, "approval") || rejection == "missing_external_evidence" {
			t.Fatalf("obsolete human-review predicate remained: %v", rejections)
		}
	}
	if len(rejections) != 0 {
		t.Fatalf("machine-bound CI evidence was rejected: %v", rejections)
	}
	context.ProvenanceStatus = "missing"
	if got := gateRejections(proofInputs{Governance: governanceInput{Promotion: promotionInput{BranchProtectionRequired: true}}, Evidence: evidenceInput{HeadSHA: bundle.HeadSHA}, Context: context}); !containsString(got, "provenance_evidence_not_verified") {
		t.Fatalf("missing machine-bound provenance evidence was not rejected: %v", got)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
