package cache

import "fmt"

// FreshnessSpec names the caller-owned inputs whose change makes an artifact
// stale. Callers may pass semantic dependency records and provenance facts;
// this package hashes them without importing the semantic compiler package.
type FreshnessSpec struct {
	Dependencies any
	Provenance   any
}

// Freshness is the canonical identity of dependency and provenance inputs.
// A valid digest of nil represents an explicitly empty input, not missing
// freshness evidence.
type Freshness struct {
	DependencyDigest Digest
	ProvenanceDigest Digest
}

// NewFreshness canonicalizes dependency and provenance values independently.
func NewFreshness(spec FreshnessSpec) (Freshness, error) {
	dependencyDigest, err := DigestOf(spec.Dependencies)
	if err != nil {
		return Freshness{}, fmt.Errorf("hash dependencies: %w", err)
	}
	provenanceDigest, err := DigestOf(spec.Provenance)
	if err != nil {
		return Freshness{}, fmt.Errorf("hash provenance: %w", err)
	}
	return Freshness{DependencyDigest: dependencyDigest, ProvenanceDigest: provenanceDigest}, nil
}

// Valid reports whether both freshness dimensions have canonical digests.
func (f Freshness) Valid() bool {
	return f.DependencyDigest.Valid() && f.ProvenanceDigest.Valid()
}

// Validate rejects freshness values that were not produced from valid inputs.
func (f Freshness) Validate() error {
	if !f.DependencyDigest.Valid() {
		return fmt.Errorf("%w: dependency digest", ErrInvalidFreshness)
	}
	if !f.ProvenanceDigest.Valid() {
		return fmt.Errorf("%w: provenance digest", ErrInvalidFreshness)
	}
	return nil
}

// Equal reports whether both dependency and provenance identities match.
func (f Freshness) Equal(other Freshness) bool {
	return f.DependencyDigest == other.DependencyDigest && f.ProvenanceDigest == other.ProvenanceDigest
}

// GetFresh returns a verified artifact only when its dependency and
// provenance identities match freshness. Get remains available for explicit
// inspection of an older artifact.
func (c *Cache) GetFresh(key Key, freshness Freshness) ([]byte, Metadata, error) {
	if err := freshness.Validate(); err != nil {
		return nil, Metadata{}, err
	}
	data, metadata, err := c.Get(key)
	if err != nil {
		return nil, Metadata{}, err
	}
	if !metadataFreshness(metadata).Equal(freshness) {
		return nil, Metadata{}, fmt.Errorf("%w: %s", ErrStale, key)
	}
	return data, metadata, nil
}

func metadataFreshness(metadata Metadata) Freshness {
	return Freshness{DependencyDigest: metadata.DependencyDigest, ProvenanceDigest: metadata.ProvenanceDigest}
}

// StaleFilter selects a scoped set of artifacts whose freshness no longer
// matches Current. At least one scope field is required to avoid broad,
// accidental invalidation.
type StaleFilter struct {
	Namespace    string
	KeyVersion   string
	ToolVersion  string
	HostStage    HostStage
	ArtifactType string
	Projection   string
	Current      Freshness
}

func (filter StaleFilter) validate() error {
	if filter.Namespace == "" && filter.KeyVersion == "" && filter.ToolVersion == "" &&
		filter.HostStage == "" && filter.ArtifactType == "" && filter.Projection == "" {
		return ErrEmptyFilter
	}
	fields := []struct {
		label string
		value string
	}{
		{"namespace", filter.Namespace}, {"key version", filter.KeyVersion},
		{"tool version", filter.ToolVersion}, {"artifact type", filter.ArtifactType},
		{"projection", filter.Projection},
	}
	for _, field := range fields {
		if err := validateKeyComponent(field.label, field.value, false); err != nil {
			return err
		}
	}
	if filter.HostStage != "" && !filter.HostStage.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidHostStage, filter.HostStage)
	}
	return filter.Current.Validate()
}

func (filter StaleFilter) matches(metadata Metadata) bool {
	return (filter.Namespace == "" || filter.Namespace == metadata.Namespace) &&
		(filter.KeyVersion == "" || filter.KeyVersion == metadata.KeyVersion) &&
		(filter.ToolVersion == "" || filter.ToolVersion == metadata.ToolVersion) &&
		(filter.HostStage == "" || filter.HostStage == metadata.HostStage) &&
		(filter.ArtifactType == "" || filter.ArtifactType == metadata.ArtifactType) &&
		(filter.Projection == "" || filter.Projection == metadata.Projection)
}
