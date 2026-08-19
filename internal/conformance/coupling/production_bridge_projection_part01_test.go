package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

func detectorInputFromCanonical(input Input) production.Input {
	registry := productionRegistryFromCanonical(input)
	config := productionConfigFromCanonical(input, registry)
	manifest := productionManifestFromCanonical(input, registry, config)
	path := productionPathFromCanonical(input.Path, input.Receipts, input.Roots)
	return production.Input{
		Schema: production.InputSchemaV1, Config: config, Registry: registry, Manifest: manifest,
		Receipts: productionReceiptsFromCanonical(input, registry, config, path), InferencePath: path,
		ExternalReceipt: productionExternalReceiptFromCanonical(input, config), WorkspaceRoot: input.Config.Profile.ID,
	}
}
func productionRegistryFromCanonical(input Input) production.Registry {
	registry := production.Registry{Schema: production.RegistrySchemaV1}
	for _, raw := range input.Registry {
		surface := production.Surface{
			SurfaceID: bridgeID(raw.RegisteredSurfaceID), CodeSymbolID: bridgeID(raw.CodeSymbolID), SemanticOwnerID: bridgeID(raw.SemanticOwnerID),
			Binding:           production.SourceMapBinding{SourceMapID: bridgeSourceMapID(raw.SourceMapID), BindingDigest: bridgeBindingDigestValues(raw.RegisteredSurfaceID, raw.CodeSymbolID, raw.SemanticOwnerID, bridgeSourceMapID(raw.SourceMapID).String()), PackageLabel: raw.PackageLabel, FileLabel: raw.FileLabel, SourceSpan: raw.SourceSpan},
			PresentationLabel: raw.RegisteredSurfaceID,
		}
		registry.Surfaces = append(registry.Surfaces, surface)
	}
	registry.Digest = bridgeRegistryDigest(registry)
	return registry
}
func productionConfigFromCanonical(input Input, registry production.Registry) production.Config {
	baseline := production.BaselineConfig{Schema: production.BaselineSchemaV1, FullSuiteRequired: true}
	baseline.Digest = bridgeBaselineDigest(baseline)
	return production.Config{Schema: production.ConfigSchemaV1, RegistryDigest: registry.Digest,
		ToolchainDigest: bridgeRawDigest(input.Config.ToolchainDigest), ProfileDigest: bridgeRawDigest(input.Config.Profile.Digest),
		SnapshotDigest: bridgeRawDigest(input.Config.ResourceBinding.SnapshotDigest), ExpectedProviderDigest: bridgeRawDigest(input.Config.ResourceBinding.ProviderDigest),
		ExpectedObserverDigest: bridgeRawDigest(input.Config.ResourceBinding.ObserverDigest), Baseline: baseline, ExternalReceiptRequired: true}
}
func productionManifestFromCanonical(input Input, registry production.Registry, config production.Config) production.ChangeManifest {
	manifest := production.ChangeManifest{Schema: production.ManifestSchemaV1, Complete: input.Manifest.Complete, ZeroChange: input.Manifest.ZeroChange,
		RegistryDigest: registry.Digest, ToolchainDigest: config.ToolchainDigest, ProfileDigest: config.ProfileDigest,
		BeforeSnapshotDigest: bridgeRawDigest(input.Manifest.BeforeSnapshotDigest), AfterSnapshotDigest: config.SnapshotDigest}
	changes := make(map[string]CodeChange, len(input.Changes))
	for _, change := range input.Changes {
		changes[change.CodeSymbolID] = change
	}
	for _, surface := range registry.Surfaces {
		change := changes[surface.CodeSymbolID.String()]
		before, after := bridgeRawDigest(change.BeforeDigest), bridgeRawDigest(change.AfterDigest)
		if input.Manifest.ZeroChange && before == "" && after == "" {
			before, after = bridgeHash("unchanged:"+surface.CodeSymbolID.String()), bridgeHash("unchanged:"+surface.CodeSymbolID.String())
		}
		manifest.Entries = append(manifest.Entries, production.ManifestEntry{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID,
			BeforeBindingDigest: surface.Binding.BindingDigest, AfterBindingDigest: surface.Binding.BindingDigest,
			BeforeBlobDigest: before, AfterBlobDigest: after,
			BeforeSourcePath: surface.Binding.FileLabel, AfterSourcePath: surface.Binding.FileLabel})
	}
	manifest.Digest = bridgeManifestDigest(manifest)
	return manifest
}
