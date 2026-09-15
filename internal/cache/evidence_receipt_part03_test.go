package cache

import (
	"errors"
	"testing"
)

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
