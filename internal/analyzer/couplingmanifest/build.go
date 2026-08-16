package couplingmanifest

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const absentPath = "<absent>"

type observed struct {
	Role          semanticbinding.Role
	Path          string
	BlobDigest    string
	BindingDigest string
}

type resolved struct {
	Surface  Surface
	Observed observed
}

type normalizedAuthority struct {
	registryDigest  string
	sourceMapDigest string
	toolchainDigest string
	profileDigest   string
	inventory       map[semantic.ID]Surface
	before          map[semantic.ID]SourceMapObservation
	head            map[semantic.ID]SourceMapObservation
}

// Build constructs a complete source-backed change manifest. Change derivation
// remains the detector's responsibility; this adapter emits the full
// registered inventory and exact before/after source evidence only.
func Build(input Input) (Manifest, error) {
	if err := validateBuildInput(input); err != nil {
		return unknownResult(err.Code, err.Message, input.Authority)
	}
	authority, err := normalizeAuthority(input.Authority)
	if err != nil {
		if typed, ok := err.(*Error); ok && typed.Status == StatusUnknown {
			return unknownResult(typed.Code, typed.Message, input.Authority)
		}
		return failResult(err, input.Authority)
	}
	beforeDigest, headDigest, digestErr := snapshotAuthorityDigests(input, authority)
	if digestErr != nil {
		return unknownResult(digestErr.Code, digestErr.Message, input.Authority)
	}
	resolvedBefore, resolvedHead, err := resolveSnapshots(input, authority)
	if err != nil {
		return resultForResolution(err, input.Authority)
	}
	manifest := completeManifest(authority, beforeDigest, headDigest, resolvedBefore, resolvedHead)
	return sealResult(manifest)
}

// BuildManifest is the descriptive spelling of Build.
func BuildManifest(input Input) (Manifest, error) { return Build(input) }

// Adapt is a vocabulary alias for Build.
func Adapt(input Input) (Manifest, error) { return Build(input) }

// New constructs a change manifest from the explicit adapter input.
func New(input Input) (Manifest, error) { return Build(input) }

func validateBuildInput(input Input) *Error {
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

func snapshotAuthorityDigests(input Input, authority normalizedAuthority) (string, string, *Error) {
	beforeDigest, err := rawDigest(input.Before.Digest)
	if err != nil {
		return "", "", unknownError(CodeInvalidSnapshot, "before snapshot digest is malformed")
	}
	headDigest, err := rawDigest(input.Head.Digest)
	if err != nil {
		return "", "", unknownError(CodeInvalidSnapshot, "head snapshot digest is malformed")
	}
	if registryDigest, _ := rawDigest(input.Before.RegistryDigest); registryDigest != authority.registryDigest {
		return "", "", unknownError(CodeAuthorityDrift, "before snapshot registry digest does not agree with authority")
	}
	if registryDigest, _ := rawDigest(input.Head.RegistryDigest); registryDigest != authority.registryDigest {
		return "", "", unknownError(CodeAuthorityDrift, "head snapshot registry digest does not agree with authority")
	}
	if sourceMapDigest, _ := rawDigest(input.Before.SourceMapDigest); sourceMapDigest != authority.sourceMapDigest {
		return "", "", unknownError(CodeAuthorityDrift, "before snapshot source-map digest does not agree with authority")
	}
	if sourceMapDigest, _ := rawDigest(input.Head.SourceMapDigest); sourceMapDigest != authority.sourceMapDigest {
		return "", "", unknownError(CodeAuthorityDrift, "head snapshot source-map digest does not agree with authority")
	}
	return beforeDigest, headDigest, nil
}

func resolveSnapshots(input Input, authority normalizedAuthority) (map[semantic.ID]resolved, map[semantic.ID]resolved, error) {
	before, err := snapshotIndex(*input.Before)
	if err != nil {
		return nil, nil, unknownError(CodeInvalidSnapshot, "%v", err)
	}
	head, err := snapshotIndex(*input.Head)
	if err != nil {
		return nil, nil, unknownError(CodeInvalidSnapshot, "%v", err)
	}
	if err := rejectUnregistered(before, authority.inventory); err != nil {
		return nil, nil, err
	}
	if err := rejectUnregistered(head, authority.inventory); err != nil {
		return nil, nil, err
	}
	if err := rejectUnobservedBindings(before, authority.before); err != nil {
		return nil, nil, err
	}
	if err := rejectUnobservedBindings(head, authority.head); err != nil {
		return nil, nil, err
	}
	resolvedBefore, err := resolveSide(before, authority.before, authority.inventory, false)
	if err != nil {
		return nil, nil, err
	}
	resolvedHead, err := resolveSide(head, authority.head, authority.inventory, true)
	if err != nil {
		return nil, nil, err
	}
	if err := requireInventoryCoverage(authority.inventory, resolvedBefore, resolvedHead); err != nil {
		return nil, nil, err
	}
	return resolvedBefore, resolvedHead, nil
}

func completeManifest(authority normalizedAuthority, beforeDigest, headDigest string, before, head map[semantic.ID]resolved) Manifest {
	surfaceIDs := sortedSurfaceIDs(authority.inventory)
	entries := buildEntries(surfaceIDs, authority, before, head)
	return Manifest{
		Schema: SchemaV1, Complete: true, ZeroChange: isZeroChange(entries),
		BeforeSnapshotDigest: beforeDigest, AfterSnapshotDigest: headDigest,
		RegistryDigest: authority.registryDigest, ToolchainDigest: authority.toolchainDigest,
		ProfileDigest: authority.profileDigest, SourceMapDigest: authority.sourceMapDigest,
		Entries: entries, Status: StatusComplete,
		ResolvedSurfaceIDs: append([]semantic.ID(nil), surfaceIDs...),
		Counts:             ComponentCounts{Registered: len(surfaceIDs), Before: len(before), Head: len(head), Resolved: len(entries)},
		Work:               Work{ComponentCount: len(entries), WorkUnits: len(entries)},
		HeadSnapshotDigest: headDigest, FullSuiteRequired: false, statsKnown: true,
	}
}

func buildEntries(surfaceIDs []semantic.ID, authority normalizedAuthority, before, head map[semantic.ID]resolved) []ManifestEntry {
	entries := make([]ManifestEntry, 0, len(surfaceIDs))
	for _, surfaceID := range surfaceIDs {
		surface := authority.inventory[surfaceID]
		entry := ManifestEntry{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID, SourceMapID: surface.Binding.SourceMapID}
		applyEntrySide(&entry, surface, before, authority.before, true)
		applyEntrySide(&entry, surface, head, authority.head, false)
		entries = append(entries, entry)
	}
	return entries
}

func applyEntrySide(entry *ManifestEntry, surface Surface, resolved map[semantic.ID]resolved, bindings map[semantic.ID]SourceMapObservation, before bool) {
	value, present := resolved[surface.SemanticOwnerID]
	if before {
		if !present {
			entry.BeforeBindingDigest, entry.BeforeBlobDigest, entry.BeforeSourcePath = absentDigest, absentDigest, absentPath
			return
		}
		entry.BeforePresent = true
		entry.BeforeBindingDigest = bindings[surface.SemanticOwnerID].SourceMapBindingDigest
		entry.BeforeSourceMapBindingDigest = entry.BeforeBindingDigest
		entry.BeforeBlobDigest, entry.BeforeSourcePath = value.Observed.BlobDigest, value.Observed.Path
		return
	}
	if !present {
		entry.AfterBindingDigest, entry.AfterBlobDigest, entry.AfterSourcePath = surface.Binding.BindingDigest, absentDigest, absentPath
		return
	}
	entry.AfterPresent = true
	entry.AfterBindingDigest = bindings[surface.SemanticOwnerID].SourceMapBindingDigest
	entry.AfterSourceMapBindingDigest = entry.AfterBindingDigest
	entry.AfterBlobDigest, entry.AfterSourcePath = value.Observed.BlobDigest, value.Observed.Path
}
