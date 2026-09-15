package cache

import (
	"errors"
	"testing"
)

func TestGetProjectionIfFreshRejectsRefMutationsWithoutWrites(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*EvidenceFreshness){
		"event ref":    func(e *EvidenceFreshness) { e.EventRef = "refs/pull/9/merge" },
		"checkout ref": func(e *EvidenceFreshness) { e.CheckoutRef = commitFixtureSHA("other-head") },
		"swapped refs": func(e *EvidenceFreshness) { e.EventRef, e.CheckoutRef = e.CheckoutRef, e.EventRef },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := sealed
			mutate(&mutated.Evidence)
			if _, _, err := cache.GetProjectionIfFresh(key, identity, evidence, mutated); !errors.Is(err, ErrInvalidReceipt) && !errors.Is(err, ErrStale) {
				t.Fatalf("rejected ref mutation = %v, want ErrInvalidReceipt or ErrStale", err)
			}
			data, metadata, err := cache.Get(key)
			if err != nil || string(data) != "projection" || metadata.Size != int64(len(data)) {
				t.Fatalf("ref rejection changed projection: %q %+v %v", data, metadata, err)
			}
			receipts, err := cache.Receipts()
			if err != nil || len(receipts) != 1 {
				t.Fatalf("ref rejection changed receipt log: %d %v", len(receipts), err)
			}
		})
	}
}
func projectionHitFixture(t *testing.T) (*Cache, Key, ProjectionIdentity, EvidenceFreshness, CacheReceipt) {
	t.Helper()
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(key, []byte("projection")); err != nil {
		t.Fatal(err)
	}
	evidence := evidenceFixture("projection-hit")
	receipt := cacheReceiptFixture(key, evidence.RunID)
	identity := ProjectionIdentity{
		SemanticClosureDigest: key.SemanticClosureDigest, SourceDigest: evidence.SourceDigest,
		IRDigest: evidence.IRDigest, OptionsDigest: key.OptionsDigest,
		Toolchain: key.Toolchain, ToolchainDigest: evidence.ToolchainDigest,
	}
	return cache, key, identity, evidence, receipt
}
