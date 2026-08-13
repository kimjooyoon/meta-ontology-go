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
	bundle.BranchProtection.MissingReason = "branch_protection_token_unavailable"
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

func TestCIPromotionPredicateUsesCIOnlyProtection(t *testing.T) {
	bundle := validProof()
	promotion := promotionInput{BranchProtectionRequired: true}
	if !promotionReady(promotion, bundle.BranchProtection) {
		t.Fatal("CI-only protection snapshot was not promotion-ready")
	}
	bundle.BranchProtection.RequiredReviews = 1
	bundle.BranchProtection.Digest = digestBranchProtection(bundle.BranchProtection)
	if promotionReady(promotion, bundle.BranchProtection) {
		t.Fatal("human review requirement was accepted by CI-only promotion predicate")
	}
}

func TestCIBranchProtectionRequiresCIOnlySnapshot(t *testing.T) {
	bundle := validProof()
	if !branchProtectionReady(bundle.BranchProtection) {
		t.Fatal("CI-only branch protection snapshot was not promotion-ready")
	}
	for _, mutate := range []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.RequiredReviews = 1 },
		func(snapshot *branchProtection) { snapshot.DismissStaleReviews = true },
		func(snapshot *branchProtection) { snapshot.RequireLastPushApproval = true },
	} {
		snapshot := bundle.BranchProtection
		mutate(&snapshot)
		if branchProtectionReady(snapshot) {
			t.Fatal("human-review branch protection predicate was accepted by CI-only gate")
		}
	}
}

func TestCIBranchProtectionRequiresExactGitHubAppBindings(t *testing.T) {
	bundle := validProof()
	mutations := []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].AppID = 1 },
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].Context = "other" },
		func(snapshot *branchProtection) {
			snapshot.RequiredCheckBindings[1].Context = snapshot.RequiredCheckBindings[0].Context
		},
	}
	for index, mutate := range mutations {
		snapshot := bundle.BranchProtection
		snapshot.RequiredCheckBindings = append([]requiredCheckBinding(nil), bundle.BranchProtection.RequiredCheckBindings...)
		mutate(&snapshot)
		if branchProtectionReady(snapshot) {
			t.Fatalf("invalid app binding mutation %d was accepted", index)
		}
	}
}

func TestCIBranchProtectionBindingDigestRoundTripsStructuredChecks(t *testing.T) {
	bundle := validProof()
	data, err := json.Marshal(bundle.BranchProtection)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip branchProtection
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if len(roundTrip.RequiredCheckBindings) != len(proofJobs) || digestBranchProtection(roundTrip) != bundle.BranchProtection.Digest {
		t.Fatalf("structured app bindings were not preserved in the protection digest: %+v", roundTrip.RequiredCheckBindings)
	}
}

func TestCIBranchProtectionMissingReasonCanonicalizesForBothStatuses(t *testing.T) {
	bundle := validProof()
	verified, err := json.Marshal(bundle.BranchProtection)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(verified), `"missing_reason":""`) {
		t.Fatalf("verified protection JSON omitted the empty missing_reason key: %s", verified)
	}
	unavailable := bundle.BranchProtection
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

func TestCIMachineBoundPromotionRejectsUnavailableProtection(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.ReadStatus = "unavailable"
	bundle.BranchProtection.Exists = false
	bundle.BranchProtection.MissingReason = "branch_protection_token_unavailable"
	context := contextInput{
		Repository: bundle.Repository, Event: bundle.Event, Ref: bundle.Ref, EventRef: bundle.EventRef,
		CheckoutRef: bundle.CheckoutRef, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA,
		HeadSHA: bundle.HeadSHA, WorkflowSHA: bundle.WorkflowSHA, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt,
		BranchProtection: bundle.BranchProtection, ArtifactsStatus: "verified",
		ProvenanceStatus: "verified",
	}
	inputs := proofInputs{
		Governance: governanceInput{Promotion: promotionInput{Source: "dev", Target: "main", RequiredChecks: append([]string(nil), proofJobs...), BranchProtectionRequired: true}},
		Evidence:   evidenceInput{Jobs: bundle.Jobs}, Jobs: bundle.Jobs, Context: context,
	}
	if machineBoundPromotionReady(inputs) {
		t.Fatal("unavailable protection was treated as promotion-ready")
	}
	context.BranchProtection.MissingReason = "unknown"
	if machineBoundPromotionReady(proofInputs{Jobs: bundle.Jobs, Context: context}) {
		t.Fatal("unknown protection observer gap was accepted")
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
