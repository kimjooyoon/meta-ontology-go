package couplingmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type observed struct {
	OwnerID       string
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
	inventory       map[string]Surface
	before          map[string]SourceMapBinding
	head            map[string]SourceMapBinding
}

// Build constructs a complete source-backed change manifest. Change derivation
// remains the detector's responsibility; this adapter emits the full resolved
// registered inventory and exact before/after evidence only.
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
		return failResult(err)
	}
	if input.Before.RegistryDigest != authority.registryDigest || input.Head.RegistryDigest != authority.registryDigest ||
		input.Before.SourceMapDigest != authority.sourceMapDigest || input.Head.SourceMapDigest != authority.sourceMapDigest {
		return unknownResult(CodeAuthorityDrift, "snapshot and registry/source-map digests do not agree", input.Authority)
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
		return failResult(err)
	}
	if err := rejectUnregistered(head, authority.inventory); err != nil {
		return failResult(err)
	}
	if err := rejectUnobservedBindings(before, authority.before); err != nil {
		return resultForResolution(err, input.Authority)
	}
	if err := rejectUnobservedBindings(head, authority.head); err != nil {
		return resultForResolution(err, input.Authority)
	}
	resolvedBefore, err := resolveSide(before, authority.before, authority.inventory)
	if err != nil {
		return resultForResolution(err, input.Authority)
	}
	resolvedHead, err := resolveSide(head, authority.head, authority.inventory)
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
			SurfaceID: registered.SurfaceID, CodeSymbolID: registered.CodeSymbolID,
			SemanticOwnerID: registered.SemanticOwnerID, SourceMapID: registered.SourceMapID,
		}
		if value, ok := resolvedBefore[registered.SemanticOwnerID]; ok {
			entry.BeforePresent = true
			entry.BeforeBindingDigest = value.Observed.BindingDigest
			entry.BeforeSourceMapBindingDigest = authority.before[registered.SemanticOwnerID].SourceMapBindingDigest
			entry.BeforeBlobDigest = value.Observed.BlobDigest
		}
		if value, ok := resolvedHead[registered.SemanticOwnerID]; ok {
			entry.AfterPresent = true
			entry.AfterBindingDigest = value.Observed.BindingDigest
			entry.AfterSourceMapBindingDigest = authority.head[registered.SemanticOwnerID].SourceMapBindingDigest
			entry.AfterBlobDigest = value.Observed.BlobDigest
		}
		entries = append(entries, entry)
	}
	resolvedIDs := append(make([]string, 0, len(surfaceIDs)), surfaceIDs...)
	counts := ComponentCounts{Registered: len(surfaceIDs), Before: len(resolvedBefore), Head: len(resolvedHead), Resolved: len(entries)}
	manifest := Manifest{
		Schema: SchemaV1, Status: StatusComplete, ObservationComplete: true,
		BeforeSnapshotDigest: input.Before.Digest, HeadSnapshotDigest: input.Head.Digest,
		RegistryDigest: authority.registryDigest, SourceMapDigest: authority.sourceMapDigest,
		ToolchainDigest: authority.toolchainDigest, ProfileDigest: authority.profileDigest,
		ResolvedSurfaceIDs: resolvedIDs, Entries: entries, Counts: counts,
		Work: Work{ComponentCount: len(entries), WorkUnits: len(entries)},
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
	if input.RegistryDigest == "" || input.SourceMapDigest == "" || input.ToolchainDigest == "" || input.ProfileDigest == "" || input.Inventory == nil || input.Before == nil || input.Head == nil {
		return normalizedAuthority{}, unknownError(CodeMissingAuthority, "registry/source-map, toolchain/profile, inventory, and both side bindings are required")
	}
	if input.Schema != "" && input.Schema != AuthoritySchemaV1 {
		return normalizedAuthority{}, failError(CodeInvalidSchema, "authority schema %q is not %q", input.Schema, AuthoritySchemaV1)
	}
	for value, label := range map[string]string{input.RegistryDigest: "registry digest", input.SourceMapDigest: "source-map digest", input.ToolchainDigest: "toolchain digest", input.ProfileDigest: "profile digest"} {
		if !validDigest(value) {
			return normalizedAuthority{}, failError(CodeMalformedBinding, "%s is malformed", label)
		}
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
	before, err := normalizeBindings(input.Before, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	head, err := normalizeBindings(input.Head, inventory)
	if err != nil {
		return normalizedAuthority{}, err
	}
	return normalizedAuthority{registryDigest: input.RegistryDigest, sourceMapDigest: input.SourceMapDigest, toolchainDigest: input.ToolchainDigest, profileDigest: input.ProfileDigest, inventory: inventory, before: before, head: head}, nil
}

func normalizeInventory(values []Surface) (map[string]Surface, error) {
	result := make(map[string]Surface, len(values))
	seenSymbols := make(map[string]string, len(values))
	seenOwners := make(map[string]string, len(values))
	seenMaps := make(map[string]string, len(values))
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
		if previous, exists := seenMaps[normalized.SourceMapID]; exists {
			return nil, failError(CodeConflictingBinding, "source-map ID %q resolves to both %q and %q", normalized.SourceMapID, previous, normalized.SurfaceID)
		}
		result[normalized.SurfaceID] = normalized
		seenSymbols[normalized.CodeSymbolID] = normalized.SurfaceID
		seenOwners[normalized.SemanticOwnerID] = normalized.SurfaceID
		seenMaps[normalized.SourceMapID] = normalized.SurfaceID
	}
	return result, nil
}

func normalizeSurface(value Surface) (Surface, error) {
	fields := []struct {
		value string
		name  string
	}{
		{value.SurfaceID, "surface ID"}, {value.CodeSymbolID, "code symbol ID"}, {value.SemanticOwnerID, "semantic owner ID"}, {value.SourceMapID, "source-map ID"},
	}
	for _, field := range fields {
		if _, err := normalizeID(field.value); err != nil {
			return Surface{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	return value, nil
}

func normalizeBindings(values []SourceMapBinding, inventory map[string]Surface) (map[string]SourceMapBinding, error) {
	result := make(map[string]SourceMapBinding, len(values))
	seenSurfaces := make(map[string]struct{}, len(values))
	seenSymbols := make(map[string]struct{}, len(values))
	seenMaps := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized, err := normalizeBinding(value)
		if err != nil {
			return nil, err
		}
		registered, ok := inventory[normalized.SurfaceID]
		if !ok {
			return nil, failError(CodeMalformedBinding, "surface ID %q is not registered", normalized.SurfaceID)
		}
		if normalized.CodeSymbolID != registered.CodeSymbolID || normalized.SemanticOwnerID != registered.SemanticOwnerID || normalized.SourceMapID != registered.SourceMapID {
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

func normalizeBinding(value SourceMapBinding) (SourceMapBinding, error) {
	fields := []struct {
		value string
		name  string
	}{
		{value.SurfaceID, "surface ID"}, {value.CodeSymbolID, "code symbol ID"}, {value.SemanticOwnerID, "semantic owner ID"}, {value.SourceMapID, "source-map ID"},
	}
	for _, field := range fields {
		if _, err := normalizeID(field.value); err != nil {
			return SourceMapBinding{}, failError(CodeMalformedBinding, "%s: %v", field.name, err)
		}
	}
	if !validRole(value.Role) {
		return SourceMapBinding{}, failError(CodeMalformedBinding, "surface %q has invalid role %q", value.SurfaceID, value.Role)
	}
	repoPath, err := normalizeRepoPath(value.Path)
	if err != nil {
		return SourceMapBinding{}, failError(CodeMalformedBinding, "surface %q path: %v", value.SurfaceID, err)
	}
	if !validDigest(value.BlobDigest) || !validRawDigest(value.BindingDigest) || !validRawDigest(value.SourceMapBindingDigest) {
		return SourceMapBinding{}, failError(CodeMalformedBinding, "surface %q has malformed source digest", value.SurfaceID)
	}
	value.Path = repoPath
	return value, nil
}

func snapshotIndex(snapshot selectiveci.Snapshot) (map[string]observed, error) {
	result := make(map[string]observed)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			if _, exists := result[binding.ID]; exists {
				return nil, fmt.Errorf("snapshot semantic owner ID %q occurs more than once", binding.ID)
			}
			result[binding.ID] = observed{OwnerID: binding.ID, Role: binding.Role, Path: source.Path, BlobDigest: source.BlobDigest, BindingDigest: binding.BindingDigest}
		}
	}
	return result, nil
}

func rejectUnregistered(values map[string]observed, inventory map[string]Surface) error {
	for ownerID := range values {
		found := false
		for _, surface := range inventory {
			if surface.SemanticOwnerID == ownerID {
				found = true
				break
			}
		}
		if !found {
			return failError(CodeMalformedBinding, "semantic owner ID %q is not registered", ownerID)
		}
	}
	return nil
}

func resolveSide(snapshot map[string]observed, bindings map[string]SourceMapBinding, inventory map[string]Surface) (map[string]resolved, error) {
	result := make(map[string]resolved, len(snapshot))
	for ownerID, value := range snapshot {
		binding, ok := bindings[ownerID]
		if !ok {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has no exact source-map binding", ownerID)
		}
		registered := inventory[binding.SurfaceID]
		if registered.SemanticOwnerID != ownerID || binding.Path != value.Path || binding.Role != value.Role || binding.BlobDigest != value.BlobDigest || binding.BindingDigest != value.BindingDigest {
			return nil, unknownError(CodeUnknownChangedSurface, "semantic owner ID %q has a stale or conflicting source-map binding", ownerID)
		}
		result[ownerID] = resolved{Surface: registered, Observed: value}
	}
	return result, nil
}

func rejectUnobservedBindings(snapshot map[string]observed, bindings map[string]SourceMapBinding) error {
	for ownerID := range bindings {
		if _, observed := snapshot[ownerID]; !observed {
			return unknownError(CodeUnknownChangedSurface, "source-map binding for semantic owner ID %q has no matching snapshot observation", ownerID)
		}
	}
	return nil
}

func requireInventoryCoverage(inventory map[string]Surface, before, head map[string]resolved) error {
	for _, surface := range inventory {
		if _, beforeOK := before[surface.SemanticOwnerID]; !beforeOK {
			if _, headOK := head[surface.SemanticOwnerID]; !headOK {
				return unknownError(CodeUnknownChangedSurface, "registered surface %q has no before or head observation", surface.SurfaceID)
			}
		}
	}
	return nil
}

func sortedSurfaceIDs(values map[string]Surface) []string {
	result := make([]string, 0, len(values))
	for surfaceID := range values {
		result = append(result, surfaceID)
	}
	sort.Strings(result)
	return result
}

func normalizeID(value string) (string, error) {
	if value == "" || value != strings.TrimSpace(value) {
		return "", fmt.Errorf("ID is empty or padded")
	}
	id, err := semantic.ParseIdentity(value)
	if err != nil || id.String() != value {
		return "", fmt.Errorf("ID %q is not canonical", value)
	}
	return value, nil
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

func validDigest(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+sha256.Size*2 || !strings.HasPrefix(value, prefix) {
		return false
	}
	_, err := hex.DecodeString(value[len(prefix):])
	return err == nil && value == strings.ToLower(value)
}

func validRawDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func resultForResolution(err error, authority RegistrySourceMap) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		return unknownResult(CodeUnknownChangedSurface, err.Error(), authority)
	}
	if typed.Status == StatusFailClosed {
		return failResult(typed)
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
	manifest := Manifest{Schema: SchemaV1, Status: StatusUnknown, FullSuiteRequired: true, ReasonCode: code, ResolvedSurfaceIDs: []string{}, Entries: []ManifestEntry{}}
	if validDigest(authority.RegistryDigest) {
		manifest.RegistryDigest = authority.RegistryDigest
	}
	if validDigest(authority.SourceMapDigest) {
		manifest.SourceMapDigest = authority.SourceMapDigest
	}
	if validDigest(authority.ToolchainDigest) {
		manifest.ToolchainDigest = authority.ToolchainDigest
	}
	if validDigest(authority.ProfileDigest) {
		manifest.ProfileDigest = authority.ProfileDigest
	}
	return sealResult(manifest, unknownError(code, "%s", message))
}

func failResult(err error) (Manifest, error) {
	typed, ok := err.(*Error)
	if !ok {
		typed = failError(CodeInvalidManifest, "%v", err)
	}
	manifest := Manifest{Schema: SchemaV1, Status: StatusFailClosed, FullSuiteRequired: true, ReasonCode: typed.Code, ResolvedSurfaceIDs: []string{}, Entries: []ManifestEntry{}}
	return sealResult(manifest, typed)
}
