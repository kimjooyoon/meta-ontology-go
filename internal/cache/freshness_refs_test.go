package cache

import (
	"errors"
	"testing"
)

func TestCacheReceiptRefsBindEqualityAndSeal(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	left := cacheReceiptFixture(key, "refs-canonical")
	right := cacheReceiptFixture(key, "refs-canonical")
	right.Evidence.EventRef = "refs/pull/9/merge"
	if left.Evidence.Equal(right.Evidence) {
		t.Fatal("event ref mutation retained freshness equality")
	}
	sealedLeft, err := left.Seal()
	if err != nil {
		t.Fatal(err)
	}
	sealedRight, err := right.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if sealedLeft.ReceiptDigest == sealedRight.ReceiptDigest {
		t.Fatal("event ref mutation retained receipt identity")
	}
	invalidCheckout := left
	invalidCheckout.Evidence.CheckoutRef = commitFixtureSHA("other-head")
	assertInvalidSeal(t, "head-mismatched checkout ref", invalidCheckout)
}

func TestCacheReceiptRefsRejectReplayAndInvalidValues(t *testing.T) {
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
	eventReplay := cacheReceiptFixture(key, "run-2")
	eventReplay.Evidence.EventID = "event-replay-id"
	eventReplay.Evidence.EventRef = sealed.Evidence.EventRef
	assertReplay(t, cache, eventReplay, "event ref/attempt")
	idReplay := cacheReceiptFixture(key, "run-3")
	idReplay.Evidence.EventID = sealed.Evidence.EventID
	idReplay.Evidence.EventRef = "refs/pull/9/head"
	assertReplay(t, cache, idReplay, "event ID/attempt")
	for name, mutate := range map[string]func(*CacheReceipt){
		"missing event ref":    func(r *CacheReceipt) { r.Evidence.EventRef = "" },
		"malformed event ref":  func(r *CacheReceipt) { r.Evidence.EventRef = "push:event" },
		"missing checkout ref": func(r *CacheReceipt) { r.Evidence.CheckoutRef = "" },
		"swapped refs": func(r *CacheReceipt) {
			r.Evidence.EventRef, r.Evidence.CheckoutRef = r.Evidence.CheckoutRef, r.Evidence.EventRef
		},
		"mismatched checkout": func(r *CacheReceipt) { r.Evidence.CheckoutRef = commitFixtureSHA("other-head") },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := receipt
			mutate(&mutated)
			assertInvalidSeal(t, name, mutated)
		})
	}
}

func assertReplay(t *testing.T, cache *Cache, receipt CacheReceipt, label string) {
	t.Helper()
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("%s replay = %v, want ErrReceiptReplay", label, err)
	}
}
