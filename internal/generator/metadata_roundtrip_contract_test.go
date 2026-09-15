package generator

import (
	"bytes"
	"reflect"
	"testing"
)

func TestProjectionMetadataV1ReplayBindsBytesAndIdentityDeterministically(t *testing.T) {
	ir := acceptanceFixture()
	first, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := bytes.Replace(first.Source, []byte("return Artifact{}"), []byte("return Artifact{}\n\t// metadata replay preserved"), 1)
	if bytes.Equal(previous, first.Source) {
		t.Fatal("fixture did not produce a protected handwritten replay input")
	}
	beforeIR := copyIR(ir)
	beforePrevious := append([]byte(nil), previous...)

	replayed, err := GenerateProjectionV1(ir, previous)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := GenerateProjectionV1(ir, previous)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(replayed.Source, repeated.Source) || !reflect.DeepEqual(replayed.SourceMap, repeated.SourceMap) {
		t.Fatal("replayed projection bytes or source map are not deterministic")
	}
	if replayed.Metadata.SemanticIRDigest != first.Metadata.SemanticIRDigest {
		t.Fatalf("replay changed normalized SemanticIR digest: first=%q replay=%q", first.Metadata.SemanticIRDigest, replayed.Metadata.SemanticIRDigest)
	}
	for _, result := range []ProjectionMetadataV1{replayed, repeated} {
		if _, err := result.CanonicalJSON(); err != nil {
			t.Fatalf("replayed projection failed canonical digest binding: %v", err)
		}
	}
	firstHash, err := replayed.CanonicalHash()
	if err != nil {
		t.Fatal(err)
	}
	repeatedHash, err := repeated.CanonicalHash()
	if err != nil || firstHash != repeatedHash {
		t.Fatalf("replayed canonical hash is not deterministic: first=%q repeated=%q err=%v", firstHash, repeatedHash, err)
	}
	for _, id := range []string{"gooo://activity/compile", "gooo://slot/compile-implementation"} {
		left, right := replayed.SourceMap.Lookup(id), repeated.SourceMap.Lookup(id)
		if len(left) != 1 || len(right) != 1 || left[0] != right[0] {
			t.Fatalf("replayed source-map identity %q changed: %#v %#v", id, left, right)
		}
	}
	if !bytes.Contains(replayed.Source, []byte("// metadata replay preserved")) {
		t.Fatal("replay did not preserve handwritten slot bytes")
	}
	if !reflect.DeepEqual(ir, beforeIR) || !bytes.Equal(previous, beforePrevious) {
		t.Fatal("replay mutated caller-owned IR or previous source")
	}
}
