package cache

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func (c *Cache) objectPath(key Key) (string, error) {
	if err := c.validatePathKey(key); err != nil {
		return "", err
	}
	digest := key.String()
	return filepath.Join(c.objects, digest[:2], digest), nil
}

func (c *Cache) validatePathKey(key Key) error {
	if !key.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidKey, key.String())
	}
	return nil
}

func validateFullKey(key Key) error {
	if !key.Valid() || key.Domain == "" || key.Namespace != key.Domain || key.Version == "" ||
		!key.HostStage.Valid() || !key.SemanticClosureDigest.Known() || !key.DependencyRoot.Known() ||
		!key.PolicySchemaDigest.Known() || !key.BuildTagsDigest.Known() || !key.OptionsDigest.Known() {
		return fmt.Errorf("%w: key must be created by NewProjectionKey", ErrInvalidKey)
	}
	for _, field := range []struct {
		label string
		value string
	}{
		{"domain", key.Domain}, {"key version", key.Version}, {"artifact kind", key.ArtifactKind},
		{"projection", key.Projection}, {"toolchain", key.Toolchain}, {"target", key.Target},
	} {
		if err := validateKeyComponent(field.label, field.value, true); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidKey, err)
		}
	}
	if key.Namespace != key.Domain || key.ToolVersion != key.Toolchain ||
		key.InputDigest != key.SemanticClosureDigest || key.DependencyDigest != key.DependencyRoot ||
		key.ProvenanceDigest != key.PolicySchemaDigest {
		return fmt.Errorf("%w: compatibility aliases differ", ErrInvalidKey)
	}
	expected := digestForProjectionKey(key)
	if key.Digest != expected {
		return fmt.Errorf("%w: key digest does not match key fields", ErrInvalidKey)
	}
	return nil
}

func validateEntryInfo(info EntryInfo) error {
	if err := validateKeyComponent("artifact type", info.ArtifactType, false); err != nil {
		return err
	}
	return validateKeyComponent("projection", info.Projection, false)
}

func validateMetadataForKey(metadata Metadata, key Key) error {
	if !metadataSane(metadata) {
		return fmt.Errorf("metadata envelope is invalid")
	}
	if metadata.Key != key.String() {
		return fmt.Errorf("metadata key mismatch")
	}
	if key.Version != "" && metadata.KeyVersion != key.Version {
		return fmt.Errorf("key version mismatch")
	}
	if key.Namespace != "" && metadata.Namespace != key.Namespace {
		return fmt.Errorf("namespace mismatch")
	}
	if key.Domain != "" && metadata.Domain != key.Domain {
		return fmt.Errorf("domain mismatch")
	}
	if key.ArtifactKind != "" && metadata.ArtifactKind != key.ArtifactKind {
		return fmt.Errorf("artifact kind mismatch")
	}
	if key.Projection != "" && metadata.Projection != key.Projection {
		return fmt.Errorf("projection mismatch")
	}
	if key.ToolVersion != "" && metadata.ToolVersion != key.ToolVersion {
		return fmt.Errorf("tool version mismatch")
	}
	if key.Toolchain != "" && metadata.Toolchain != key.Toolchain {
		return fmt.Errorf("toolchain mismatch")
	}
	if key.Target != "" && metadata.Target != key.Target {
		return fmt.Errorf("target mismatch")
	}
	if key.HostStage != "" && metadata.HostStage != key.HostStage {
		return fmt.Errorf("host stage mismatch")
	}
	if key.InputDigest.Valid() && metadata.InputDigest != key.InputDigest {
		return fmt.Errorf("input digest mismatch")
	}
	if key.SemanticClosureDigest.Known() && metadata.SemanticClosureDigest != key.SemanticClosureDigest {
		return fmt.Errorf("semantic closure digest mismatch")
	}
	if key.DependencyRoot.Known() && metadata.DependencyRoot != key.DependencyRoot {
		return fmt.Errorf("dependency root mismatch")
	}
	if key.PolicySchemaDigest.Known() && metadata.PolicySchemaDigest != key.PolicySchemaDigest {
		return fmt.Errorf("policy schema digest mismatch")
	}
	if key.BuildTagsDigest.Known() && metadata.BuildTagsDigest != key.BuildTagsDigest {
		return fmt.Errorf("build tags digest mismatch")
	}
	if key.OptionsDigest.Valid() && metadata.OptionsDigest != key.OptionsDigest {
		return fmt.Errorf("options digest mismatch")
	}
	if key.DependencyDigest.Valid() && metadata.DependencyDigest != key.DependencyDigest {
		return fmt.Errorf("dependency digest mismatch")
	}
	if key.ProvenanceDigest.Valid() && metadata.ProvenanceDigest != key.ProvenanceDigest {
		return fmt.Errorf("provenance digest mismatch")
	}
	return nil
}

func metadataSane(metadata Metadata) bool {
	if metadata.FormatVersion != metadataVersion || metadata.Key == "" || !isDigestName(metadata.Key) {
		return false
	}
	if !metadata.InputDigest.Known() || !metadata.SemanticClosureDigest.Known() ||
		!metadata.DependencyRoot.Known() || !metadata.PolicySchemaDigest.Known() ||
		!metadata.BuildTagsDigest.Known() || !metadata.OptionsDigest.Known() ||
		!metadata.DependencyDigest.Known() || !metadata.ProvenanceDigest.Known() ||
		!metadata.ContentDigest.Valid() {
		return false
	}
	if !metadata.Reconstructable || metadata.Size < 0 || metadata.CreatedAt.IsZero() {
		return false
	}
	if !metadata.HostStage.Valid() || validateKeyComponent("key version", metadata.KeyVersion, true) != nil ||
		validateKeyComponent("domain", metadata.Domain, true) != nil ||
		validateKeyComponent("namespace", metadata.Namespace, true) != nil ||
		validateKeyComponent("artifact kind", metadata.ArtifactKind, true) != nil ||
		validateKeyComponent("tool version", metadata.ToolVersion, false) != nil ||
		validateKeyComponent("toolchain", metadata.Toolchain, true) != nil ||
		validateKeyComponent("target", metadata.Target, true) != nil ||
		validateKeyComponent("artifact type", metadata.ArtifactType, false) != nil ||
		validateKeyComponent("projection", metadata.Projection, false) != nil {
		return false
	}
	if metadata.Domain != metadata.Namespace || metadata.Toolchain != metadata.ToolVersion ||
		metadata.InputDigest != metadata.SemanticClosureDigest ||
		metadata.DependencyDigest != metadata.DependencyRoot ||
		metadata.ProvenanceDigest != metadata.PolicySchemaDigest {
		return false
	}
	key := ProjectionKey{
		Digest: metadataDigestPlaceholder(metadata.Key), Domain: metadata.Domain, Namespace: metadata.Namespace,
		Version: metadata.KeyVersion, ArtifactKind: metadata.ArtifactKind, Projection: metadata.Projection,
		HostStage: metadata.HostStage, SemanticClosureDigest: metadata.SemanticClosureDigest,
		DependencyRoot: metadata.DependencyRoot, PolicySchemaDigest: metadata.PolicySchemaDigest,
		Toolchain: metadata.Toolchain, ToolVersion: metadata.ToolVersion, Target: metadata.Target,
		BuildTagsDigest: metadata.BuildTagsDigest, OptionsDigest: metadata.OptionsDigest,
		InputDigest: metadata.InputDigest, DependencyDigest: metadata.DependencyDigest,
		ProvenanceDigest: metadata.ProvenanceDigest,
	}
	if metadata.Key != digestForProjectionKey(key).String() {
		return false
	}
	return metadata.MetadataDigest.Valid() && digestMetadata(metadata) == metadata.MetadataDigest
}

func metadataDigestPlaceholder(serialized string) Digest { return Digest(serialized) }

func corruptError(key Key, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrCorrupt, key, reason)
}

func validateKeyComponent(label, value string, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", label)
	}
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("%s must not contain NUL", label)
	}
	return nil
}
