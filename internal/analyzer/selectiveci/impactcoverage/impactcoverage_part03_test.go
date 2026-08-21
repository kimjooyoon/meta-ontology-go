package impactcoverage

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"testing"
)

func assertDigests(t *testing.T, base, head selectiveci.Snapshot, got Result) {
	t.Helper()
	input := NewInput(&base, &head)
	if got.InputDigest != input.Digest() || got.OutputDigest != got.StableDigest() {
		t.Fatalf("digest vector = input %q/%q output %q/%q", got.InputDigest,
			input.Digest(), got.OutputDigest, got.StableDigest())
	}
	if got.BaseSnapshotDigest != base.Digest || got.HeadSnapshotDigest != head.Digest ||
		got.BaseSourceMapDigest != base.SourceMapDigest || got.HeadSourceMapDigest != head.SourceMapDigest ||
		got.BaseRegistryDigest != base.RegistryDigest || got.HeadRegistryDigest != head.RegistryDigest {
		t.Fatalf("authority digest vector = %#v", got)
	}
}
func TestInvalidAndDuplicateSnapshotRecordsFailClosed(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	stale := base
	stale.Digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	got := Observe(NewInput(&base, &stale))
	if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot ||
		!got.FullSuiteRequired || len(got.ChangedStableIDs) != 0 {
		t.Fatalf("stale snapshot result = %#v", got)
	}
	duplicate := base
	duplicate.Sources = append(append([]selectiveci.Source{}, base.Sources...), base.Sources[0])
	got = Observe(NewInput(&base, &duplicate))
	if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot || len(got.ChangedStableIDs) != 0 {
		t.Fatalf("duplicate snapshot result = %#v", got)
	}
}
func TestRootLocationAndOrderingAreInvariant(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("./pkg/a.go", "a-1", "urn:gooo:entity:a"))
	head := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	got := Observe(NewInput(&base, &head))
	if got.Decision != DecisionExact || got.Reason != ReasonNoChange || got.ChangedBlobCount != 0 {
		t.Fatalf("root-location result = %#v", got)
	}
	baseJSON, err := base.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	headJSON, err := head.CanonicalJSON()
	if err != nil || !bytes.Equal(baseJSON, headJSON) {
		t.Fatalf("root-location canonical bytes differ: %s/%s/%v", baseJSON, headJSON, err)
	}
}
func TestNilAuthorityAndOverflowFailClosed(t *testing.T) {
	valid := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	for name, input := range map[string]Input{
		"nil-base": NewInput(nil, &valid), "nil-head": NewInput(&valid, nil),
	} {
		t.Run(name, func(t *testing.T) {
			got := Observe(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot || !got.FullSuiteRequired {
				t.Fatalf("nil authority result = %#v", got)
			}
		})
	}
	if _, err := checkedAdd(^uint64(0), 1); err == nil {
		t.Fatal("work-unit overflow was accepted")
	}
	if units, err := checkedAdd(4, 5); err != nil || units != 9 {
		t.Fatalf("checked work units = %d/%v", units, err)
	}
}
