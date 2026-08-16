package couplingmanifest

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

var (
	_ Manifest                     = detector.ChangeManifest{}
	_ Result                       = detector.Result{}
	_ func([]byte) (Result, error) = DecodeResult
)

func TestBuildUsesDetectorAuthorityAndPassThroughCodecs(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	input := testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source}, []surfaceFixture{{Owner: source.Bindings[0].ID, Suffix: "order"}})
	output, err := BuildDetailed(input)
	if err != nil {
		t.Fatalf("BuildDetailed: %v", err)
	}
	if !output.Manifest.Complete || !output.Manifest.ZeroChange || output.Manifest.Schema != detector.ManifestSchemaV1 {
		t.Fatalf("manifest = %#v", output.Manifest)
	}
	if output.DetectorResult.Status != detector.StatusUnknown || output.DetectorResult.Reasons[0].Code != detector.ReasonExternalReceiptMissing {
		t.Fatalf("structural detector result = %#v", output.DetectorResult)
	}
	if output.Metadata.Status != ConstructionComplete || output.Metadata.Counts != (ComponentCounts{Registered: 1, Before: 1, Head: 1, Resolved: 1}) {
		t.Fatalf("metadata = %#v", output.Metadata)
	}

	packet := detectorInput(output.Manifest, input.Authority)
	wantBytes, err := detector.EncodeInput(packet)
	if err != nil {
		t.Fatalf("detector EncodeInput: %v", err)
	}
	gotBytes, err := EncodeInput(packet)
	if err != nil || !bytes.Equal(gotBytes, wantBytes) {
		t.Fatalf("adapter changed detector input bytes: %v", err)
	}
	decoded, err := DecodeJSON(wantBytes)
	if err != nil {
		t.Fatalf("adapter DecodeJSON: %v", err)
	}
	replayed, err := detector.EncodeInput(decoded)
	if err != nil || !bytes.Equal(replayed, wantBytes) {
		t.Fatalf("detector input replay changed bytes: %v", err)
	}
	resultBytes, err := detector.EncodeResult(output.DetectorResult)
	if err != nil {
		t.Fatalf("detector EncodeResult: %v", err)
	}
	decodedResult, err := DecodeResult(resultBytes)
	if err != nil {
		t.Fatalf("adapter DecodeResult: %v", err)
	}
	replayedResult, err := EncodeResult(decodedResult)
	if err != nil || !bytes.Equal(replayedResult, resultBytes) {
		t.Fatalf("detector result replay changed bytes: %v", err)
	}
}

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

func TestBuildRejectsForgedMissingAndStaleDetectorAuthority(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	base := testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source}, []surfaceFixture{{Owner: source.Bindings[0].ID, Suffix: "order"}})
	cases := []struct {
		name   string
		mutate func(*Input)
	}{
		{name: "missing authority schema", mutate: func(input *Input) { input.Authority.Schema = "" }},
		{name: "stale authority snapshot", mutate: func(input *Input) { input.Authority.SnapshotDigest = testDigest("stale") }},
		{name: "forged source-map binding", mutate: func(input *Input) { input.SourceMap.Head[0].SourceMapBindingDigest = testDigest("forged") }},
		{name: "candidate binding", mutate: func(input *Input) {
			input.SourceMap.CandidateBindings = append([]SourceMapObservation{}, input.SourceMap.Head...)
		}},
		{name: "derived binding", mutate: func(input *Input) {
			input.SourceMap.DerivedBindings = append([]SourceMapObservation{}, input.SourceMap.Head...)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := cloneInput(base)
			tc.mutate(&input)
			output, err := BuildDetailed(input)
			if err == nil || output.Manifest.Complete || len(output.Manifest.Entries) != 0 {
				t.Fatalf("forged input was accepted or repaired: output=%#v err=%v", output, err)
			}
		})
	}

	valid, err := Build(base)
	if err != nil {
		t.Fatalf("valid base: %v", err)
	}
	forgedManifest := valid
	forgedManifest.RegistryDigest = rawTestDigest(testDigest("forged-registry"))
	forgedManifest.Digest = detectorManifestDigest(forgedManifest)
	result := ValidateManifest(forgedManifest, base.Authority)
	if result.Status == detector.StatusPass || result.Reasons[0].Code != detector.ReasonDigestMismatch {
		t.Fatalf("forged manifest was laundered: %#v", result)
	}
	if valid.RegistryDigest == forgedManifest.RegistryDigest {
		t.Fatal("authority mutation was overwritten")
	}
}

func TestDetectorResultRejectsStaleBindingWithoutAdapterRepair(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	input := testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source}, []surfaceFixture{{Owner: source.Bindings[0].ID, Suffix: "order"}})
	manifest, err := Build(input)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	stale := manifest
	stale.Entries = append([]detector.ManifestEntry(nil), manifest.Entries...)
	stale.Entries[0].AfterBindingDigest = testDigest("stale-binding")
	stale.Digest = detectorManifestDigest(stale)
	before := stale
	result := ValidateManifest(stale, input.Authority)
	if result.Status == detector.StatusPass || len(result.Reasons) == 0 {
		t.Fatalf("stale detector binding accepted: %#v", result)
	}
	if !reflect.DeepEqual(stale, before) {
		t.Fatal("detector validation repaired or mutated the manifest")
	}
}

func cloneInput(input Input) Input {
	copy := input
	copy.Authority.Registry.Surfaces = append([]detector.Surface(nil), input.Authority.Registry.Surfaces...)
	copy.SourceMap.Before = append([]SourceMapObservation(nil), input.SourceMap.Before...)
	copy.SourceMap.Head = append([]SourceMapObservation(nil), input.SourceMap.Head...)
	copy.SourceMap.CandidateBindings = append([]SourceMapObservation(nil), input.SourceMap.CandidateBindings...)
	copy.SourceMap.DerivedBindings = append([]SourceMapObservation(nil), input.SourceMap.DerivedBindings...)
	return copy
}
