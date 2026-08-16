package selectiveci

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

func TestBuildAndCanonicalJSONAreOrderIndependent(t *testing.T) {
	first := testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	second := testInput(t, "pkg/customer.go", "Customer", "urn:gooo:entity:customer")

	left, err := Build(SnapshotInput{
		Sources:         []SourceInput{first, second},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{first.Bindings[0].ID, second.Bindings[0].ID},
	})
	if err != nil {
		t.Fatalf("Build left: %v", err)
	}
	right, err := Build(SnapshotInput{
		Sources:         []SourceInput{second, first},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{second.Bindings[0].ID, first.Bindings[0].ID},
	})
	if err != nil {
		t.Fatalf("Build right: %v", err)
	}
	leftJSON, err := left.CanonicalJSON()
	if err != nil {
		t.Fatalf("left canonical JSON: %v", err)
	}
	rightJSON, err := right.CanonicalJSON()
	if err != nil {
		t.Fatalf("right canonical JSON: %v", err)
	}
	if left.Digest != right.Digest || !bytes.Equal(leftJSON, rightJSON) {
		t.Fatalf("input permutation changed snapshot: %s/%s", left.Digest, right.Digest)
	}
	decoded, err := DecodeSnapshot(leftJSON)
	if err != nil {
		t.Fatalf("DecodeSnapshot: %v", err)
	}
	if decoded.Digest != left.Digest {
		t.Fatalf("decoded digest = %q, want %q", decoded.Digest, left.Digest)
	}
	var roundTrip Snapshot
	if err := json.Unmarshal(leftJSON, &roundTrip); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if roundTrip.Digest != left.Digest {
		t.Fatalf("round-trip digest = %q, want %q", roundTrip.Digest, left.Digest)
	}
}

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

func TestDiffChangedDeletedRelocatedAndUnchanged(t *testing.T) {
	base := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	unchanged := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	renamed := testSnapshot(t, "pkg/new.go", "RenamedOrder", "urn:gooo:entity:order")
	deleted, err := Build(SnapshotInput{
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{"urn:gooo:entity:order"},
	})
	if err != nil {
		t.Fatalf("empty head snapshot: %v", err)
	}

	cases := []struct {
		name string
		head Snapshot
		want []string
	}{
		{name: "identical exact binding", head: unchanged, want: []string{}},
		{name: "rename", head: renamed, want: []string{"urn:gooo:entity:order"}},
		{name: "deletion", head: deleted, want: []string{"urn:gooo:entity:order"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Diff(base, tc.head)
			if err != nil {
				t.Fatalf("Diff: %v", err)
			}
			if got.Status != StatusBound || got.FullSuiteFallback || !reflect.DeepEqual(got.ChangedIDs, tc.want) {
				t.Fatalf("delta = %#v, want IDs %#v", got, tc.want)
			}
		})
	}
}

func TestDiffRelocationIncludesStableID(t *testing.T) {
	base := testSnapshot(t, "pkg/old.go", "Order", "urn:gooo:entity:order")
	relocated := testSnapshot(t, "internal/order.go", "Order", "urn:gooo:entity:order")
	got, err := Diff(base, relocated)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !reflect.DeepEqual(got.ChangedIDs, []string{"urn:gooo:entity:order"}) {
		t.Fatalf("relocation IDs = %#v", got.ChangedIDs)
	}
}

func TestBuildRejectsUncertainAuthorityWithoutPartialIDs(t *testing.T) {
	valid := testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")

	cases := []struct {
		name string
		edit func(*SnapshotInput)
		code ErrorCode
	}{
		{name: "missing binding", edit: func(input *SnapshotInput) {
			input.Sources[0].Bindings = nil
		}, code: CodeMissingBinding},
		{name: "ambiguous source attachment", edit: func(input *SnapshotInput) {
			input.Sources[0].Path = "pkg/other.go"
		}, code: CodeAmbiguousBinding},
		{name: "duplicate binding", edit: func(input *SnapshotInput) {
			input.Sources[0].Bindings = append(append([]semanticbinding.Binding(nil), valid.Bindings...), valid.Bindings[0])
		}, code: CodeDuplicateBinding},
		{name: "unregistered ID", edit: func(input *SnapshotInput) {
			input.RegisteredIDs = []string{"urn:gooo:entity:other"}
		}, code: CodeUnregisteredID},
		{name: "malformed path", edit: func(input *SnapshotInput) {
			input.Sources[0].Path = "../outside.go"
		}, code: CodeMalformedPath},
		{name: "malformed digest", edit: func(input *SnapshotInput) {
			input.Sources[0].BlobDigest = "not-a-digest"
		}, code: CodeMalformedDigest},
		{name: "candidate-only identity", edit: func(input *SnapshotInput) {
			input.CandidateBindings = valid.Bindings
		}, code: CodeCandidateIdentity},
		{name: "derived-only identity", edit: func(input *SnapshotInput) {
			input.DerivedBindings = valid.Bindings
		}, code: CodeDerivedIdentity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := SnapshotInput{
				Sources:         []SourceInput{testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")},
				SourceMapDigest: testDigest("source-map"),
				RegistryDigest:  testDigest("registry"),
				RegisteredIDs:   []string{"urn:gooo:entity:order"},
			}
			tc.edit(&input)
			got, err := Build(input)
			if err == nil {
				t.Fatal("Build accepted uncertain authority")
			}
			var typed *Error
			if !errors.As(err, &typed) || typed.Code != tc.code {
				t.Fatalf("error = %v, want code %q", err, tc.code)
			}
			if got.Status != StatusUnknown || !got.FullSuiteFallback || len(got.Sources) != 0 || got.Digest != "" {
				t.Fatalf("unknown snapshot retained partial authority: %#v", got)
			}
		})
	}
}

func TestDiffStaleSnapshotFallsBackWithoutPartialIDs(t *testing.T) {
	base := testSnapshot(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	head := testSnapshot(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	head.Sources[0].BlobDigest = testDigest("tampered")
	got, err := Diff(base, head)
	if err == nil {
		t.Fatal("Diff accepted stale head snapshot")
	}
	if got.Status != StatusUnknown || !got.FullSuiteFallback || len(got.ChangedIDs) != 0 {
		t.Fatalf("stale delta = %#v, want UNKNOWN/full-suite/no IDs", got)
	}
}

func testSnapshot(t *testing.T, path, name, id string) Snapshot {
	t.Helper()
	input := testInput(t, path, name, id)
	result, err := Build(SnapshotInput{
		Sources:         []SourceInput{input},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{id},
	})
	if err != nil {
		t.Fatalf("Build %s: %v", path, err)
	}
	return result
}

func testInput(t *testing.T, path, name, id string) SourceInput {
	t.Helper()
	source := []byte(fmt.Sprintf("package fixture\n\n//gooo:bind id=%q role=\"HANDWRITTEN_IMPL\"\nfunc %s() {}\n", id, name))
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{
		Filename: path, PackagePath: "fixture", Source: source,
	}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semanticbinding.Extract = %#v, err=%v", result, err)
	}
	return SourceInput{Path: path, BlobDigest: testDigest(string(source)), Bindings: result.Bindings}
}

func testDigest(value string) string { return digest([]byte(value)) }
