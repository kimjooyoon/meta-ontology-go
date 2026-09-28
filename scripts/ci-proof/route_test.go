package main

import "testing"

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
	if isPromotionBundle(bundle) {
		t.Fatal("main push was classified as a promotion")
	}
	if promotionAuthorizationFor(bundle) != nil {
		t.Fatal("main push acquired a promotion authorization")
	}
}

func TestDeclaredContextRouteMustMatchTuple(t *testing.T) {
	context := contextInput{Event: "push", BaseRef: "main", HeadRef: "dev", Route: proofRoutePromotionMain}
	if validContextProofRoute(context) {
		t.Fatal("forged promotion route matched a main push")
	}
	context.Route = proofRouteProtectedPushMain
	if !validContextProofRoute(context) {
		t.Fatal("deterministic main push route was rejected")
	}
}
