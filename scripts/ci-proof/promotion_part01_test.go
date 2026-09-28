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
		Repository: bundle.Repository, PRNumber: bundle.PRNumber, Action: "synchronize", State: "open", Mergeable: true, MergeableState: "clean",
		BaseRepo: bundle.Repository, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadRepo: bundle.Repository, HeadRef: "dev", HeadSHA: bundle.HeadSHA,
		LiveDevSHA: bundle.HeadSHA, LiveMainSHA: bundle.BaseSHA, Topology: promotionTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: bundle.BaseSHA},
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
