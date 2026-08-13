package cache

import (
	"context"
	"errors"
	"testing"
)

func partialSpec(part string, value, revision int) PartialSpec {
	return PartialSpec{
		Part: part,
		KeySpec: KeySpec{
			Version: "v1", Namespace: "billing", ToolVersion: "compiler-1",
			Inputs: map[string]any{"value": value}, OptionsDigest: mustOptionsDigest(map[string]any{"mode": "fast"}),
			Freshness: FreshnessSpec{
				Dependencies: map[string]any{"revision": revision},
				Provenance:   map[string]any{"source": "main.gooo"},
			},
		},
	}
}

func TestPartialKeyReusesOnlyUnchangedFragment(t *testing.T) {
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	header := partialSpec("header", 1, 1)
	body := partialSpec("body", 1, 1)
	for _, spec := range []PartialSpec{header, body} {
		if _, err := NewPartialKey(spec); err != nil {
			t.Fatal(err)
		}
		if err := cache.PutPartial(spec, []byte(spec.Part)); err != nil {
			t.Fatal(err)
		}
	}
	changedHeader := partialSpec("header", 2, 2)
	unchangedBody := partialSpec("body", 1, 1)
	oldBodyKey, err := NewPartialKey(body)
	if err != nil {
		t.Fatal(err)
	}
	newBodyKey, err := NewPartialKey(unchangedBody)
	if err != nil {
		t.Fatal(err)
	}
	if oldBodyKey != newBodyKey {
		t.Fatal("unchanged fragment did not retain its key")
	}
	var bodyComputes int
	_, data, metadata, hit, err := cache.GetOrComputePartial(context.Background(), unchangedBody,
		func() ([]byte, error) {
			bodyComputes++
			return []byte("rebuilt body"), nil
		})
	if err != nil || !hit || string(data) != "body" || bodyComputes != 0 {
		t.Fatalf("body reuse = %q, hit=%v, computes=%d, err=%v", data, hit, bodyComputes, err)
	}
	if metadata.Key != newBodyKey.String() {
		t.Fatalf("body metadata key = %q, want %q", metadata.Key, newBodyKey)
	}
	var headerComputes int
	newHeaderKey, data, _, hit, err := cache.GetOrComputePartial(context.Background(), changedHeader,
		func() ([]byte, error) {
			headerComputes++
			return []byte("new header"), nil
		})
	if err != nil || hit || string(data) != "new header" || headerComputes != 1 {
		t.Fatalf("header recompute = %q, hit=%v, computes=%d, err=%v", data, hit, headerComputes, err)
	}
	if newHeaderKey == oldBodyKey {
		t.Fatal("changed fragment collided with body key")
	}
}

func TestPartialFreshnessSeparatesProvenance(t *testing.T) {
	first := partialSpec("body", 1, 1)
	second := partialSpec("body", 1, 2)
	firstKey, err := NewPartialKey(first)
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := NewPartialKey(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstKey == secondKey {
		t.Fatal("provenance change did not change partial key")
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(firstKey, []byte("old")); err != nil {
		t.Fatal(err)
	}
	current, err := NewFreshness(second.KeySpec.Freshness)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.GetFresh(firstKey, current); !errors.Is(err, ErrStale) {
		t.Fatalf("stale partial read = %v, want ErrStale", err)
	}
}

func TestNewPartialKeyRejectsEmptyPart(t *testing.T) {
	if _, err := NewPartialKey(PartialSpec{KeySpec: KeySpec{Namespace: "billing"}}); err == nil {
		t.Fatal("empty partial name was accepted")
	}
}
