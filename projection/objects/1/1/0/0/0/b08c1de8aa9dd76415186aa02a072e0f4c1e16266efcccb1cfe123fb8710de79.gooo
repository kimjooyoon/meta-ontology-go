package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"reflect"
	"testing"
)

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
