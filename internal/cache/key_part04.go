package cache

import (
	"fmt"
)

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
