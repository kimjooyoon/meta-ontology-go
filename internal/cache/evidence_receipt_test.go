package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
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
		"domain":     func(r *CacheReceipt) { r.Domain = "other" },
		"version":    func(r *CacheReceipt) { r.KeyVersion = "other" },
		"host stage": func(r *CacheReceipt) { r.HostStage = GoooHostedStage },
		"projection": func(r *CacheReceipt) { r.Projection = "other" },
		"artifact":   func(r *CacheReceipt) { r.ArtifactKind = "other" },
		"options":    func(r *CacheReceipt) { r.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"}) },
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

func TestCacheReceiptC1C4OptionsDigestFailsClosed(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "options")
	missing := receipt
	missing.OptionsDigest = ""
	if err := missing.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("missing options digest = %v, want ErrInvalidReceipt", err)
	}
	variant := projectionSpec()
	variant.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"})
	variantKey, err := NewProjectionKey(variant)
	if err != nil {
		t.Fatal(err)
	}
	if variantKey == key {
		t.Fatal("options mutation retained projection identity")
	}
	receipt.OptionsDigest = variantKey.OptionsDigest
	if err := receipt.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("mismatched options digest = %v, want ErrInvalidReceipt", err)
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
	assertInvalidSeal(t, "missing bundle", missing)
	missingEvent := receipt
	missingEvent.Evidence.EventID = ""
	assertInvalidSeal(t, "missing event", missingEvent)
	missingAttempt := receipt
	missingAttempt.Evidence.Attempt = 0
	assertInvalidSeal(t, "missing attempt", missingAttempt)
	missingSHA := receipt
	missingSHA.Evidence.HeadSHA = ""
	assertInvalidSeal(t, "missing head SHA", missingSHA)
	missingJobs := receipt
	missingJobs.Evidence.Jobs = nil
	assertInvalidSeal(t, "missing canonical jobs", missingJobs)
	zero := receipt
	zero.Evidence.BundleDigest = Digest("0000000000000000000000000000000000000000000000000000000000000000")
	assertInvalidSeal(t, "zero bundle", zero)
	unknownRef := receipt
	unknownRef.EvidenceRefs = nil
	unknownRef.Evidence.EvidenceRefs = nil
	assertInvalidSeal(t, "missing evidence ref", unknownRef)
	unboundRef := receipt
	unboundRef.EvidenceRefs = append([]EvidenceRef(nil), receipt.EvidenceRefs...)
	unboundRef.EvidenceRefs = append(unboundRef.EvidenceRefs,
		EvidenceRef{Name: "unbound", Digest: HashBytes([]byte("unbound"))})
	unboundRef.Evidence.EvidenceRefs = append([]EvidenceRef(nil), unboundRef.EvidenceRefs...)
	assertInvalidSeal(t, "unbound evidence ref", unboundRef)
	zeroRef := receipt
	zeroRef.EvidenceRefs = []EvidenceRef{{Name: "source", Digest: Digest("0000000000000000000000000000000000000000000000000000000000000000")}}
	zeroRef.Evidence.EvidenceRefs = zeroRef.EvidenceRefs
	assertInvalidSeal(t, "zero evidence ref", zeroRef)
}

func assertInvalidSeal(t *testing.T, name string, receipt CacheReceipt) {
	t.Helper()
	if _, err := receipt.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("%s = %v, want ErrInvalidReceipt", name, err)
	}
}

func TestReceiptLogRejectsSymlinkWithoutMutatingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires platform privileges on Windows")
	}
	cache, _, _, _, receipt := projectionHitFixture(t)
	target := filepath.Join(t.TempDir(), "outside.jsonl")
	original := []byte("outside\n")
	if err := os.WriteFile(target, original, 0o600); err != nil {
		t.Fatal(err)
	}
	receiptLog := filepath.Join(cache.Root(), receiptsFileName)
	if err := os.Symlink(target, receiptLog); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("append through receipt symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	if _, err := cache.Receipts(); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("read through receipt symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("symlink target mutated: %q", got)
	}
}

func TestReceiptAppendSerializesAcrossCacheHandles(t *testing.T) {
	cache, _, _, _, receipt := projectionHitFixture(t)
	second, err := Open(cache.Root())
	if err != nil {
		t.Fatal(err)
	}
	release, err := acquireReceiptFileLock(cache.receipts)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		_, err := second.AppendReceipt(receipt)
		done <- err
	}()
	<-started
	select {
	case err := <-done:
		release()
		t.Fatalf("append completed while receipt lock was held: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("serialized append = %v", err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("cross-handle replay = %v, want ErrReceiptReplay", err)
	}
	if records, err := second.Receipts(); err != nil || len(records) != 1 {
		t.Fatalf("receipt records = %d, %v; want one", len(records), err)
	}
}

func TestReceiptLockRejectsSymlinkWithoutMutatingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires platform privileges on Windows")
	}
	cache, _, _, _, receipt := projectionHitFixture(t)
	target := filepath.Join(t.TempDir(), "outside.lock")
	original := []byte("outside lock\n")
	if err := os.WriteFile(target, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, cache.receipts+".lock"); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("append through lock symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("lock symlink target mutated: %q", got)
	}
}
