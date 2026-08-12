package cache

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func evidenceFixture(run string) EvidenceFreshness {
	return EvidenceFreshness{
		BaseDigest: HashBytes([]byte("base")), HeadDigest: HashBytes([]byte("head")), RunID: run,
		PredecessorDigests: []Digest{HashBytes([]byte("generator")), HashBytes([]byte("semantic"))},
		SourceDigest:       HashBytes([]byte("source")), IRDigest: HashBytes([]byte("ir")),
		PolicyDigest: HashBytes([]byte("policy")), ToolchainDigest: HashBytes([]byte("go1.26.5")),
		TargetDigest: HashBytes([]byte("darwin/arm64")), BundleDigest: HashBytes([]byte("bundle-" + run)),
		EvidenceRefs: []EvidenceRef{
			{Name: "source", Digest: HashBytes([]byte("source-ref"))},
			{Name: "bundle", Digest: HashBytes([]byte("bundle-ref"))},
		},
	}
}

func cacheReceiptFixture(key Key, run string) CacheReceipt {
	evidence := evidenceFixture(run)
	return CacheReceipt{
		SchemaVersion: cacheReceiptSchemaVersion, CacheKey: key.Digest, ArtifactKind: key.ArtifactKind,
		Projection:            key.Projection,
		SemanticClosureDigest: key.SemanticClosureDigest, DependencyRoot: key.DependencyRoot,
		DirectDependencies: []Digest{HashBytes([]byte("direct"))}, PolicySchemaDigest: key.PolicySchemaDigest,
		Toolchain: key.Toolchain, Target: key.Target, BuildTagsDigest: key.BuildTagsDigest,
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

func TestCacheReceiptBindsProjectionAndContent(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "binding")
	if err := receipt.ValidateForKey(key); err != nil {
		t.Fatal(err)
	}
	if err := receipt.ValidateForData([]byte("projection")); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*CacheReceipt){
		"projection": func(r *CacheReceipt) { r.Projection = "other" },
		"artifact":   func(r *CacheReceipt) { r.ArtifactKind = "other" },
		"content":    func(r *CacheReceipt) { r.ContentDigest = HashBytes([]byte("other")) },
		"size":       func(r *CacheReceipt) { r.Size++ },
		"rebuild":    func(r *CacheReceipt) { r.Reconstructable = false },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := receipt
			mutate(&mutated)
			if name == "content" || name == "size" {
				if err := mutated.ValidateForData([]byte("projection")); !errors.Is(err, ErrInvalidReceipt) {
					t.Fatalf("mutated data receipt = %v, want ErrInvalidReceipt", err)
				}
				return
			}
			if err := mutated.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mutated identity receipt = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestEvidenceFreshnessC4RejectsStaleAndReplayTuples(t *testing.T) {
	current := evidenceFixture("run-current")
	for name, mutate := range map[string]func(*EvidenceFreshness){
		"base":  func(e *EvidenceFreshness) { e.BaseDigest = HashBytes([]byte("new-base")) },
		"head":  func(e *EvidenceFreshness) { e.HeadDigest = HashBytes([]byte("new-head")) },
		"run":   func(e *EvidenceFreshness) { e.RunID = "run-other" },
		"prior": func(e *EvidenceFreshness) { e.PredecessorDigests[0] = HashBytes([]byte("other")) },
	} {
		t.Run(name, func(t *testing.T) {
			stale := canonicalEvidence(current)
			mutate(&stale)
			if stale.Matches(current) {
				t.Fatal("stale evidence matched current tuple")
			}
		})
	}
}

func TestCacheReceiptC3C5RequiresImmutableEvidenceBundle(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "run-1")
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("replayed receipt = %v, want ErrReceiptReplay", err)
	}
	receipts, err := cache.Receipts()
	if err != nil || len(receipts) != 1 {
		t.Fatalf("receipt log = %d, %v", len(receipts), err)
	}
	missing := receipt
	missing.Evidence.BundleDigest = ""
	if _, err := missing.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("missing bundle = %v, want ErrInvalidReceipt", err)
	}
	zero := receipt
	zero.Evidence.BundleDigest = Digest("0000000000000000000000000000000000000000000000000000000000000000")
	if _, err := zero.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("zero bundle = %v, want ErrInvalidReceipt", err)
	}
	unknownRef := receipt
	unknownRef.EvidenceRefs = nil
	unknownRef.Evidence.EvidenceRefs = nil
	if _, err := unknownRef.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("missing evidence ref = %v, want ErrInvalidReceipt", err)
	}
	zeroRef := receipt
	zeroRef.EvidenceRefs = []EvidenceRef{{Name: "source", Digest: Digest("0000000000000000000000000000000000000000000000000000000000000000")}}
	zeroRef.Evidence.EvidenceRefs = zeroRef.EvidenceRefs
	if _, err := zeroRef.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("zero evidence ref = %v, want ErrInvalidReceipt", err)
	}
	benchmark := BenchmarkReceipt{SchemaVersion: benchmarkReceiptSchemaVersion, Fixture: "partial",
		BaseDigest: key.Digest, HeadDigest: key.Digest, RunID: "bench-1", Filesystem: "local",
		JobIDs:         map[string]string{"policy": "1", "semantic": "2", "format": "3", "race": "4", "test": "5"},
		P50Nanoseconds: 10, P95Nanoseconds: 20}
	if err := benchmark.Validate(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("incomplete benchmark = %v, want ErrInvalidReceipt", err)
	}
	benchmark.JobIDs["vet"] = "6"
	if err := benchmark.Validate(); err != nil {
		t.Fatalf("complete benchmark = %v", err)
	}
}
