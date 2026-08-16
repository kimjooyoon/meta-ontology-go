package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const absentPath = "<absent>"

type observed struct {
	Role          string
	Path          string
	BlobDigest    string
	BindingDigest string
}

type resolved struct {
	Observed observed
}

// BuildOutput keeps detector output and adapter metadata separate. The
// detector result is produced by detector.Evaluate and is never synthesized by
// this package.
type BuildOutput struct {
	Manifest       detector.ChangeManifest
	DetectorResult detector.Result
	Metadata       Metadata
}

// Build constructs the exact detector ChangeManifest. Construction metadata is
// available through BuildDetailed; detector semantic decisions remain outside
// this method.
func Build(input Input) (Manifest, error) {
	output, err := BuildDetailed(input)
	return output.Manifest, err
}

// BuildDetailed constructs a manifest and runs the detector's own structural
// validation path against it. A missing external resource receipt is the
// expected result here because this adapter does not make the detector's final
// coupling decision.
func BuildDetailed(input Input) (BuildOutput, error) {
	if err := validateSnapshots(input); err != nil {
		return failedOutput(err), err
	}
	if err := validateSourceMapContext(input); err != nil {
		return failedOutput(err), err
	}
	beforeDigest, err := rawDigest(input.Before.Digest)
	if err != nil {
		constructionErr := unknownError(CodeInvalidSnapshot, "before snapshot digest is malformed")
		return failedOutput(constructionErr), constructionErr
	}
	headDigest, err := rawDigest(input.Head.Digest)
	if err != nil {
		constructionErr := unknownError(CodeInvalidSnapshot, "head snapshot digest is malformed")
		return failedOutput(constructionErr), constructionErr
	}
	if err := matchSnapshotAuthority(input, beforeDigest, headDigest); err != nil {
		return failedOutput(err), err
	}
	before, head, resolveErr := resolveSnapshots(input)
	if resolveErr != nil {
		return failedOutput(resolveErr), resolveErr
	}
	manifest := makeManifest(input.Authority, beforeDigest, headDigest, before, head)
	detectorResult := ValidateManifest(manifest, input.Authority)
	if err := acceptStructuralDetectorResult(detectorResult); err != nil {
		return BuildOutput{Manifest: manifest, DetectorResult: detectorResult}, err
	}
	metadata := completeMetadata(input.SourceMap.Digest, input.Authority.Registry.Surfaces, before, head)
	return BuildOutput{Manifest: manifest, DetectorResult: detectorResult, Metadata: metadata}, nil
}

// ValidateManifest exposes the detector's exact result for callers that have a
// complete detector input packet. This wrapper performs no normalization.
func ValidateManifest(manifest detector.ChangeManifest, authority detector.AuthorityContext) detector.Result {
	return detector.Evaluate(detectorInput(manifest, authority), authority)
}

// Evaluate forwards an exact detector packet and authority context unchanged.
func Evaluate(input detector.Input, authority detector.AuthorityContext) detector.Result {
	return detector.Evaluate(input, authority)
}

// BuildManifest is the descriptive spelling of Build.
func BuildManifest(input Input) (Manifest, error) { return Build(input) }

// Adapt is a vocabulary alias for Build.
func Adapt(input Input) (Manifest, error) { return Build(input) }

// New constructs a manifest from the explicit adapter input.
func New(input Input) (Manifest, error) { return Build(input) }

func detectorInput(manifest detector.ChangeManifest, authority detector.AuthorityContext) detector.Input {
	return detector.Input{
		Schema: detector.InputSchemaV1,
		Config: detector.Config{
			Schema: detector.ConfigSchemaV1, RegistryDigest: authority.Registry.Digest,
			ToolchainDigest: authority.ToolchainDigest, ProfileDigest: authority.ProfileDigest,
			SnapshotDigest: authority.SnapshotDigest, ExpectedProviderDigest: authority.ExpectedProviderDigest,
			ExpectedObserverDigest: authority.ExpectedObserverDigest, Baseline: authority.Baseline,
			ExternalReceiptRequired: authority.ExternalReceiptRequired,
		},
		Registry: authority.Registry, Manifest: manifest, Receipts: []detector.CouplingReceipt{},
	}
}

func validateSnapshots(input Input) *ConstructionError {
	if input.Before == nil || input.Head == nil {
		return unknownError(CodeMissingSnapshot, "before and head snapshots are required")
	}
	if err := input.Before.Validate(); err != nil {
		return unknownError(CodeInvalidSnapshot, "before snapshot is not valid: %s", err.Error())
	}
	if err := input.Head.Validate(); err != nil {
		return unknownError(CodeInvalidSnapshot, "head snapshot is not valid: %s", err.Error())
	}
	return nil
}

func matchSnapshotAuthority(input Input, beforeDigest, headDigest string) *ConstructionError {
	if input.Authority.Schema == "" || input.Authority.Registry.Schema == "" || input.Authority.Registry.Digest == "" {
		return unknownError(CodeMissingAuthority, "detector authority context is incomplete")
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
