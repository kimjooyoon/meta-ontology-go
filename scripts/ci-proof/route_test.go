package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func TestProofRouteClassifierSeparatesMainPush(t *testing.T) {
	cases := []struct {
		event string
		base  string
		want  string
	}{
		{"pull_request", "dev", proofRouteFeatureDev},
		{"pull_request", "main", proofRoutePromotionMain},
		{"push", "dev", proofRouteProtectedPushDev},
		{"push", "main", proofRouteProtectedPushMain},
	}
	for _, item := range cases {
		got, err := classifyProofRoute(item.event, item.base)
		if err != nil || got != item.want {
			t.Fatalf("route %s:%s = %q, %v", item.event, item.base, got, err)
		}
	}
}

func TestMainPushCannotAcquirePromotionCapabilities(t *testing.T) {
	bundle := validProof()
	bundle.Event = "push"
	bundle.PRNumber = 0
	bundle.BaseRef = "main"
	bundle.Ref = "refs/heads/main"
	bundle.EventRef = bundle.Ref
	bundle.BranchProtection = validBranchProtection(bundle)
	if isPromotionBundle(bundle) {
		t.Fatal("main push was classified as a promotion")
	}
	if promotionAuthorizationFor(bundle) != nil {
		t.Fatal("main push acquired a promotion authorization")
	}
	if err := validateGuardianEvidence(nil, bundle); err != nil {
		t.Fatalf("main push required promotion Guardian evidence: %v", err)
	}
}

func TestDeclaredContextRouteMustMatchTuple(t *testing.T) {
	context := contextInput{Event: "push", BaseRef: "main", Route: proofRoutePromotionMain}
	if validContextProofRoute(context) {
		t.Fatal("forged promotion route matched a main push")
	}
	context.Route = proofRouteProtectedPushMain
	if !validContextProofRoute(context) {
		t.Fatal("deterministic main push route was rejected")
	}
}

func TestFoundationPromotionRouteRequiresExactIdentity(t *testing.T) {
	context := contextInput{
		Repository: verify.FoundationPromotionRepository,
		Event:      "pull_request",
		Route:      proofRouteFoundationPromotion,
		BaseRef:    verify.FoundationPromotionBaseBranch,
		BaseSHA:    verify.FoundationPromotionBaseSHA,
		HeadSHA:    "a123456789012345678901234567890123456789",
		PRNumber:   verify.FoundationPromotionPRNumber,
		FoundationPromotion: &foundationPromotionEvidence{
			HeadRef: verify.FoundationPromotionHeadBranch,
			HeadSHA: "a123456789012345678901234567890123456789",
		},
	}
	if !validContextProofRoute(context) {
		t.Fatal("exact Foundation promotion identity was rejected")
	}
	context.PRNumber = 601
	if validContextProofRoute(context) {
		t.Fatal("reused Foundation route identity was accepted")
	}
}
