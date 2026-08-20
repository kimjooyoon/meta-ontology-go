package selectiveci

import (
	"bytes"
	"testing"
)

func TestBuildPermutationProperty(t *testing.T) {
	inputs := []SourceInput{
		testInput(t, "pkg/a.go", "A", "urn:gooo:entity:a"),
		testInput(t, "pkg/b.go", "B", "urn:gooo:entity:b"),
		testInput(t, "pkg/c.go", "C", "urn:gooo:entity:c"),
	}
	permutations := [][]int{
		{0, 1, 2}, {0, 2, 1}, {1, 0, 2},
		{1, 2, 0}, {2, 0, 1}, {2, 1, 0},
	}
	var wantDigest string
	var wantJSON []byte
	for index, permutation := range permutations {
		sources := make([]SourceInput, len(permutation))
		registered := make([]string, len(permutation))
		for position, sourceIndex := range permutation {
			sources[position] = inputs[sourceIndex]
			registered[position] = inputs[sourceIndex].Bindings[0].ID
		}
		snapshot, err := Build(SnapshotInput{
			Sources: sources, SourceMapDigest: testDigest("source-map"),
			RegistryDigest: testDigest("registry"), RegisteredIDs: registered,
		})
		if err != nil {
			t.Fatalf("permutation %d: Build: %v", index, err)
		}
		jsonBytes, err := snapshot.CanonicalJSON()
		if err != nil {
			t.Fatalf("permutation %d: CanonicalJSON: %v", index, err)
		}
		if index == 0 {
			wantDigest, wantJSON = snapshot.Digest, jsonBytes
			continue
		}
		if snapshot.Digest != wantDigest || !bytes.Equal(jsonBytes, wantJSON) {
			t.Fatalf("permutation %d changed canonical snapshot", index)
		}
	}
}
func TestDecodeSnapshotRejectsNonCanonicalOrUnknownJSON(t *testing.T) {
	snapshot := testSnapshot(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	canonical, err := snapshot.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	for name, data := range map[string][]byte{
		"whitespace":     append([]byte(" \n"), append(canonical, '\n')...),
		"unknown field":  bytes.Replace(canonical, []byte(`"digest"`), []byte(`"extra":true,"digest"`), 1),
		"trailing value": append(append([]byte(nil), canonical...), []byte("{}")...),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeSnapshot(data); err == nil {
				t.Fatal("accepted non-canonical snapshot")
			}
		})
	}
}
