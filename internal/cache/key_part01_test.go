package cache

import (
	"errors"
	"strings"
	"testing"
)

func makeTestKey(t *testing.T, version, namespace string) Key {
	t.Helper()
	key, err := NewKey(KeySpec{
		Version: version, Namespace: namespace, ToolVersion: "compiler-1",
		Inputs: map[string]any{"source": "main.gooo"}, OptionsDigest: mustOptionsDigest(map[string]any{"mode": "fast"}),
		Freshness: testFreshnessSpec(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return key
}
func testFreshnessSpec() FreshnessSpec {
	return FreshnessSpec{
		Dependencies: map[string]any{"revision": "known"},
		Provenance:   map[string]any{"evidence": "known"},
	}
}
func TestNewKeyIsCanonicalAndVersioned(t *testing.T) {
	first := makeTestKey(t, "v1", "billing")
	second := makeTestKey(t, "v1", "billing")
	if first != second {
		t.Fatal("equal key specs produced different keys")
	}
	changedVersion := makeTestKey(t, "v2", "billing")
	if first.String() == changedVersion.String() {
		t.Fatal("key version did not change the content address")
	}
	changedNamespace := makeTestKey(t, "v1", "payments")
	if first.String() == changedNamespace.String() {
		t.Fatal("namespace did not change the content address")
	}
	parsed, err := ParseKey(strings.ToUpper(first.String()))
	if err != nil || parsed.String() != first.String() {
		t.Fatalf("serialized key did not round-trip: %v %q", err, parsed)
	}
}
func TestNewKeyRejectsInvalidComponentsAndCanonicalValues(t *testing.T) {
	if _, err := NewKey(KeySpec{Namespace: ""}); err == nil {
		t.Fatal("empty namespace was accepted")
	}
	if _, err := NewKey(KeySpec{Namespace: "billing\x00"}); err == nil {
		t.Fatal("NUL namespace was accepted")
	}
	if _, err := NewKey(KeySpec{Namespace: "billing", Inputs: func() {}}); err == nil {
		t.Fatal("unsupported input value was accepted")
	}
}
func TestNewKeyRejectsOpaqueOptions(t *testing.T) {
	_, err := NewKey(KeySpec{
		Namespace: "billing", ToolVersion: "compiler-1", Inputs: map[string]any{"source": "main.gooo"},
		Options: map[string]any{"mode": "fast"}, Freshness: testFreshnessSpec(),
	})
	if !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("opaque options = %v, want ErrUnknownFreshness", err)
	}
}
