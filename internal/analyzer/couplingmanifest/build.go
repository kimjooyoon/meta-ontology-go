package couplingmanifest

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
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
	if input.Before == nil || input.Head == nil {
		return unknownResult(CodeMissingSnapshot, "before and head snapshots are required", input.Authority)
	}
	if err := input.Before.Validate(); err != nil {
		return unknownResult(CodeInvalidSnapshot, "before snapshot is not valid: "+err.Error(), input.Authority)
	}
	if err := input.Head.Validate(); err != nil {
		return unknownResult(CodeInvalidSnapshot, "head snapshot is not valid: "+err.Error(), input.Authority)
	}
	authority, err := normalizeAuthority(input.Authority)
	if err != nil {
		if typed, ok := err.(*Error); ok && typed.Status == StatusUnknown {
			return unknownResult(typed.Code, typed.Message, input.Authority)
		}
		return failResult(err, input.Authority)
	}
	beforeSnapshotDigest, _ := rawDigest(input.Before.Digest)
	headSnapshotDigest, _ := rawDigest(input.Head.Digest)
	if registryDigest, _ := rawDigest(input.Before.RegistryDigest); registryDigest != authority.registryDigest {
		return unknownResult(CodeAuthorityDrift, "before snapshot registry digest does not agree with authority", input.Authority)
	}
	if registryDigest, _ := rawDigest(input.Head.RegistryDigest); registryDigest != authority.registryDigest {
		return unknownResult(CodeAuthorityDrift, "head snapshot registry digest does not agree with authority", input.Authority)
	}
	if sourceMapDigest, _ := rawDigest(input.Before.SourceMapDigest); sourceMapDigest != authority.sourceMapDigest {
		return unknownResult(CodeAuthorityDrift, "before snapshot source-map digest does not agree with authority", input.Authority)
	}
	if sourceMapDigest, _ := rawDigest(input.Head.SourceMapDigest); sourceMapDigest != authority.sourceMapDigest {
		return unknownResult(CodeAuthorityDrift, "head snapshot source-map digest does not agree with authority", input.Authority)
	}

	before, err := snapshotIndex(*input.Before)
	if err != nil {
		return unknownResult(CodeInvalidSnapshot, err.Error(), input.Authority)
	}
	head, err := snapshotIndex(*input.Head)
	if err != nil {
		return unknownResult(CodeInvalidSnapshot, err.Error(), input.Authority)
	}
	if err := rejectUnregistered(before, authority.inventory); err != nil {
		return failResult(err, input.Authority)
	}
	if err := rejectUnregistered(head, authority.inventory); err != nil {
		return failResult(err, input.Authority)
	}
	if err := rejectUnobservedBindings(before, authority.before); err != nil {
		return resultForResolution(err, input.Authority)
	}
	if err := rejectUnobservedBindings(head, authority.head); err != nil {
		return resultForResolution(err, input.Authority)
	}
	resolvedBefore, err := resolveSide(before, authority.before, authority.inventory, false)
	if err != nil {
		return resultForResolution(err, input.Authority)
	}
	resolvedHead, err := resolveSide(head, authority.head, authority.inventory, true)
	if err != nil {
		return resultForResolution(err, input.Authority)
	}
	if err := requireInventoryCoverage(authority.inventory, resolvedBefore, resolvedHead); err != nil {
		return resultForResolution(err, input.Authority)
	}

	surfaceIDs := sortedSurfaceIDs(authority.inventory)
	entries := make([]ManifestEntry, 0, len(surfaceIDs))
	for _, surfaceID := range surfaceIDs {
		registered := authority.inventory[surfaceID]
		entry := ManifestEntry{
			SurfaceID:       registered.SurfaceID,
			CodeSymbolID:    registered.CodeSymbolID,
			SemanticOwnerID: registered.SemanticOwnerID,
			SourceMapID:     registered.Binding.SourceMapID,
		}
		if value, ok := resolvedBefore[registered.SemanticOwnerID]; ok {
			entry.BeforePresent = true
			entry.BeforeBindingDigest = authority.before[registered.SemanticOwnerID].SourceMapBindingDigest
			entry.BeforeSourceMapBindingDigest = entry.BeforeBindingDigest
			entry.BeforeBlobDigest = value.Observed.BlobDigest
			entry.BeforeSourcePath = value.Observed.Path
		} else {
			entry.BeforeBindingDigest = absentDigest
			entry.BeforeBlobDigest = absentDigest
			entry.BeforeSourcePath = absentPath
		}
		if value, ok := resolvedHead[registered.SemanticOwnerID]; ok {
			entry.AfterPresent = true
			entry.AfterBindingDigest = authority.head[registered.SemanticOwnerID].SourceMapBindingDigest
			entry.AfterSourceMapBindingDigest = entry.AfterBindingDigest
			entry.AfterBlobDigest = value.Observed.BlobDigest
			entry.AfterSourcePath = value.Observed.Path
		} else {
			// The detector contract requires the after binding to remain the
			// registered current binding. Absence is represented by the reserved
			// blob digest and locator while the adapter metadata retains presence.
			entry.AfterBindingDigest = registered.Binding.BindingDigest
			entry.AfterBlobDigest = absentDigest
			entry.AfterSourcePath = absentPath
		}
		entries = append(entries, entry)
	}

	counts := ComponentCounts{Registered: len(surfaceIDs), Before: len(resolvedBefore), Head: len(resolvedHead), Resolved: len(entries)}
	manifest := Manifest{
		Schema: SchemaV1, Complete: true,
		ZeroChange:           isZeroChange(entries),
		BeforeSnapshotDigest: beforeSnapshotDigest,
		AfterSnapshotDigest:  headSnapshotDigest,
		RegistryDigest:       authority.registryDigest,
		ToolchainDigest:      authority.toolchainDigest,
		ProfileDigest:        authority.profileDigest,
		SourceMapDigest:      authority.sourceMapDigest,
		Entries:              entries,
		Status:               StatusComplete,
		ResolvedSurfaceIDs:   append([]semantic.ID(nil), surfaceIDs...),
		Counts:               counts,
		Work:                 Work{ComponentCount: len(entries), WorkUnits: len(entries)},
		HeadSnapshotDigest:   headSnapshotDigest,
		FullSuiteRequired:    false,
		statsKnown:           true,
	}
	return sealResult(manifest)
}

// BuildManifest is the descriptive spelling of Build.
func BuildManifest(input Input) (Manifest, error) { return Build(input) }

// Adapt is a vocabulary alias for Build.
func Adapt(input Input) (Manifest, error) { return Build(input) }

// New constructs a change manifest from the explicit adapter input.
func New(input Input) (Manifest, error) { return Build(input) }

func normalizeAuthority(input RegistrySourceMap) (normalizedAuthority, error) {
	if input.Schema == "" || input.RegistryDigest == "" || input.SourceMapDigest == "" || input.ToolchainDigest == "" || input.ProfileDigest == "" || input.Inventory == nil || input.Before == nil || input.Head == nil {
		return normalizedAuthority{}, unknownError(CodeMissingAuthority, "versioned registry/source-map, toolchain/profile, inventory, and both side bindings are required")
	}
	if input.Schema != AuthoritySchemaV1 {
		return normalizedAuthority{}, failError(CodeInvalidSchema, "authority schema %q is not %q", input.Schema, AuthoritySchemaV1)
	}
	registryDigest, err := normalizeDigest(input.RegistryDigest, "registry digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	sourceMapDigest, err := normalizeDigest(input.SourceMapDigest, "source-map digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	toolchainDigest, err := normalizeDigest(input.ToolchainDigest, "toolchain digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	profileDigest, err := normalizeDigest(input.ProfileDigest, "profile digest")
	if err != nil {
		return normalizedAuthority{}, err
	}
	if len(input.CandidateBindings) != 0 {
		return normalizedAuthority{}, unknownError(CodeCandidateBinding, "candidate bindings cannot become authoritative")
	}
	if len(input.DerivedBindings) != 0 {
		return normalizedAuthority{}, unknownError(CodeDerivedBinding, "derived bindings cannot become authoritative")
	}
	inventory, err := normalizeInventory(input.Inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	before, err := normalizeObservations(input.Before, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	head, err := normalizeObservations(input.Head, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	return normalizedAuthority{registryDigest: registryDigest, sourceMapDigest: sourceMapDigest, toolchainDigest: toolchainDigest, profileDigest: profileDigest, inventory: inventory, before: before, head: head}, nil
}

func normalizeInventory(values []Surface) (map[semantic.ID]Surface, error) {
	result := make(map[semantic.ID]Surface, len(values))
	seenSymbols := make(map[semantic.ID]semantic.ID, len(values))
	seenOwners := make(map[semantic.ID]semantic.ID, len(values))
	seenMaps := make(map[semantic.ID]semantic.ID, len(values))
	for _, value := range values {
		normalized, err := normalizeSurface(value)
		if err != nil {
			return nil, err
		}
		if _, exists := result[normalized.SurfaceID]; exists {
			return nil, failError(CodeDuplicateBinding, "surface ID %q occurs more than once", normalized.SurfaceID)
		}
		if previous, exists := seenSymbols[normalized.CodeSymbolID]; exists {
			return nil, failError(CodeConflictingBinding, "code symbol ID %q resolves to both %q and %q", normalized.CodeSymbolID, previous, normalized.SurfaceID)
		}
		if previous, exists := seenOwners[normalized.SemanticOwnerID]; exists {
			return nil, failError(CodeConflictingBinding, "semantic owner ID %q resolves to both %q and %q", normalized.SemanticOwnerID, previous, normalized.SurfaceID)
		}
		if previous, exists := seenMaps[normalized.Binding.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q resolves to both %q and %q", normalized.Binding.SourceMapID, previous, normalized.SurfaceID)
		}
		result[normalized.SurfaceID] = normalized
		seenSymbols[normalized.CodeSymbolID] = normalized.SurfaceID
		seenOwners[normalized.SemanticOwnerID] = normalized.SurfaceID
		seenMaps[normalized.Binding.SourceMapID] = normalized.SurfaceID
	}
	return result, nil
}

func normalizeSurface(value Surface) (Surface, error) {
	for _, field := range []struct {
		value semantic.ID
		name  string
	}{
		{value.SurfaceID, "surface ID"},
		{value.CodeSymbolID, "code symbol ID"},
		{value.SemanticOwnerID, "semantic owner ID"},
		{value.Binding.SourceMapID, "source-map ID"},
	} {
		if _, err := normalizeID(field.value); err != nil {
			return Surface{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	bindingDigest, err := normalizeDigest(value.Binding.BindingDigest, "source-map binding digest")
	if err != nil {
		return Surface{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	value.Binding.BindingDigest = bindingDigest
	if expected := sourceMapBindingDigest(value); expected != bindingDigest {
		return Surface{}, failError(CodeConflictingBinding, "surface %q has a stale source-map binding digest", value.SurfaceID)
	}
	return value, nil
}

func normalizeObservations(values []SourceMapObservation, inventory map[semantic.ID]Surface) (map[semantic.ID]SourceMapObservation, error) {
	result := make(map[semantic.ID]SourceMapObservation, len(values))
	seenSurfaces := make(map[semantic.ID]struct{}, len(values))
	seenSymbols := make(map[semantic.ID]struct{}, len(values))
	seenMaps := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		normalized, err := normalizeObservation(value)
		if err != nil {
			return nil, err
		}
		registered, ok := inventory[normalized.SurfaceID]
		if !ok {
			return nil, failError(CodeMalformedBinding, "surface ID %q is not registered", normalized.SurfaceID)
		}
		if normalized.CodeSymbolID != registered.CodeSymbolID || normalized.SemanticOwnerID != registered.SemanticOwnerID || normalized.SourceMapID != registered.Binding.SourceMapID {
			return nil, failError(CodeConflictingBinding, "surface ID %q conflicts with its registered identity tuple", normalized.SurfaceID)
		}
		if _, exists := seenSurfaces[normalized.SurfaceID]; exists {
			return nil, failError(CodeDuplicateBinding, "surface ID %q occurs more than once", normalized.SurfaceID)
		}
		if _, exists := seenSymbols[normalized.CodeSymbolID]; exists {
			return nil, failError(CodeConflictingBinding, "code symbol ID %q occurs more than once", normalized.CodeSymbolID)
		}
		if _, exists := seenMaps[normalized.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q occurs more than once", normalized.SourceMapID)
		}
		result[normalized.SemanticOwnerID] = normalized
		seenSurfaces[normalized.SurfaceID] = struct{}{}
		seenSymbols[normalized.CodeSymbolID] = struct{}{}
		seenMaps[normalized.SourceMapID] = struct{}{}
	}
	return result, nil
}

func normalizeObservation(value SourceMapObservation) (SourceMapObservation, error) {
	for _, field := range []struct {
		value semantic.ID
		name  string
	}{
		{value.SurfaceID, "surface ID"},
		{value.CodeSymbolID, "code symbol ID"},
		{value.SemanticOwnerID, "semantic owner ID"},
		{value.SourceMapID, "source-map ID"},
	} {
		if _, err := normalizeID(field.value); err != nil {
			return SourceMapObservation{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	if !validRole(value.Role) {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q has invalid role %q", value.SurfaceID, value.Role)
	}
	repoPath, err := normalizeRepoPath(value.Path)
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q path: %v", value.SurfaceID, err)
	}
	blobDigest, err := normalizeDigest(value.BlobDigest, "blob digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	bindingDigest, err := normalizeDigest(value.BindingDigest, "binding digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	sourceMapBindingDigest, err := normalizeDigest(value.SourceMapBindingDigest, "source-map binding digest")
	if err != nil {
		return SourceMapObservation{}, failError(CodeMalformedBinding, "surface %q: %v", value.SurfaceID, err)
	}
	value.Path, value.BlobDigest, value.BindingDigest, value.SourceMapBindingDigest = repoPath, blobDigest, bindingDigest, sourceMapBindingDigest
	return value, nil
}

func snapshotIndex(snapshot selectiveci.Snapshot) (map[semantic.ID]observed, error) {
	result := make(map[semantic.ID]observed)
	for _, source := range snapshot.Sources {
		blobDigest, err := normalizeDigest(source.BlobDigest, "snapshot blob digest")
		if err != nil {
			return nil, err
		}
		for _, binding := range source.Bindings {
			ownerID, err := normalizeIDString(binding.ID)
			if err != nil {
				return nil, err
			}
			bindingDigest, err := normalizeDigest(binding.BindingDigest, "snapshot binding digest")
			if err != nil {
				return nil, err
			}
			if _, exists := result[ownerID]; exists {
				return nil, fmt.Errorf("snapshot semantic owner ID %q occurs more than once", ownerID)
			}
			result[ownerID] = observed{Role: binding.Role, Path: source.Path, BlobDigest: blobDigest, BindingDigest: bindingDigest}
		}
	}
	return result, nil
}

func rejectUnregistered(values map[semantic.ID]observed, inventory map[semantic.ID]Surface) error {
	knownOwners := make(map[semantic.ID]struct{}, len(inventory))
	for _, surface := range inventory {
		knownOwners[surface.SemanticOwnerID] = struct{}{}
	}
	for ownerID := range values {
		if _, ok := knownOwners[ownerID]; !ok {
			return failError(CodeMalformedBinding, "semantic owner ID %q is not registered", ownerID)
		}
	}
	return nil
}

func resolveSide(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation, inventory map[semantic.ID]Surface, current bool) (map[semantic.ID]resolved, error) {
	result := make(map[semantic.ID]resolved, len(snapshot))
	for ownerID, value := range snapshot {
		binding, ok := bindings[ownerID]
		if !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has no exact source-map binding", ownerID)
		}
		registered := inventory[binding.SurfaceID]
		if registered.SemanticOwnerID != ownerID || binding.Path != value.Path || binding.Role != value.Role || binding.BlobDigest != value.BlobDigest || binding.BindingDigest != value.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has a stale or conflicting source-map binding", ownerID)
		}
		if current && binding.SourceMapBindingDigest != registered.Binding.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has a stale current source-map binding", ownerID)
		}
		result[ownerID] = resolved{Surface: registered, Observed: value}
	}
	return result, nil
}

func rejectUnobservedBindings(snapshot map[semantic.ID]observed, bindings map[semantic.ID]SourceMapObservation) error {
	for ownerID := range bindings {
		if _, observed := snapshot[ownerID]; !observed {
			return unknownError(CodeUnknownChangedSurface, "source-map binding for semantic owner ID %q has no matching snapshot observation", ownerID)
		}
	}
	return nil
}

func requireInventoryCoverage(inventory map[semantic.ID]Surface, before, head map[semantic.ID]resolved) error {
	for _, surface := range inventory {
		if _, beforeOK := before[surface.SemanticOwnerID]; !beforeOK {
			if _, headOK := head[surface.SemanticOwnerID]; !headOK {
				return unknownError(CodeUnknownChangedSurface, "registered surface %q has no before or head observation", surface.SurfaceID)
			}
		}
	}
	return nil
}

func sortedSurfaceIDs(values map[semantic.ID]Surface) []semantic.ID {
	result := make([]semantic.ID, 0, len(values))
	for surfaceID := range values {
		result = append(result, surfaceID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func normalizeID(value semantic.ID) (semantic.ID, error) {
	if value == "" {
		return "", fmt.Errorf("ID is empty")
	}
	parsed, err := semantic.ParseIdentity(value.String())
	if err != nil || parsed != value {
		return "", fmt.Errorf("ID %q is not canonical", value)
	}
	return parsed, nil
}

func normalizeIDString(value string) (semantic.ID, error) {
	return normalizeID(semantic.ID(value))
}

func normalizeRepoPath(value string) (string, error) {
	if value == "" || value != strings.TrimSpace(value) || strings.Contains(value, "\\") || strings.HasPrefix(value, "/") || (len(value) > 1 && value[1] == ':') {
		return "", fmt.Errorf("path %q is not a repository-relative path", value)
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path %q escapes the repository", value)
	}
	return clean, nil
}

func validRole(role semanticbinding.Role) bool {
	return role == semanticbinding.RoleHandwrittenImpl || role == semanticbinding.RoleGeneratedImpl || role == semanticbinding.RoleAdapter
}

func normalizeDigest(value, label string) (string, error) {
	canonical, err := rawDigest(value)
	if err != nil {
		return "", failError(CodeMalformedBinding, "%s is malformed", label)
	}
	return canonical, nil
}

func rawDigest(value string) (string, error) {
	if strings.HasPrefix(value, "sha256:") {
		value = strings.TrimPrefix(value, "sha256:")
	}
	if len(value) != 64 || value != strings.ToLower(value) {
		return "", fmt.Errorf("digest is not lowercase SHA-256")
	}
	for _, char := range value {
		if char < '0' || char > '9' && char < 'a' || char > 'f' {
			return "", fmt.Errorf("digest is not lowercase SHA-256")
		}
	}
	return value, nil
}

func sourceMapBindingDigest(surface Surface) string {
	var builder strings.Builder
	field(&builder, surface.SurfaceID.String())
	field(&builder, surface.CodeSymbolID.String())
	field(&builder, surface.SemanticOwnerID.String())
	field(&builder, surface.Binding.SourceMapID.String())
	return stableDigest(builder.String())
}

func isZeroChange(entries []ManifestEntry) bool {
	for _, entry := range entries {
		if entry.BeforeBindingDigest != entry.AfterBindingDigest || entry.BeforeBlobDigest != entry.AfterBlobDigest {
			return false
		}
	}
	return true
}

func resultForResolution(err error, authority RegistrySourceMap) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		return unknownResult(CodeUnknownChangedSurface, err.Error(), authority)
	}
	if typed.Status == StatusFailClosed {
		return failResult(typed, authority)
	}
	return unknownResult(typed.Code, typed.Message, authority)
}

func failError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: StatusFailClosed, FullSuiteRequired: true}
}

func unknownError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: StatusUnknown, FullSuiteRequired: true}
}

func unknownResult(code ErrorCode, message string, authority RegistrySourceMap) (Manifest, error) {
	manifest := incompleteManifest(StatusUnknown, code, authority)
	return sealResult(manifest, unknownError(code, "%s", message))
}

func failResult(err error, authority RegistrySourceMap) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		typed = failError(CodeInvalidManifest, "%v", err)
	}
	manifest := incompleteManifest(StatusFailClosed, typed.Code, authority)
	return sealResult(manifest, typed)
}

func incompleteManifest(status Status, code ErrorCode, authority RegistrySourceMap) Manifest {
	manifest := Manifest{
		Schema: SchemaV1, Complete: false, ZeroChange: false,
		Entries: []ManifestEntry{}, Status: status, FullSuiteRequired: true,
		ReasonCode: code, ResolvedSurfaceIDs: []semantic.ID{}, Counts: ComponentCounts{}, Work: Work{}, statsKnown: true,
	}
	if digest, err := rawDigest(authority.RegistryDigest); err == nil {
		manifest.RegistryDigest = digest
	}
	if digest, err := rawDigest(authority.SourceMapDigest); err == nil {
		manifest.SourceMapDigest = digest
	}
	if digest, err := rawDigest(authority.ToolchainDigest); err == nil {
		manifest.ToolchainDigest = digest
	}
	if digest, err := rawDigest(authority.ProfileDigest); err == nil {
		manifest.ProfileDigest = digest
	}
	return manifest
}
