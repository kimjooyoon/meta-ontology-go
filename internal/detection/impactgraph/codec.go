package impactgraph

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

// DecodeJSON decodes exactly one strict graph document.
func DecodeJSON(data []byte) (Graph, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var raw *Graph
	if err := decoder.Decode(&raw); err != nil {
		return Graph{}, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if raw == nil {
		return Graph{}, fmt.Errorf("%w: expected one graph object", ErrInvalidDocument)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Graph{}, fmt.Errorf("%w: multiple JSON values", ErrInvalidDocument)
		}
		return Graph{}, fmt.Errorf("%w: trailing data: %v", ErrInvalidDocument, err)
	}
	return raw.Normalized()
}

// Decode is the strict JSON graph decoder.
func Decode(data []byte) (Graph, error) { return DecodeJSON(data) }

// EncodeJSON returns the canonical compact JSON document followed by LF.
func EncodeJSON(graph Graph) ([]byte, error) {
	canonical, err := graph.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}

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
