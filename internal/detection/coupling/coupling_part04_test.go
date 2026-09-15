package coupling

import (
	"testing"
)

func TestLocalizationBenefitIsFiniteAndComponentwise(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	second := fixture.input.Registry.Surfaces[0]
	second.SurfaceID, second.CodeSymbolID, second.SemanticOwnerID = fixtureID("surface-two"), fixtureID("code-two"), fixtureID("owner-two")
	second.Binding.SourceMapID = fixtureID("source-map-two")
	second.Binding.BindingDigest = bindingDigest(second)
	fixture.input.Registry.Surfaces = append(fixture.input.Registry.Surfaces, second)
	fixture.input.Registry.Digest = stableDigest(registryCanonical(fixture.input.Registry))
	fixture.input.Config.RegistryDigest = fixture.input.Registry.Digest
	fixture.input.Manifest.RegistryDigest = fixture.input.Registry.Digest
	fixture.input.Manifest.Entries = append(fixture.input.Manifest.Entries, ManifestEntry{
		SurfaceID: second.SurfaceID, CodeSymbolID: second.CodeSymbolID, SemanticOwnerID: second.SemanticOwnerID,
		BeforeBindingDigest: second.Binding.BindingDigest, AfterBindingDigest: second.Binding.BindingDigest,
		BeforeBlobDigest: fixtureDigest("second-blob"), AfterBlobDigest: fixtureDigest("second-blob"),
		BeforeSourcePath: "second-before.go", AfterSourcePath: "second-after.go",
	})
	fixture.input.Manifest.ZeroChange = false
	fixture.input.Manifest.Digest = stableDigest(manifestCanonical(fixture.input.Manifest))
	fixture.authorityContext.Registry = fixture.input.Registry
	fixture.authorityContext.Registry.Surfaces = append([]Surface(nil), fixture.input.Registry.Surfaces...)
	result := Evaluate(fixture.input, fixture.authorityContext)
	if result.Status != StatusUnknown || result.Observation.ChangedSurfaces.Value != 1 || result.Observation.Receipts.Value != 1 {
		t.Fatalf("localization baseline = %#v", result)
	}
	if result.Observation.ChangedSurfaces.Value >= uint64(len(fixture.input.Registry.Surfaces)) {
		t.Fatal("changed-surface localization was not finite")
	}
}
