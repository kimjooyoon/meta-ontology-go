package cache

import (
	"errors"
	"testing"
)

func TestGetProjectionIfFreshRequiresRecordedEvidence(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	if _, _, err := cache.GetProjectionIfFresh(key, identity, evidence, receipt); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("unrecorded receipt = %v, want ErrInvalidReceipt", err)
	}
	before, beforeMetadata, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != "projection" || beforeMetadata.Size != int64(len(before)) {
		t.Fatalf("projection changed after rejected receipt: %q %+v", before, beforeMetadata)
	}
}

func TestGetProjectionIfFreshRejectsMutationsWithoutWrites(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	got, metadata, err := cache.GetProjectionIfFresh(key, identity, evidence, sealed)
	if err != nil || string(got) != "projection" || metadata.Size != int64(len(got)) {
		t.Fatalf("fresh projection = %q %+v %v", got, metadata, err)
	}
	receiptCount := func() int {
		receipts, err := cache.Receipts()
		if err != nil {
			t.Fatal(err)
		}
		return len(receipts)
	}
	for name, mutate := range map[string]func(*ProjectionIdentity, *EvidenceFreshness){
		"source": func(i *ProjectionIdentity, e *EvidenceFreshness) {
			i.SourceDigest = HashBytes([]byte("changed-source"))
		},
		"IR": func(i *ProjectionIdentity, e *EvidenceFreshness) {
			i.IRDigest = HashBytes([]byte("changed-ir"))
		},
		"options": func(i *ProjectionIdentity, e *EvidenceFreshness) {
			i.OptionsDigest = HashBytes([]byte("changed-options"))
		},
		"stale tuple": func(i *ProjectionIdentity, e *EvidenceFreshness) {
			e.HeadDigest = HashBytes([]byte("stale-head"))
		},
		"stale predecessor": func(i *ProjectionIdentity, e *EvidenceFreshness) {
			e.PredecessorDigests[0] = HashBytes([]byte("stale-predecessor"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			mutatedIdentity, mutatedEvidence := identity, canonicalEvidence(evidence)
			mutate(&mutatedIdentity, &mutatedEvidence)
			if _, _, err := cache.GetProjectionIfFresh(key, mutatedIdentity, mutatedEvidence, sealed); !errors.Is(err, ErrStale) {
				t.Fatalf("rejected mutation = %v, want ErrStale", err)
			}
			if receiptCount() != 1 {
				t.Fatal("rejection mutated append-only receipt log")
			}
			data, currentMetadata, err := cache.Get(key)
			if err != nil || string(data) != "projection" || currentMetadata != metadata {
				t.Fatalf("rejection mutated projection: %q %+v %v", data, currentMetadata, err)
			}
		})
	}
}

func TestGetProjectionIfFreshRejectsReplayedReceipt(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("replayed receipt append = %v, want ErrReceiptReplay", err)
	}
	if _, _, err := cache.GetProjectionIfFresh(key, identity, evidence, sealed); err != nil {
		t.Fatalf("single recorded receipt should remain usable: %v", err)
	}
}

func TestGetProjectionIfFreshRejectsUnknownFreshness(t *testing.T) {
	cache, key, identity, evidence, receipt := projectionHitFixture(t)
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	for name, invoke := range map[string]func() error{
		"unknown identity": func() error {
			unknown := identity
			unknown.SourceDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, unknown, evidence, sealed)
			return err
		},
		"unknown current": func() error {
			unknown := evidence
			unknown.SourceDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, identity, unknown, sealed)
			return err
		},
		"unknown receipt": func() error {
			unknown := sealed
			unknown.Evidence.PolicyDigest = ""
			_, _, err := cache.GetProjectionIfFresh(key, identity, evidence, unknown)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := invoke()
			if !errors.Is(err, ErrUnknownFreshness) && !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("unknown freshness = %v", err)
			}
			data, metadata, err := cache.Get(key)
			if err != nil || string(data) != "projection" || metadata.Size != int64(len(data)) {
				t.Fatalf("unknown freshness changed projection: %q %+v %v", data, metadata, err)
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
