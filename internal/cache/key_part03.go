package cache

import (
	"fmt"
)

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
	optionsDigest, err := requireOptionsDigest(spec.OptionsDigest, spec.Options)
	if err != nil {
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
		BuildTags: spec.BuildTags, OptionsDigest: optionsDigest,
	})
	return key, err
}
