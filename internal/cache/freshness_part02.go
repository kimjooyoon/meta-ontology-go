package cache

import (
	"fmt"
)

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
