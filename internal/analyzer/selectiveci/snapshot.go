// Package selectiveci provides deterministic, source-backed selective-CI
// snapshot and change-binding operations.
package selectiveci

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	// SchemaV1 is the versioned canonical JSON schema for Snapshot.
	SchemaV1 = "gooo/selective-ci-snapshot/v1"

	// FormatVersion is an alias for callers that use format vocabulary.
	FormatVersion = SchemaV1
)

// Status is the state of a source-backed selective-CI result.
type Status string

const (
	StatusBound   Status = "BOUND"
	StatusUnknown Status = "UNKNOWN"
)

// ErrorCode identifies a fail-closed construction or validation failure.
type ErrorCode string

const (
	CodeInput             ErrorCode = "selectiveci.input"
	CodeMissingBinding    ErrorCode = "selectiveci.missing-binding"
	CodeAmbiguousBinding  ErrorCode = "selectiveci.ambiguous-binding"
	CodeUnregisteredID    ErrorCode = "selectiveci.unregistered-id"
	CodeDuplicateBinding  ErrorCode = "selectiveci.duplicate-binding"
	CodeMalformedPath     ErrorCode = "selectiveci.malformed-path"
	CodeMalformedDigest   ErrorCode = "selectiveci.malformed-digest"
	CodeCandidateIdentity ErrorCode = "selectiveci.candidate-identity"
	CodeDerivedIdentity   ErrorCode = "selectiveci.derived-identity"
	CodeStaleSnapshot     ErrorCode = "selectiveci.stale-snapshot"
	CodeInvalidBinding    ErrorCode = "selectiveci.invalid-binding"
	CodeInvalidStatus     ErrorCode = "selectiveci.invalid-status"
	CodeInvalidSchema     ErrorCode = "selectiveci.invalid-schema"
)

// Error is a deterministic fail-closed result. A construction or diff error
// always requires the caller to use its full-suite fallback.
type Error struct {
	Code              ErrorCode
	Message           string
	FullSuiteFallback bool
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return string(e.Code) + ": " + e.Message
}

func fail(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), FullSuiteFallback: true}
}

// SourceInput is one explicit repository source and its semanticbinding
// records. Path is a repository-relative source-map path; no name lookup is
// performed to associate a binding with a source.
type SourceInput struct {
	Path       string
	BlobDigest string
	Bindings   []semanticbinding.Binding
}

// SnapshotInput contains all authority needed to construct a Snapshot. A nil
// RegisteredIDs slice is deliberately different from an empty registry: the
// former means registry membership was not supplied and is UNKNOWN.
type SnapshotInput struct {
	Sources         []SourceInput
	SourceMapDigest string
	RegistryDigest  string
	RegisteredIDs   []string

	// CandidateBindings and DerivedBindings are observations, not authoritative
	// semantic bindings. Supplying either is rejected rather than promoted or
	// silently ignored.
	CandidateBindings []semanticbinding.Binding
	DerivedBindings   []semanticbinding.Binding
}

// Input is the concise spelling of SnapshotInput.
type Input = SnapshotInput

// ManifestInput and SnapshotManifest are vocabulary aliases for integrations
// that call the persisted object a manifest.
type ManifestInput = SnapshotInput

// Binding is the explicit, source-bound manifest representation of a
// semanticbinding.Binding.
type Binding struct {
	ID            string               `json:"id"`
	Role          semanticbinding.Role `json:"role"`
	Status        Status               `json:"status"`
	BindingDigest string               `json:"binding_digest"`
}

// Source is one canonical manifest source record.
type Source struct {
	Path       string    `json:"path"`
	BlobDigest string    `json:"blob_digest"`
	Bindings   []Binding `json:"bindings"`
}

// Snapshot is a canonical, source-backed manifest. A BOUND snapshot always
// has a valid Digest that covers every field except Digest itself.
type Snapshot struct {
	Schema            string   `json:"schema"`
	Status            Status   `json:"status"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
	SourceMapDigest   string   `json:"source_map_digest"`
	RegistryDigest    string   `json:"registry_digest"`
	RegisteredIDs     []string `json:"registered_ids"`
	Sources           []Source `json:"sources"`
	Digest            string   `json:"digest"`
}

// SnapshotManifest is the persisted-manifest spelling of Snapshot.
type SnapshotManifest = Snapshot

// SourceRecord and BindingRecord are explicit manifest vocabulary aliases.
type SourceRecord = Source
type BindingRecord = Binding

// Delta is the source-backed union of stable IDs whose exact binding changed.
// ChangedIDs is empty for an exact replay and also for every UNKNOWN result.
type Delta struct {
	Status            Status   `json:"status"`
	ChangedIDs        []string `json:"changed_ids"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
}

// NewSnapshot constructs a canonical source-backed snapshot from explicit
// semanticbinding records. It returns an UNKNOWN snapshot and an Error for
// every incomplete or conflicting authority input.
func NewSnapshot(input SnapshotInput) (Snapshot, error) {
	return Build(input)
}

// Build is the primary Snapshot constructor.
func Build(input SnapshotInput) (Snapshot, error) {
	if input.RegisteredIDs == nil {
		return unknownSnapshot(fail(CodeInput, "registered IDs are required"))
	}
	if len(input.CandidateBindings) != 0 {
		return unknownSnapshot(fail(CodeCandidateIdentity, "candidate-only bindings cannot become authoritative"))
	}
	if len(input.DerivedBindings) != 0 {
		return unknownSnapshot(fail(CodeDerivedIdentity, "derived-only bindings cannot become authoritative"))
	}

	registered, err := normalizeRegisteredIDs(input.RegisteredIDs)
	if err != nil {
		return unknownSnapshot(err)
	}
	sourceMapDigest, err := normalizeDigest(input.SourceMapDigest, "source-map digest")
	if err != nil {
		return unknownSnapshot(err)
	}
	registryDigest, err := normalizeDigest(input.RegistryDigest, "registry digest")
	if err != nil {
		return unknownSnapshot(err)
	}

	sources, err := normalizeSources(input.Sources, registered)
	if err != nil {
		return unknownSnapshot(err)
	}
	result := Snapshot{
		Schema:            SchemaV1,
		Status:            StatusBound,
		FullSuiteFallback: false,
		SourceMapDigest:   sourceMapDigest,
		RegistryDigest:    registryDigest,
		RegisteredIDs:     sortedIDs(registered),
		Sources:           sources,
	}
	unsigned, err := result.unsignedJSON()
	if err != nil {
		return unknownSnapshot(err)
	}
	result.Digest = digest(unsigned)
	return result, nil
}

// BuildSnapshot is the descriptive spelling of Build.
func BuildSnapshot(input SnapshotInput) (Snapshot, error) { return Build(input) }

// CanonicalJSON returns the strict canonical JSON representation of the
// snapshot, including and verifying its digest.
func (s Snapshot) CanonicalJSON() ([]byte, error) {
	normalized, err := normalizeSnapshot(s)
	if err != nil {
		return nil, err
	}
	unsigned, err := normalized.unsignedJSON()
	if err != nil {
		return nil, err
	}
	want := digest(unsigned)
	if normalized.Digest != want {
		return nil, fail(CodeStaleSnapshot, "snapshot digest %q does not match canonical content %q", normalized.Digest, want)
	}
	return json.Marshal(wireForSnapshot(normalized))
}

// StableHash returns the digest bound into the canonical snapshot.
func (s Snapshot) StableHash() string { return s.Digest }

// Validate checks schema, source bindings, registry membership, and the
// digest-bound canonical representation without consulting external state.
func (s Snapshot) Validate() error {
	_, err := s.CanonicalJSON()
	return err
}

// MarshalJSON makes ordinary encoding/json use the strict canonical form.
func (s Snapshot) MarshalJSON() ([]byte, error) { return s.CanonicalJSON() }

// UnmarshalJSON accepts only the exact canonical JSON representation.
func (s *Snapshot) UnmarshalJSON(data []byte) error {
	parsed, err := DecodeSnapshot(data)
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// DecodeSnapshot decodes and verifies a strict canonical JSON snapshot.
func DecodeSnapshot(data []byte) (Snapshot, error) {
	var wire snapshotWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return Snapshot{}, fail(CodeInvalidSchema, "decode snapshot JSON: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Snapshot{}, fail(CodeInvalidSchema, "snapshot JSON has trailing values")
		}
		return Snapshot{}, fail(CodeInvalidSchema, "decode snapshot JSON after object: %v", err)
	}
	snapshot := Snapshot{
		Schema:            wire.Schema,
		Status:            wire.Status,
		FullSuiteFallback: wire.FullSuiteFallback,
		SourceMapDigest:   wire.SourceMapDigest,
		RegistryDigest:    wire.RegistryDigest,
		RegisteredIDs:     wire.RegisteredIDs,
		Sources:           wire.Sources,
		Digest:            wire.Digest,
	}
	canonical, err := snapshot.CanonicalJSON()
	if err != nil {
		return Snapshot{}, err
	}
	if !bytes.Equal(canonical, data) {
		return Snapshot{}, fail(CodeInvalidSchema, "snapshot JSON is not canonical")
	}
	return snapshot, nil
}

// Diff computes the union of base and head stable IDs for changed, deleted,
// added, or relocated exact source bindings. It is order-independent and
// returns no IDs on UNKNOWN.
func Diff(base, head Snapshot) (Delta, error) {
	left, err := normalizeSnapshot(base)
	if err != nil {
		return unknownDelta(err)
	}
	right, err := normalizeSnapshot(head)
	if err != nil {
		return unknownDelta(err)
	}

	leftBindings := bindingIndex(left)
	rightBindings := bindingIndex(right)
	if left.SourceMapDigest != right.SourceMapDigest || left.RegistryDigest != right.RegistryDigest {
		ids := unionBindingIDs(leftBindings, rightBindings)
		return Delta{Status: StatusBound, ChangedIDs: ids}, nil
	}

	ids := unionBindingIDs(leftBindings, rightBindings)
	changed := make([]string, 0, len(ids))
	for _, id := range ids {
		before, beforeOK := leftBindings[id]
		after, afterOK := rightBindings[id]
		if !beforeOK || !afterOK || before != after {
			changed = append(changed, id)
		}
	}
	return Delta{Status: StatusBound, ChangedIDs: changed}, nil
}

// DiffSnapshots is an explicit alias for callers that distinguish manifests
// from other snapshots in their own code.
func DiffSnapshots(base, head Snapshot) (Delta, error) { return Diff(base, head) }

// CanonicalJSON returns the canonical JSON encoding of a delta. Delta is not a
// persisted authority object, so no digest is embedded.
func (d Delta) CanonicalJSON() ([]byte, error) {
	if d.Status != StatusBound && d.Status != StatusUnknown {
		return nil, fail(CodeInvalidStatus, "unknown delta status %q", d.Status)
	}
	ids := append([]string(nil), d.ChangedIDs...)
	sort.Strings(ids)
	if d.Status == StatusUnknown && len(ids) != 0 {
		return nil, fail(CodeInvalidStatus, "UNKNOWN delta cannot contain partial IDs")
	}
	type wireDelta struct {
		Status            Status   `json:"status"`
		ChangedIDs        []string `json:"changed_ids"`
		FullSuiteFallback bool     `json:"full_suite_fallback"`
	}
	return json.Marshal(wireDelta{Status: d.Status, ChangedIDs: ids, FullSuiteFallback: d.FullSuiteFallback})
}

func (d Delta) IsEmpty() bool { return d.Status == StatusBound && len(d.ChangedIDs) == 0 }

type snapshotWire struct {
	Schema            string   `json:"schema"`
	Status            Status   `json:"status"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
	SourceMapDigest   string   `json:"source_map_digest"`
	RegistryDigest    string   `json:"registry_digest"`
	RegisteredIDs     []string `json:"registered_ids"`
	Sources           []Source `json:"sources"`
	Digest            string   `json:"digest"`
}

func wireForSnapshot(s Snapshot) snapshotWire {
	return snapshotWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources, Digest: s.Digest,
	}
}

func (s Snapshot) unsignedJSON() ([]byte, error) {
	type unsignedWire struct {
		Schema            string   `json:"schema"`
		Status            Status   `json:"status"`
		FullSuiteFallback bool     `json:"full_suite_fallback"`
		SourceMapDigest   string   `json:"source_map_digest"`
		RegistryDigest    string   `json:"registry_digest"`
		RegisteredIDs     []string `json:"registered_ids"`
		Sources           []Source `json:"sources"`
	}
	return json.Marshal(unsignedWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources,
	})
}

func normalizeSnapshot(s Snapshot) (Snapshot, error) {
	if s.Schema != SchemaV1 {
		return Snapshot{}, fail(CodeInvalidSchema, "schema %q is not %q", s.Schema, SchemaV1)
	}
	if s.Status != StatusBound || s.FullSuiteFallback || s.Digest == "" {
		return Snapshot{}, fail(CodeInvalidStatus, "snapshot must be BOUND with a digest")
	}
	if s.RegisteredIDs == nil || s.Sources == nil {
		return Snapshot{}, fail(CodeInvalidSchema, "registered IDs and sources must be explicit arrays")
	}
	registered, err := normalizeRegisteredIDs(s.RegisteredIDs)
	if err != nil {
		return Snapshot{}, err
	}
	sourceMapDigest, err := normalizeDigest(s.SourceMapDigest, "source-map digest")
	if err != nil {
		return Snapshot{}, err
	}
	registryDigest, err := normalizeDigest(s.RegistryDigest, "registry digest")
	if err != nil {
		return Snapshot{}, err
	}
	sources, err := normalizeManifestSources(s.Sources, registered)
	if err != nil {
		return Snapshot{}, err
	}
	if !validDigest(s.Digest) {
		return Snapshot{}, fail(CodeMalformedDigest, "snapshot digest is malformed")
	}
	s.RegisteredIDs = sortedIDs(registered)
	s.SourceMapDigest = sourceMapDigest
	s.RegistryDigest = registryDigest
	s.Sources = sources
	unsigned, err := s.unsignedJSON()
	if err != nil {
		return Snapshot{}, err
	}
	if digest(unsigned) != s.Digest {
		return Snapshot{}, fail(CodeStaleSnapshot, "snapshot digest does not match canonical content")
	}
	return s, nil
}

func normalizeSources(inputs []SourceInput, registered map[string]struct{}) ([]Source, error) {
	sources := make([]Source, 0, len(inputs))
	seenPaths := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		repoPath, err := normalizeRepoPath(input.Path)
		if err != nil {
			return nil, err
		}
		if _, exists := seenPaths[repoPath]; exists {
			return nil, fail(CodeDuplicateBinding, "source path %q is duplicated", repoPath)
		}
		seenPaths[repoPath] = struct{}{}
		blobDigest, err := normalizeDigest(input.BlobDigest, "blob digest")
		if err != nil {
			return nil, err
		}
		if len(input.Bindings) == 0 {
			return nil, fail(CodeMissingBinding, "source %q has no explicit semantic binding", repoPath)
		}
		bindings := make([]Binding, 0, len(input.Bindings))
		for _, binding := range input.Bindings {
			record, err := normalizeBinding(binding, repoPath, registered)
			if err != nil {
				return nil, err
			}
			bindings = append(bindings, record)
		}
		sort.Slice(bindings, func(i, j int) bool { return compareBindings(bindings[i], bindings[j]) < 0 })
		sources = append(sources, Source{Path: repoPath, BlobDigest: blobDigest, Bindings: bindings})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	if err := rejectDuplicateManifestIDs(sources); err != nil {
		return nil, err
	}
	return sources, nil
}

func normalizeManifestSources(inputs []Source, registered map[string]struct{}) ([]Source, error) {
	sources := make([]Source, 0, len(inputs))
	seenPaths := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		repoPath, err := normalizeRepoPath(input.Path)
		if err != nil {
			return nil, err
		}
		if _, exists := seenPaths[repoPath]; exists {
			return nil, fail(CodeDuplicateBinding, "source path %q is duplicated", repoPath)
		}
		seenPaths[repoPath] = struct{}{}
		blobDigest, err := normalizeDigest(input.BlobDigest, "blob digest")
		if err != nil {
			return nil, err
		}
		if len(input.Bindings) == 0 {
			return nil, fail(CodeMissingBinding, "source %q has no explicit semantic binding", repoPath)
		}
		bindings := append([]Binding(nil), input.Bindings...)
		for index := range bindings {
			binding, err := normalizeManifestBinding(bindings[index], repoPath, registered)
			if err != nil {
				return nil, err
			}
			bindings[index] = binding
		}
		sort.Slice(bindings, func(i, j int) bool { return compareBindings(bindings[i], bindings[j]) < 0 })
		sources = append(sources, Source{Path: repoPath, BlobDigest: blobDigest, Bindings: bindings})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	if err := rejectDuplicateManifestIDs(sources); err != nil {
		return nil, err
	}
	return sources, nil
}

func normalizeBinding(binding semanticbinding.Binding, repoPath string, registered map[string]struct{}) (Binding, error) {
	spanPath, err := normalizeRepoPath(binding.Span.Filename)
	if err != nil {
		return Binding{}, fail(CodeMissingBinding, "binding %q has no valid source span: %v", binding.ID, err)
	}
	if spanPath != repoPath {
		return Binding{}, fail(CodeAmbiguousBinding, "binding %q is explicitly attached to %q, not %q", binding.ID, spanPath, repoPath)
	}
	if !validRole(binding.Role) {
		return Binding{}, fail(CodeInvalidBinding, "binding %q has invalid role %q", binding.ID, binding.Role)
	}
	id, err := normalizeID(binding.ID)
	if err != nil {
		return Binding{}, err
	}
	if _, ok := registered[id]; !ok {
		return Binding{}, fail(CodeUnregisteredID, "binding ID %q is not in the explicit registry", id)
	}
	expected := binding.StableHash()
	if !validRawDigest(binding.Digest) || binding.Digest != expected || binding.CanonicalDigest != expected {
		return Binding{}, fail(CodeStaleSnapshot, "binding %q digest does not match its explicit semantic record", id)
	}
	return Binding{ID: id, Role: binding.Role, Status: StatusBound, BindingDigest: expected}, nil
}

func normalizeManifestBinding(binding Binding, repoPath string, registered map[string]struct{}) (Binding, error) {
	if binding.Status != StatusBound {
		return Binding{}, fail(CodeInvalidStatus, "binding %q is not BOUND", binding.ID)
	}
	id, err := normalizeID(binding.ID)
	if err != nil {
		return Binding{}, err
	}
	if _, ok := registered[id]; !ok {
		return Binding{}, fail(CodeUnregisteredID, "binding ID %q is not in the explicit registry", id)
	}
	if !validRole(binding.Role) {
		return Binding{}, fail(CodeInvalidBinding, "binding %q has invalid role %q", id, binding.Role)
	}
	if !validRawDigest(binding.BindingDigest) {
		return Binding{}, fail(CodeMalformedDigest, "binding %q digest is malformed", id)
	}
	return Binding{ID: id, Role: binding.Role, Status: StatusBound, BindingDigest: binding.BindingDigest}, nil
}

func rejectDuplicateManifestIDs(sources []Source) error {
	seen := make(map[string]string)
	for _, source := range sources {
		for _, binding := range source.Bindings {
			if previous, exists := seen[binding.ID]; exists {
				if previous == source.Path {
					return fail(CodeDuplicateBinding, "binding ID %q is duplicated in source %q", binding.ID, source.Path)
				}
				return fail(CodeAmbiguousBinding, "binding ID %q is bound by %q and %q", binding.ID, previous, source.Path)
			}
			seen[binding.ID] = source.Path
		}
	}
	return nil
}

func normalizeRegisteredIDs(values []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		id, err := normalizeID(value)
		if err != nil {
			return nil, err
		}
		if _, exists := result[id]; exists {
			return nil, fail(CodeDuplicateBinding, "registered ID %q is duplicated", id)
		}
		result[id] = struct{}{}
	}
	return result, nil
}

func sortedIDs(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizeID(raw string) (string, error) {
	if raw != strings.TrimSpace(raw) {
		return "", fail(CodeInvalidBinding, "semantic ID is padded")
	}
	id, err := semantic.ParseIdentity(raw)
	if err != nil {
		return "", fail(CodeInvalidBinding, "semantic ID %q is invalid: %v", raw, err)
	}
	if id.String() != raw {
		return "", fail(CodeInvalidBinding, "semantic ID %q is not canonical", raw)
	}
	return raw, nil
}

func normalizeRepoPath(raw string) (string, error) {
	if raw == "" || raw != strings.TrimSpace(raw) || strings.Contains(raw, "\\") {
		return "", fail(CodeMalformedPath, "repository path %q is malformed", raw)
	}
	if strings.HasPrefix(raw, "/") || path.IsAbs(raw) || (len(raw) >= 2 && raw[1] == ':') {
		return "", fail(CodeMalformedPath, "repository path %q must be relative", raw)
	}
	clean := path.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fail(CodeMalformedPath, "repository path %q escapes the repository", raw)
	}
	return clean, nil
}

func normalizeDigest(raw, label string) (string, error) {
	if !validDigest(raw) {
		return "", fail(CodeMalformedDigest, "%s %q is not a lowercase sha256 digest", label, raw)
	}
	return raw, nil
}

func validDigest(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+sha256.Size*2 || !strings.HasPrefix(value, prefix) {
		return false
	}
	for _, char := range value[len(prefix):] {
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func validRawDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func validRole(role semanticbinding.Role) bool {
	switch role {
	case semanticbinding.RoleHandwrittenImpl, semanticbinding.RoleGeneratedImpl, semanticbinding.RoleAdapter:
		return true
	default:
		return false
	}
}

func compareBindings(left, right Binding) int {
	if value := strings.Compare(left.ID, right.ID); value != 0 {
		return value
	}
	if value := strings.Compare(string(left.Role), string(right.Role)); value != 0 {
		return value
	}
	return strings.Compare(left.BindingDigest, right.BindingDigest)
}

type indexedBinding struct {
	Path          string
	BlobDigest    string
	ID            string
	Role          semanticbinding.Role
	Status        Status
	BindingDigest string
}

func bindingIndex(snapshot Snapshot) map[string]indexedBinding {
	result := make(map[string]indexedBinding)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			result[binding.ID] = indexedBinding{
				Path: source.Path, BlobDigest: source.BlobDigest, ID: binding.ID,
				Role: binding.Role, Status: binding.Status, BindingDigest: binding.BindingDigest,
			}
		}
	}
	return result
}

func unionBindingIDs(left, right map[string]indexedBinding) []string {
	set := make(map[string]struct{}, len(left)+len(right))
	for id := range left {
		set[id] = struct{}{}
	}
	for id := range right {
		set[id] = struct{}{}
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func unknownSnapshot(err error) (Snapshot, error) {
	return Snapshot{Schema: SchemaV1, Status: StatusUnknown, FullSuiteFallback: true}, err
}

func unknownDelta(err error) (Delta, error) {
	return Delta{Status: StatusUnknown, FullSuiteFallback: true}, err
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
