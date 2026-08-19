package cache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

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
