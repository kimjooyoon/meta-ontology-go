package impactgraph_test

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeRejectsMalformedGraphs(t *testing.T) {
	for _, fixture := range []string{
		"duplicate-node.json",
		"duplicate-edge.json",
		"illegal-endpoint-kinds.json",
		"duplicate-json-field.json",
		"unknown-json-field.json",
		"trailing-json.json",
	} {
		t.Run(fixture, func(t *testing.T) {
			if _, err := impactgraph.Decode(fixtureBytes(t, fixture)); err == nil {
				t.Fatalf("Decode(%q) accepted malformed input", fixture)
			}
		})
	}
}
func TestCanonicalAndDigestReplayIgnoreInsertionOrder(t *testing.T) {
	first := decodeFixture(t, "positive-3of3.json")
	second := decodeFixture(t, "positive-3of3-reordered.json")

	firstCanonical := []byte(first.Canonical())
	secondCanonical := []byte(second.Canonical())
	if !bytes.Equal(firstCanonical, secondCanonical) {
		t.Fatalf("insertion order changed canonical bytes:\n%s\n---\n%s", firstCanonical, secondCanonical)
	}

	firstDigest := first.Digest()
	secondDigest := second.Digest()
	if firstDigest == "" || firstDigest != secondDigest {
		t.Fatalf("insertion order changed digest: %q vs %q", firstDigest, secondDigest)
	}
}
func decodeFixture(t *testing.T, name string) impactgraph.Graph {
	t.Helper()
	graph, err := impactgraph.Decode(fixtureBytes(t, name))
	if err != nil {
		t.Fatalf("Decode(%q): %v", name, err)
	}
	return graph
}
func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return data
}
