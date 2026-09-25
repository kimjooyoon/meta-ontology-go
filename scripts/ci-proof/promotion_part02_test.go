package main

import (
	"strings"
	"testing"
)

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
