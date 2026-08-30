package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func TestFoundationPromotionRouteRequiresExactIdentity(t *testing.T) {
	context := contextInput{Repository: verify.FoundationPromotionRepository, Event: "pull_request", Route: proofRouteFoundationPromotion, BaseRef: verify.FoundationPromotionBaseBranch, BaseSHA: verify.FoundationPromotionBaseSHA, HeadSHA: "a123456789012345678901234567890123456789", PRNumber: verify.FoundationPromotionPRNumber, FoundationPromotion: &foundationPromotionEvidence{HeadRef: verify.FoundationPromotionHeadBranch, HeadSHA: "a123456789012345678901234567890123456789"}}
	if !validContextProofRoute(context) {
		t.Fatal("exact Foundation promotion identity was rejected")
	}
	context.PRNumber = 601
	if validContextProofRoute(context) {
		t.Fatal("reused Foundation route identity was accepted")
	}
}
