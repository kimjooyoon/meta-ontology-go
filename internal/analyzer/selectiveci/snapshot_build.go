package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func fail(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), FullSuiteFallback: true}
}

// NewSnapshot constructs a canonical source-backed snapshot from explicit
// semanticbinding records.
func NewSnapshot(input SnapshotInput) (Snapshot, error) { return Build(input) }

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
		Schema: SchemaV1, Status: StatusBound,
		SourceMapDigest: sourceMapDigest, RegistryDigest: registryDigest,
		RegisteredIDs: sortedIDs(registered), Sources: sources,
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
		return Binding{}, fail(CodeAmbiguousBinding, "binding %q is attached to %q, not %q", binding.ID, spanPath, repoPath)
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

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func unknownSnapshot(err error) (Snapshot, error) {
	return Snapshot{Schema: SchemaV1, Status: StatusUnknown, FullSuiteFallback: true}, err
}
