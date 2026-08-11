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
	if !key.Valid() || key.Version == "" || key.Namespace == "" || !key.InputDigest.Valid() || !key.OptionsDigest.Valid() {
		return fmt.Errorf("%w: key must be created by NewKey", ErrInvalidKey)
	}
	if err := validateKeyComponent("key version", key.Version, true); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	if err := validateKeyComponent("namespace", key.Namespace, true); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	if err := validateKeyComponent("tool version", key.ToolVersion, false); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	expected := digestForKey(key.Version, key.Namespace, key.ToolVersion, key.InputDigest, key.OptionsDigest)
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
	if key.ToolVersion != "" && metadata.ToolVersion != key.ToolVersion {
		return fmt.Errorf("tool version mismatch")
	}
	if key.InputDigest.Valid() && metadata.InputDigest != key.InputDigest {
		return fmt.Errorf("input digest mismatch")
	}
	if key.OptionsDigest.Valid() && metadata.OptionsDigest != key.OptionsDigest {
		return fmt.Errorf("options digest mismatch")
	}
	return nil
}

func metadataSane(metadata Metadata) bool {
	if metadata.FormatVersion != metadataVersion || metadata.Key == "" || !isDigestName(metadata.Key) {
		return false
	}
	if !metadata.InputDigest.Valid() || !metadata.OptionsDigest.Valid() || !metadata.ContentDigest.Valid() {
		return false
	}
	if !metadata.Reconstructable || metadata.Size < 0 || metadata.CreatedAt.IsZero() {
		return false
	}
	if validateKeyComponent("key version", metadata.KeyVersion, true) != nil ||
		validateKeyComponent("namespace", metadata.Namespace, true) != nil ||
		validateKeyComponent("tool version", metadata.ToolVersion, false) != nil ||
		validateKeyComponent("artifact type", metadata.ArtifactType, false) != nil ||
		validateKeyComponent("projection", metadata.Projection, false) != nil {
		return false
	}
	if metadata.Key != digestForKey(metadata.KeyVersion, metadata.Namespace, metadata.ToolVersion,
		metadata.InputDigest, metadata.OptionsDigest).String() {
		return false
	}
	return metadata.MetadataDigest.Valid() && digestMetadata(metadata) == metadata.MetadataDigest
}

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
