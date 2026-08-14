package main

import (
	"strings"
	"testing"
)

func validPromotionBundleFixture() proofBundle {
	evidence, _ := validGuardianEvidenceFixture()
	bundle := validProof()
	bundle.Repository = evidence.Repository
	bundle.Event = "pull_request"
	bundle.PRNumber = evidence.PRNumber
	bundle.BaseRef = "main"
	bundle.BaseSHA = evidence.BaseSHA
	bundle.HeadSHA = evidence.HeadSHA
	bundle.Ref = "refs/pull/7/merge"
	bundle.EventRef = bundle.Ref
	bundle.CheckoutRef = bundle.HeadSHA
	bundle.RunID = 300
	bundle.RunAttempt = 1
	bundle.WorkflowSHA = bundle.HeadSHA
	for index := range bundle.Jobs {
		bundle.Jobs[index].HeadSHA = bundle.HeadSHA
		bundle.Jobs[index].RunID = bundle.RunID
		bundle.Jobs[index].RunAttempt = bundle.RunAttempt
	}
	bundle.Artifacts[0].Name = "ci-evidence-300-1"
	bundle.Artifacts[0].RunID = bundle.RunID
	bundle.Artifacts[0].RunAttempt = bundle.RunAttempt
	bundle.BranchProtection = evidence.BranchProtection
	bundle.DevBranchProtection = evidence.DevBranchProtection
	bundle.DomainEvidence = validDomainEvidence(bundle)
	bundle.GuardianEvidence = &evidence
	bundle.PromotionObservation = &promotionObservation{
		Repository: bundle.Repository, PRNumber: bundle.PRNumber, Action: "synchronize", State: "open", Mergeable: true, MergeableState: "clean",
		BaseRepo: bundle.Repository, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadRepo: bundle.Repository, HeadRef: "dev", HeadSHA: bundle.HeadSHA,
		LiveDevSHA: bundle.HeadSHA, LiveMainSHA: bundle.BaseSHA, Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: bundle.BaseSHA},
	}
	bundle.PromotionAuthorization = promotionAuthorizationFor(bundle)
	bundle.Digests.Bundle = ""
	payload, _ := marshalProof(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	bundle.PromotionAuthorization.ProofDigest = bundle.Digests.Bundle
	return bundle
}

func TestPromotionOperatorRequiresACompletePassingProof(t *testing.T) {
	bundle := validPromotionBundleFixture()
	if err := validateProof(bundle); err != nil {
		t.Fatalf("valid promotion fixture rejected: %v", err)
	}
	if !promotionOperatorReady(bundle) {
		t.Fatal("complete clean promotion proof was not authorized")
	}
	bundle.Decision = "FAIL_CLOSED"
	bundle.PromotionAuthorization = promotionAuthorizationFor(bundle)
	if promotionOperatorReady(bundle) {
		t.Fatal("green jobs with a FAIL_CLOSED proof were authorized")
	}
}

func TestPromotionOperatorRejectsDraftDirtyOrStaleObservation(t *testing.T) {
	mutations := []func(*promotionObservation){
		func(observation *promotionObservation) { observation.Draft = true },
		func(observation *promotionObservation) { observation.Mergeable = false },
		func(observation *promotionObservation) { observation.MergeableState = "behind" },
		func(observation *promotionObservation) { observation.LiveDevSHA = strings.Repeat("f", 40) },
	}
	for index, mutate := range mutations {
		bundle := validPromotionBundleFixture()
		mutate(bundle.PromotionObservation)
		if promotionOperatorReady(bundle) {
			t.Fatalf("promotion observation mutation %d was authorized", index)
		}
	}
}

func TestPromotionAuthorizationIsBoundToProofDigest(t *testing.T) {
	bundle := validPromotionBundleFixture()
	bundle.PromotionAuthorization.ProofDigest = strings.Repeat("a", 64)
	if promotionOperatorReady(bundle) {
		t.Fatal("promotion authorization with a forged proof digest was accepted")
	}
}
