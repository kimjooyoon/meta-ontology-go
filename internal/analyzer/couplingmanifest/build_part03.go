package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func matchSnapshotAuthority(input Input, beforeDigest, headDigest string) *ConstructionError {
	if input.Authority.Schema == "" || input.Authority.Registry.Schema == "" || input.Authority.Registry.Digest == "" {
		return unknownError(CodeMissingAuthority, "detector authority context is incomplete")
	}
	if input.Authority.Registry.Surfaces == nil {
		return unknownError(CodeMissingAuthority, "detector authority registry surfaces are unavailable")
	}
	if input.SourceMap.Digest == "" {
		return unknownError(CodeMissingAuthority, "source-map digest is unavailable")
	}
	beforeRegistry, beforeRegistryErr := rawDigest(input.Before.RegistryDigest)
	headRegistry, headRegistryErr := rawDigest(input.Head.RegistryDigest)
	beforeSourceMap, beforeSourceMapErr := rawDigest(input.Before.SourceMapDigest)
	headSourceMap, headSourceMapErr := rawDigest(input.Head.SourceMapDigest)
	if beforeRegistryErr != nil || headRegistryErr != nil || beforeRegistry != input.Authority.Registry.Digest || headRegistry != input.Authority.Registry.Digest {
		return unknownError(CodeAuthorityDrift, "snapshot registry digest differs from detector authority")
	}
	if beforeSourceMapErr != nil || headSourceMapErr != nil || beforeSourceMap != input.SourceMap.Digest || headSourceMap != input.SourceMap.Digest {
		return unknownError(CodeAuthorityDrift, "snapshot source-map digest differs from source-map context")
	}
	if headDigest != input.Authority.SnapshotDigest {
		return unknownError(CodeAuthorityDrift, "head snapshot digest differs from detector authority")
	}
	return nil
}
func makeManifest(authority detector.AuthorityContext, beforeDigest, headDigest string, before, head map[semantic.ID]resolved) detector.ChangeManifest {
	entries := make([]detector.ManifestEntry, 0, len(authority.Registry.Surfaces))
	for _, surface := range sortedSurfaces(authority.Registry.Surfaces) {
		entry := detector.ManifestEntry{
			SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID,
			BeforeBindingDigest: absentDigest, AfterBindingDigest: surface.Binding.BindingDigest,
			BeforeBlobDigest: absentDigest, AfterBlobDigest: absentDigest,
			BeforeSourcePath: absentPath, AfterSourcePath: absentPath,
		}
		if value, ok := before[surface.SemanticOwnerID]; ok {
			entry.BeforeBindingDigest, entry.BeforeBlobDigest, entry.BeforeSourcePath = value.Observed.BindingDigest, value.Observed.BlobDigest, value.Observed.Path
		}
		if value, ok := head[surface.SemanticOwnerID]; ok {
			entry.AfterBlobDigest, entry.AfterSourcePath = value.Observed.BlobDigest, value.Observed.Path
		}
		entries = append(entries, entry)
	}
	manifest := detector.ChangeManifest{
		Schema: detector.ManifestSchemaV1, Complete: true,
		ZeroChange: zeroChange(entries), RegistryDigest: authority.Registry.Digest,
		ToolchainDigest: authority.ToolchainDigest, ProfileDigest: authority.ProfileDigest,
		BeforeSnapshotDigest: beforeDigest, AfterSnapshotDigest: headDigest, Entries: entries,
	}
	manifest.Digest = detectorManifestDigest(manifest)
	return manifest
}
func zeroChange(entries []detector.ManifestEntry) bool {
	for _, entry := range entries {
		if entry.BeforeBindingDigest != entry.AfterBindingDigest || entry.BeforeBlobDigest != entry.AfterBlobDigest {
			return false
		}
	}
	return true
}
