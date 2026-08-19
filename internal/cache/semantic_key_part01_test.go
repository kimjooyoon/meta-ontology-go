package cache

import (
	"errors"
	"strings"
	"testing"
)

func TestSemanticProjectionKeyUsesStableIRIdentity(t *testing.T) {
	original := loadBillingIR(t, func(source string) string { return source })
	renamed := loadBillingIR(t, func(source string) string {
		source = strings.Replace(source, "entity Order id", "entity Purchase id", 1)
		return strings.Replace(source, "PayOrder(Order,", "PayOrder(Purchase,", 1)
	})
	changedID := loadBillingIR(t, func(source string) string {
		return strings.Replace(source, `billing://entity/order`, `billing://entity/purchase`, 1)
	})

	originalDigest, err := SemanticDigest(original)
	if err != nil {
		t.Fatal(err)
	}
	renamedDigest, err := SemanticDigest(renamed)
	if err != nil {
		t.Fatal(err)
	}
	changedDigest, err := SemanticDigest(changedID)
	if err != nil {
		t.Fatal(err)
	}
	if originalDigest != renamedDigest {
		t.Fatalf("display rename changed semantic digest: %s != %s", originalDigest, renamedDigest)
	}
	if originalDigest == changedDigest {
		t.Fatal("stable ID mutation retained semantic digest")
	}

	spec := projectionSpec()
	spec.Domain = ""
	spec.Namespace = ""
	spec.SemanticClosureDigest = ""
	originalKey, err := NewSemanticProjectionKey(original, spec)
	if err != nil {
		t.Fatal(err)
	}
	renamedKey, err := NewSemanticProjectionKey(renamed, spec)
	if err != nil {
		t.Fatal(err)
	}
	changedKey, err := NewSemanticProjectionKey(changedID, spec)
	if err != nil {
		t.Fatal(err)
	}
	if originalKey != renamedKey {
		t.Fatal("presentation-only rename changed semantic projection key")
	}
	if originalKey == changedKey {
		t.Fatal("stable ID mutation retained semantic projection key")
	}
	if originalKey.Domain != "billing" || originalKey.Namespace != "billing" {
		t.Fatalf("semantic namespace was not bound to key: %+v", originalKey)
	}

	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(originalKey, []byte("generated billing projection")); err != nil {
		t.Fatal(err)
	}
	data, _, hit, err := cache.GetOrCompute(t.Context(), renamedKey, func() ([]byte, error) {
		t.Fatal("presentation-only rename recomputed a semantic projection")
		return nil, nil
	})
	if err != nil || !hit || string(data) != "generated billing projection" {
		t.Fatalf("rename cache lookup = %q, hit=%v, err=%v", data, hit, err)
	}
	if _, _, err := cache.Get(changedKey); !errors.Is(err, ErrNotFound) {
		t.Fatalf("stable ID mutation lookup = %v, want ErrNotFound", err)
	}
}
