package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
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
	return b.String()
}

func (n Node) StableHash() string {
	return StableHashString(n.SemanticCanonical())
}

func (n Node) Hash() string {
	return n.StableHash()
}

func (f Fact) Canonical() string {
	if normalized, err := f.Normalized(); err == nil {
		f = normalized
	}
	var b strings.Builder
	b.WriteString("fact\t")
	writeCanonicalField(&b, f.Status.String())
	writeCanonicalField(&b, f.Subject.String())
	writeCanonicalField(&b, f.Predicate.String())
	writeCanonicalField(&b, f.Object.String())
	writeCanonicalField(&b, f.Reason)
	writeCanonicalSpan(&b, f.Span)
	return b.String()
}

func (f Fact) SemanticCanonical() string {
	if normalized, err := f.Normalized(); err == nil {
		f = normalized
	}
	var b strings.Builder
	b.WriteString("fact\t")
	writeCanonicalField(&b, f.Status.String())
	writeCanonicalField(&b, f.Subject.String())
	writeCanonicalField(&b, f.Predicate.String())
	writeCanonicalField(&b, f.Object.String())
	return b.String()
}

func (f Fact) StableHash() string {
	return StableHashString(f.SemanticCanonical())
}

func (f Fact) Hash() string {
	return f.StableHash()
}

func (g Graph) Canonical() string {
	var b strings.Builder
	for _, node := range g.Nodes() {
		b.WriteString(node.Canonical())
		b.WriteByte('\n')
	}
	for _, fact := range g.AllFacts() {
		b.WriteString(fact.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}

// SemanticCanonical is the authoritative meaning projection. It excludes
// source spans, names, aliases, candidate facts, and candidate explanations.
// Candidate observations remain available through Canonical and AllFacts but
// cannot change authority-gating hashes until explicitly promoted.
func (g Graph) SemanticCanonical() string {
	var b strings.Builder
	for _, node := range g.Nodes() {
		b.WriteString(node.SemanticCanonical())
		b.WriteByte('\n')
	}
	for _, fact := range g.DeterministicFacts() {
		b.WriteString(fact.SemanticCanonical())
		b.WriteByte('\n')
	}
	return b.String()
}

func (g Graph) StableHash() string {
	return StableHashString(g.SemanticCanonical())
}

func (g Graph) Hash() string {
	return g.StableHash()
}

func writeCanonicalField(b *strings.Builder, value string) {
	b.WriteString(strconv.Quote(value))
	b.WriteByte('\t')
}

func writeCanonicalSpan(b *strings.Builder, span Span) {
	writeCanonicalField(b, span.File)
	writeCanonicalField(b, strconv.Itoa(span.Start.Offset))
	writeCanonicalField(b, strconv.Itoa(span.Start.Line))
	writeCanonicalField(b, strconv.Itoa(span.Start.Column))
	writeCanonicalField(b, strconv.Itoa(span.End.Offset))
	writeCanonicalField(b, strconv.Itoa(span.End.Line))
	writeCanonicalField(b, strconv.Itoa(span.End.Column))
}
