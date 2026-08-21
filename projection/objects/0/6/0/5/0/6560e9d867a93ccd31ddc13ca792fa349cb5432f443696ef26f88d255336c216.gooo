package cache

import (
	"encoding/json"
	"slices"
	"testing"
)

func cacheReceiptFixture(key Key, run string) CacheReceipt {
	evidence := evidenceFixture(run)
	return CacheReceipt{
		SchemaVersion: cacheReceiptSchemaVersion, CacheKey: key.Digest, Domain: key.Domain,
		KeyVersion: key.Version, HostStage: key.HostStage, ArtifactKind: key.ArtifactKind,
		Projection:            key.Projection,
		SemanticClosureDigest: key.SemanticClosureDigest, DependencyRoot: key.DependencyRoot,
		DirectDependencies: []Digest{HashBytes([]byte("direct"))}, PolicySchemaDigest: key.PolicySchemaDigest,
		Toolchain: key.Toolchain, Target: key.Target, BuildTagsDigest: key.BuildTagsDigest,
		OptionsDigest: key.OptionsDigest,
		ContentDigest: HashBytes([]byte("projection")), Size: int64(len("projection")), Reconstructable: true,
		EvidenceRefs: evidence.EvidenceRefs, ProducerHost: "go-hosted", Status: ReceiptRecomputed,
		Evidence: evidence,
	}
}
func TestCacheReceiptCanonicalizationIsPresentationStable(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	left := cacheReceiptFixture(key, "canonical")
	right := cacheReceiptFixture(key, "canonical")
	right.EvidenceRefs = append([]EvidenceRef(nil), right.EvidenceRefs...)
	slices.Reverse(right.Evidence.PredecessorDigests)
	slices.Reverse(right.Evidence.EvidenceRefs)
	slices.Reverse(right.EvidenceRefs)
	sealedLeft, err := left.Seal()
	if err != nil {
		t.Fatal(err)
	}
	sealedRight, err := right.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if sealedLeft.ReceiptDigest != sealedRight.ReceiptDigest {
		t.Fatalf("presentation changed receipt digest: %s != %s", sealedLeft.ReceiptDigest, sealedRight.ReceiptDigest)
	}
	leftJSON, err := json.Marshal(sealedLeft)
	if err != nil {
		t.Fatal(err)
	}
	rightJSON, err := json.Marshal(sealedRight)
	if err != nil {
		t.Fatal(err)
	}
	if string(leftJSON) != string(rightJSON) {
		t.Fatal("canonical receipt JSON differs by presentation order")
	}
}
