package main

import (
	"encoding/json"
	"strings"
	"testing"
)

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
