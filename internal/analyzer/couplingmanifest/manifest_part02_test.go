package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"testing"
)

func TestBuildRepresentsAdditionsRemovalsAndPermutation(t *testing.T) {
	a := testSource(t, "pkg/a.go", "A", "urn:gooo:entity:a")
	b := testSource(t, "pkg/b.go", "B", "urn:gooo:entity:b")
	surfaces := []surfaceFixture{{Owner: a.Bindings[0].ID, Suffix: "a"}, {Owner: b.Bindings[0].ID, Suffix: "b"}}
	added, err := Build(testInput(t, []selectiveci.SourceInput{a}, []selectiveci.SourceInput{a, b}, surfaces))
	if err != nil {
		t.Fatalf("addition: %v", err)
	}
	if len(added.Entries) != 2 || added.Entries[1].BeforeBindingDigest != absentDigest || added.Entries[1].AfterBlobDigest == absentDigest || added.ZeroChange {
		t.Fatalf("addition manifest = %#v", added)
	}
	removed, err := Build(testInput(t, []selectiveci.SourceInput{a, b}, []selectiveci.SourceInput{a}, surfaces))
	if err != nil {
		t.Fatalf("removal: %v", err)
	}
	if removed.Entries[1].AfterBlobDigest != absentDigest || removed.Entries[1].AfterSourcePath != absentPath {
		t.Fatalf("removal manifest = %#v", removed)
	}

	firstInput := testInput(t, []selectiveci.SourceInput{a, b}, []selectiveci.SourceInput{a, b}, surfaces)
	secondInput := testInput(t, []selectiveci.SourceInput{b, a}, []selectiveci.SourceInput{b, a}, swap(surfaces))
	first, err := Build(firstInput)
	if err != nil {
		t.Fatalf("first permutation: %v", err)
	}
	second, err := Build(secondInput)
	if err != nil {
		t.Fatalf("second permutation: %v", err)
	}
	if first.Digest != second.Digest {
		t.Fatalf("permutation changed detector digest: %s/%s", first.Digest, second.Digest)
	}
}
