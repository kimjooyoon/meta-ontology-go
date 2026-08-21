package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// StableHash returns the lowercase SHA-256 digest of bytes. It is the common
// primitive used by graph, IR, node, and fact stable hashes.
func StableHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
func StableHashString(value string) string {
	return StableHash([]byte(value))
}
func (n Node) Canonical() string {
	if normalized, err := n.Normalized(); err == nil {
		n = normalized
	}
	var b strings.Builder
	b.WriteString("node\t")
	writeCanonicalField(&b, n.ID.String())
	writeCanonicalField(&b, n.Kind.String())
	writeCanonicalField(&b, n.Namespace.String())
	writeCanonicalField(&b, n.Name)
	for _, alias := range n.Aliases {
		writeCanonicalField(&b, alias)
	}
	writeCanonicalSpan(&b, n.Span)
	for _, field := range n.Fields {
		b.WriteString(field.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}
func (n Node) SemanticCanonical() string {
	if normalized, err := n.Normalized(); err == nil {
		n = normalized
	}
	var b strings.Builder
	b.WriteString("node\t")
	writeCanonicalField(&b, n.ID.String())
	writeCanonicalField(&b, n.Kind.String())
	writeCanonicalField(&b, n.Namespace.String())
	for _, field := range n.Fields {
		b.WriteString(field.SemanticCanonical())
		b.WriteByte('\n')
	}
	return b.String()
}
func (r TypeRef) Canonical() string {
	if normalized, err := r.Normalized(); err == nil {
		r = normalized
	}
	var b strings.Builder
	writeCanonicalField(&b, r.ID.String())
	writeCanonicalField(&b, r.Namespace.String())
	writeCanonicalField(&b, r.Name)
	return b.String()
}
func (r TypeRef) SemanticCanonical() string {
	if normalized, err := r.Normalized(); err == nil {
		r = normalized
	}
	return r.ID.String()
}
