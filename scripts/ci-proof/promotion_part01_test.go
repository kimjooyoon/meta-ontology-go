package main

import (
	"strings"
	"testing"
)

func validPromotionBundleFixture() proofBundle {
	bundle := validProof()
	bundle.Repository = "owner/repo"
	bundle.Event = "pull_request"
	bundle.PRNumber = 7
	bundle.BaseRef = "main"
	bundle.HeadRef = "dev"
	bundle.BaseSHA = strings.Repeat("b", 40)
	bundle.HeadSHA = strings.Repeat("a", 40)
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
	bundle.DomainEvidence = validDomainEvidence(bundle)
	bundle.PromotionObservation = &promotionObservation{
		Repository: bundle.Repository, PRNumber: bundle.PRNumber, Action: "synchronize", Mode: "DIRECT_DEV_HEAD", State: "open", Mergeable: true, MergeableState: "clean",
		BaseRepo: bundle.Repository, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadRepo: bundle.Repository, HeadRef: "dev", HeadSHA: bundle.HeadSHA,
		HeadParentSHA: bundle.BaseSHA, HeadTreeSHA: strings.Repeat("d", 40), LiveDevSHA: bundle.HeadSHA, LiveDevTreeSHA: strings.Repeat("d", 40), LiveMainSHA: bundle.BaseSHA,
		Topology: promotionTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: bundle.BaseSHA},
	}
	bundle.PromotionAuthorization = promotionAuthorizationFor(bundle)
	bundle.Digests.Bundle = ""
	payload, _ := marshalProof(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	bundle.PromotionAuthorization.ProofDigest = bundle.Digests.Bundle
	return bundle
}

func validSnapshotPromotionBundleFixture() proofBundle {
	bundle := validPromotionBundleFixture()
	observation := bundle.PromotionObservation
	bundle.HeadSHA = strings.Repeat("c", 40)
	bundle.HeadRef = "agent/main-promotion-snapshot-" + observation.LiveDevSHA
	bundle.CheckoutRef = bundle.HeadSHA
	bundle.WorkflowSHA = bundle.HeadSHA
	for index := range bundle.Jobs {
		bundle.Jobs[index].HeadSHA = bundle.HeadSHA
	}
	observation.Mode = "DEV_TREE_SNAPSHOT"
	observation.HeadRef = bundle.HeadRef
	observation.HeadSHA = bundle.HeadSHA
	observation.HeadParentSHA = bundle.BaseSHA
	observation.HeadTreeSHA = observation.LiveDevTreeSHA
	observation.Topology = promotionTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: bundle.BaseSHA}
	bundle.DomainEvidence = validDomainEvidence(bundle)
	bundle.PromotionAuthorization = promotionAuthorizationFor(bundle)
	bundle.Digests.Bundle = ""
	payload, _ := marshalProof(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	bundle.PromotionAuthorization.ProofDigest = bundle.Digests.Bundle
	return bundle
}

func TestPromotionOperatorAcceptsExactDevTreeSnapshotOnMainLineage(t *testing.T) {
	bundle := validSnapshotPromotionBundleFixture()
	if err := validateProof(bundle); err != nil {
		t.Fatalf("valid dev-tree snapshot promotion fixture rejected: %v", err)
	}
	if !promotionOperatorReady(bundle) {
		t.Fatal("exact dev-tree snapshot with a direct main parent was not authorized")
	}
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
