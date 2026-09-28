package main

import (
	"strings"
	"testing"
)

func rebindPromotionFixture(bundle *proofBundle) {
	bundle.PromotionAuthorization = promotionAuthorizationFor(*bundle)
	bundle.Digests.Bundle = ""
	payload, _ := marshalProof(*bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	bundle.PromotionAuthorization.ProofDigest = bundle.Digests.Bundle
}

func TestPromotionOperatorRejectsDraftDirtyOrStaleObservation(t *testing.T) {
	mutations := []func(*promotionObservation){
		func(observation *promotionObservation) { observation.Draft = true },
		func(observation *promotionObservation) { observation.Mergeable = false },
		func(observation *promotionObservation) { observation.MergeableState = "behind" },
		func(observation *promotionObservation) { observation.LiveDevSHA = strings.Repeat("f", 40) },
		func(observation *promotionObservation) { observation.LiveDevTreeSHA = strings.Repeat("f", 40) },
		func(observation *promotionObservation) { observation.Topology.MergeBaseSHA = strings.Repeat("f", 40) },
		func(observation *promotionObservation) { observation.HeadTreeSHA = strings.Repeat("f", 40) },
	}
	for index, mutate := range mutations {
		bundle := validPromotionBundleFixture()
		mutate(bundle.PromotionObservation)
		rebindPromotionFixture(&bundle)
		if promotionOperatorReady(bundle) {
			t.Fatalf("promotion observation mutation %d was authorized", index)
		}
	}
}

func TestSnapshotPromotionRejectsStaleSourceTreeOrWrongParent(t *testing.T) {
	mutations := []func(*promotionObservation){
		func(observation *promotionObservation) {
			observation.HeadRef = "agent/main-promotion-snapshot-" + strings.Repeat("f", 40)
		},
		func(observation *promotionObservation) { observation.HeadParentSHA = strings.Repeat("f", 40) },
		func(observation *promotionObservation) { observation.LiveDevTreeSHA = strings.Repeat("e", 40) },
		func(observation *promotionObservation) { observation.Topology.AheadBy = 2 },
	}
	for index, mutate := range mutations {
		bundle := validSnapshotPromotionBundleFixture()
		mutate(bundle.PromotionObservation)
		rebindPromotionFixture(&bundle)
		if promotionOperatorReady(bundle) {
			t.Fatalf("snapshot promotion mutation %d was authorized", index)
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
