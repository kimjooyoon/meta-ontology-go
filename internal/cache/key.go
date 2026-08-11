package cache

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"sort"
	"strings"
)

const (
	// DefaultKeyVersion is the current cache-key schema.
	DefaultKeyVersion    = ProjectionKeyVersion
	ProjectionKeyVersion = "v2"
	keyDomain            = "gooo-projection-key\x00"
)

// ProjectionKeySpec is the typed identity of one reconstructable projection.
// Every field affects the content address; metadata cannot change identity
// after a key has been created.
type ProjectionKeySpec struct {
	Domain                string
	Namespace             string
	Version               string
	ArtifactKind          string
	Projection            string
	HostStage             HostStage
	SemanticClosureDigest Digest
	DependencyRoot        Digest
	PolicySchemaDigest    Digest
	Toolchain             string
	ToolVersion           string
	Target                string
	BuildTags             []string
	Options               any
}

// KeySpec is the compatibility constructor for callers that provide source
// inputs and freshness records instead of precomputed typed digests.
type KeySpec struct {
	Version      string
	Domain       string
	Namespace    string
	ArtifactKind string
	Projection   string
	ToolVersion  string
	Toolchain    string
	Target       string
	BuildTags    []string
	HostStage    HostStage
	Inputs       any
	Options      any
	Freshness    FreshnessSpec
}

// ProjectionKey is the content address of one projection. Its fields are
// comparable typed identity, so callers can safely use it as a test value.
type ProjectionKey struct {
	Digest                Digest
	Domain                string
	Namespace             string
	Version               string
	ArtifactKind          string
	Projection            string
	HostStage             HostStage
	SemanticClosureDigest Digest
	DependencyRoot        Digest
	PolicySchemaDigest    Digest
	Toolchain             string
	ToolVersion           string
	Target                string
	BuildTagsDigest       Digest
	OptionsDigest         Digest
	InputDigest           Digest
	DependencyDigest      Digest
	ProvenanceDigest      Digest
}

// Key is retained as the cache API name while exposing the ProjectionKey v2
// identity contract.
type Key = ProjectionKey

// NewProjectionKey creates a v2 key from already typed identity digests.
func NewProjectionKey(spec ProjectionKeySpec) (ProjectionKey, error) {
	normalized, err := normalizeProjectionSpec(spec)
	if err != nil {
		return ProjectionKey{}, err
	}
	buildTagsDigest, err := digestBuildTags(normalized.BuildTags)
	if err != nil {
		return ProjectionKey{}, err
	}
	optionsDigest, err := DigestOf(normalized.Options)
	if err != nil {
		return ProjectionKey{}, fmt.Errorf("hash options: %w", err)
	}
	key := ProjectionKey{
		Domain: normalized.Domain, Namespace: normalized.Domain, Version: normalized.Version,
		ArtifactKind: normalized.ArtifactKind, Projection: normalized.Projection,
		HostStage: normalized.HostStage, SemanticClosureDigest: normalized.SemanticClosureDigest,
		DependencyRoot: normalized.DependencyRoot, PolicySchemaDigest: normalized.PolicySchemaDigest,
		Toolchain: normalized.Toolchain, ToolVersion: normalized.Toolchain, Target: normalized.Target,
		BuildTagsDigest: buildTagsDigest, OptionsDigest: optionsDigest,
		InputDigest: normalized.SemanticClosureDigest, DependencyDigest: normalized.DependencyRoot,
		ProvenanceDigest: normalized.PolicySchemaDigest,
	}
	key.Digest = digestForProjectionKey(key)
	return key, nil
}

// NewKey hashes compatibility inputs into the v2 typed identity.
func NewKey(spec KeySpec) (Key, error) {
	version := spec.Version
	if version == "" {
		version = DefaultKeyVersion
	}
	domain := spec.Domain
	if domain == "" {
		domain = spec.Namespace
	}
	toolchain := spec.Toolchain
	if toolchain == "" {
		toolchain = spec.ToolVersion
	}
	if err := validateKeyComponent("domain", domain, true); err != nil {
		return Key{}, err
	}
	if spec.HostStage != "" && !spec.HostStage.Valid() {
		return Key{}, fmt.Errorf("%w: %q", ErrInvalidHostStage, spec.HostStage)
	}
	if err := validateKeyComponent("toolchain", toolchain, true); err != nil {
		return Key{}, err
	}
	inputDigest, err := DigestOf(spec.Inputs)
	if err != nil {
		return Key{}, fmt.Errorf("hash inputs: %w", err)
	}
	freshness, err := NewFreshness(spec.Freshness)
	if err != nil {
		return Key{}, err
	}
	key, err := NewProjectionKey(ProjectionKeySpec{
		Domain: domain, Namespace: spec.Namespace, Version: version,
		ArtifactKind: defaultString(spec.ArtifactKind, "projection"),
		Projection:   defaultString(spec.Projection, "default"), HostStage: spec.HostStage,
		SemanticClosureDigest: inputDigest, DependencyRoot: freshness.DependencyDigest,
		PolicySchemaDigest: freshness.ProvenanceDigest, Toolchain: toolchain,
		ToolVersion: spec.ToolVersion, Target: defaultString(spec.Target, "default"),
		BuildTags: spec.BuildTags, Options: spec.Options,
	})
	return key, err
}

func normalizeProjectionSpec(spec ProjectionKeySpec) (ProjectionKeySpec, error) {
	if spec.Version == "" {
		spec.Version = ProjectionKeyVersion
	}
	if spec.Domain == "" {
		spec.Domain = spec.Namespace
	}
	if spec.Namespace != "" && spec.Domain != spec.Namespace {
		return ProjectionKeySpec{}, fmt.Errorf("domain and namespace differ")
	}
	if spec.Toolchain == "" {
		spec.Toolchain = spec.ToolVersion
	}
	if spec.HostStage == "" {
		spec.HostStage = DefaultHostStage
	}
	if spec.ToolVersion != "" && spec.Toolchain != spec.ToolVersion {
		return ProjectionKeySpec{}, fmt.Errorf("toolchain and tool version differ")
	}
	for _, field := range []struct {
		label    string
		value    string
		required bool
	}{
		{"domain", spec.Domain, true}, {"key version", spec.Version, true},
		{"artifact kind", spec.ArtifactKind, true}, {"projection", spec.Projection, true},
		{"toolchain", spec.Toolchain, true}, {"target", spec.Target, true},
	} {
		if err := validateKeyComponent(field.label, field.value, field.required); err != nil {
			return ProjectionKeySpec{}, err
		}
	}
	if !spec.HostStage.Valid() {
		return ProjectionKeySpec{}, fmt.Errorf("%w: %q", ErrInvalidHostStage, spec.HostStage)
	}
	for _, item := range []struct {
		label string
		value Digest
	}{
		{"semantic closure digest", spec.SemanticClosureDigest},
		{"dependency root", spec.DependencyRoot}, {"policy schema digest", spec.PolicySchemaDigest},
	} {
		if !item.value.Known() {
			return ProjectionKeySpec{}, fmt.Errorf("%w: %s", ErrUnknownFreshness, item.label)
		}
	}
	tags, err := canonicalBuildTags(spec.BuildTags)
	if err != nil {
		return ProjectionKeySpec{}, err
	}
	spec.BuildTags = tags
	return spec, nil
}

func canonicalBuildTags(tags []string) ([]string, error) {
	canonical := append([]string(nil), tags...)
	for _, tag := range canonical {
		if err := validateKeyComponent("build tag", tag, true); err != nil {
			return nil, err
		}
	}
	sort.Strings(canonical)
	unique := canonical[:0]
	for _, tag := range canonical {
		if len(unique) == 0 || unique[len(unique)-1] != tag {
			unique = append(unique, tag)
		}
	}
	return unique, nil
}

func digestBuildTags(tags []string) (Digest, error) {
	if tags == nil {
		tags = []string{}
	}
	digest, err := DigestOf(tags)
	if err != nil {
		return "", fmt.Errorf("hash build tags: %w", err)
	}
	return digest, nil
}

func digestForProjectionKey(key ProjectionKey) Digest {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(keyDomain))
	for _, value := range []string{
		key.Domain, key.Version, key.ArtifactKind, key.Projection, key.HostStage.String(),
		key.SemanticClosureDigest.String(), key.DependencyRoot.String(), key.PolicySchemaDigest.String(),
		key.Toolchain, key.Target, key.BuildTagsDigest.String(), key.OptionsDigest.String(),
	} {
		writeKeyPart(hasher, value)
	}
	return Digest(fmt.Sprintf("%x", hasher.Sum(nil)))
}

// ParseKey validates a serialized digest. Descriptive fields are unavailable
// after parsing, but metadata validates the stored object before returning it.
func ParseKey(serialized string) (Key, error) {
	serialized = strings.ToLower(serialized)
	digest := Digest(serialized)
	if !digest.Valid() {
		return Key{}, fmt.Errorf("invalid cache key %q", serialized)
	}
	return Key{Digest: digest}, nil
}

// Valid reports whether the key has a safe serialized digest.
func (k Key) Valid() bool { return k.Digest.Valid() }

// String returns the stable serialized key digest.
func (k Key) String() string { return k.Digest.String() }

// Freshness returns compatibility dependency and provenance identities.
func (k Key) Freshness() Freshness {
	return Freshness{DependencyDigest: k.DependencyRoot, ProvenanceDigest: k.PolicySchemaDigest}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func writeKeyPart(hasher hash.Hash, value string) {
	var length [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(length[:], uint64(len(value)))
	_, _ = hasher.Write(length[:n])
	_, _ = hasher.Write([]byte(value))
}
