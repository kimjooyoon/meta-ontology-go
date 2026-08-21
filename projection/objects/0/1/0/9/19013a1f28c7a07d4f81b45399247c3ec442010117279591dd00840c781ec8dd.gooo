package cache

import (
	"fmt"
	"path/filepath"
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
