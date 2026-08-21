package coupling

import (
	"testing"
)

func fixtureContextFor(t *testing.T, claim ChangeClaim) fixtureContext {
	t.Helper()
	owner, code, surface := fixtureID("owner"), fixtureID("code-symbol"), fixtureID("surface")
	sourceMap := fixtureID("source-map")
	registrySurface := Surface{
		SurfaceID: surface, CodeSymbolID: code, SemanticOwnerID: owner,
		Binding:           SourceMapBinding{SourceMapID: sourceMap, PackageLabel: "billing", FileLabel: "old.go", SourceSpan: "1:1-1:2"},
		PresentationLabel: "PayOrder",
	}
	registrySurface.Binding.BindingDigest = bindingDigest(registrySurface)
	registry := Registry{Schema: RegistrySchemaV1, Surfaces: []Surface{registrySurface}}
	registry.Digest = stableDigest(registryCanonical(registry))
	baseline := BaselineConfig{Schema: BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = stableDigest(baselineCanonical(baseline))
	config := Config{
		Schema: ConfigSchemaV1, RegistryDigest: registry.Digest, ToolchainDigest: fixtureDigest("toolchain"),
		ProfileDigest: fixtureDigest("profile"), SnapshotDigest: fixtureDigest("snapshot-after"),
		ExpectedProviderDigest: fixtureDigest("provider"), ExpectedObserverDigest: fixtureDigest("observer"),
		Baseline: baseline, ExternalReceiptRequired: true,
	}
	beforeSemantic, afterSemantic := fixtureDigest("semantic-before"), fixtureDigest("semantic-after")
	beforeSource, afterSource := fixtureDigest("source-before"), fixtureDigest("source-after")
	if claim == ChangeClaimNoDelta {
		afterSemantic, afterSource = beforeSemantic, beforeSource
	}
	beforeBlob, afterBlob := fixtureDigest("blob-before"), fixtureDigest("blob-after")
	entry := ManifestEntry{
		SurfaceID: surface, CodeSymbolID: code, SemanticOwnerID: owner,
		BeforeBindingDigest: registrySurface.Binding.BindingDigest, AfterBindingDigest: registrySurface.Binding.BindingDigest,
		BeforeBlobDigest: beforeBlob, AfterBlobDigest: afterBlob,
		BeforeSourcePath: "/workspace/old.go", AfterSourcePath: "/relocated/new.go",
	}
	manifest := ChangeManifest{
		Schema: ManifestSchemaV1, Complete: true, ZeroChange: false, RegistryDigest: registry.Digest,
		ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeSnapshotDigest: fixtureDigest("snapshot-before"), AfterSnapshotDigest: config.SnapshotDigest,
		Entries: []ManifestEntry{entry},
	}
	manifest.Digest = stableDigest(manifestCanonical(manifest))
	return fixtureContext{owner: owner, code: code, surface: surface, registry: registry, config: config, manifest: manifest,
		beforeBlob: beforeBlob, afterBlob: afterBlob, beforeSemantic: beforeSemantic, afterSemantic: afterSemantic,
		beforeSource: beforeSource, afterSource: afterSource}
}
