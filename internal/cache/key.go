package cache

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"strings"
)

const (
	// DefaultKeyVersion is the current cache-key schema. Bumping it invalidates
	// old keys without requiring a destructive cache migration.
	DefaultKeyVersion = "v1"
	keyDomain         = "gooo-content-cache-key\x00"
)

// KeySpec describes every input that can affect a reconstructable projection.
// Inputs and Options are canonicalized before hashing, so map iteration order
// cannot change the resulting key.
type KeySpec struct {
	Version     string
	Namespace   string
	ToolVersion string
	Inputs      any
	Options     any
}

// Key is the content address of one projection computation. The digest is the
// only part used in a filesystem path; the remaining fields make metadata and
// invalidation inspectable.
type Key struct {
	Digest        Digest
	Version       string
	Namespace     string
	ToolVersion   string
	InputDigest   Digest
	OptionsDigest Digest
}

// NewKey creates a versioned content-addressed key.
func NewKey(spec KeySpec) (Key, error) {
	version := spec.Version
	if version == "" {
		version = DefaultKeyVersion
	}
	if err := validateKeyComponent("key version", version, false); err != nil {
		return Key{}, err
	}
	if err := validateKeyComponent("namespace", spec.Namespace, true); err != nil {
		return Key{}, err
	}
	if err := validateKeyComponent("tool version", spec.ToolVersion, false); err != nil {
		return Key{}, err
	}

	inputDigest, err := DigestOf(spec.Inputs)
	if err != nil {
		return Key{}, fmt.Errorf("hash inputs: %w", err)
	}
	optionsDigest, err := DigestOf(spec.Options)
	if err != nil {
		return Key{}, fmt.Errorf("hash options: %w", err)
	}

	digest := digestForKey(version, spec.Namespace, spec.ToolVersion, inputDigest, optionsDigest)

	return Key{
		Digest:        digest,
		Version:       version,
		Namespace:     spec.Namespace,
		ToolVersion:   spec.ToolVersion,
		InputDigest:   inputDigest,
		OptionsDigest: optionsDigest,
	}, nil
}

func digestForKey(version, namespace, toolVersion string, inputDigest, optionsDigest Digest) Digest {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(keyDomain))
	writeKeyPart(hasher, version)
	writeKeyPart(hasher, namespace)
	writeKeyPart(hasher, toolVersion)
	writeKeyPart(hasher, inputDigest.String())
	writeKeyPart(hasher, optionsDigest.String())
	return Digest(fmt.Sprintf("%x", hasher.Sum(nil)))
}

// ParseKey validates a serialized digest. Parsed keys intentionally omit the
// descriptive fields; cache reads still work, while namespace-based
// invalidation should use keys returned by NewKey or inspect Metadata first.
func ParseKey(serialized string) (Key, error) {
	serialized = strings.ToLower(serialized)
	digest := Digest(serialized)
	if !digest.Valid() {
		return Key{}, fmt.Errorf("invalid cache key %q", serialized)
	}
	return Key{Digest: digest}, nil
}

// Valid reports whether the key has a safe, complete digest.
func (k Key) Valid() bool { return k.Digest.Valid() }

// String returns the stable serialized key digest.
func (k Key) String() string { return k.Digest.String() }

func writeKeyPart(hasher hash.Hash, value string) {
	var length [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(length[:], uint64(len(value)))
	_, _ = hasher.Write(length[:n])
	_, _ = hasher.Write([]byte(value))
}
