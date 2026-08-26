package impactgraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// CanonicalJSON returns the normalized graph in stable JSON order.
func (graph Graph) CanonicalJSON() ([]byte, error) {
	normalized, err := graph.Normalized()
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("encode impact graph: %w", err)
	}
	return encoded, nil
}

// Canonical returns the stable JSON representation, or an empty string for an
// invalid graph. Call Validate or CanonicalJSON when the error is required.
func (graph Graph) Canonical() string {
	encoded, err := graph.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(encoded)
}

// Digest returns the SHA-256 hex digest of the canonical graph JSON. Invalid
// graphs have no digest and return the empty string.
func (graph Graph) Digest() string {
	encoded, err := graph.CanonicalJSON()
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// StableDigest is an explicit spelling of Digest.
func (graph Graph) StableDigest() string { return graph.Digest() }

// SHA256 is an explicit spelling of Digest.
func (graph Graph) SHA256() string { return graph.Digest() }
