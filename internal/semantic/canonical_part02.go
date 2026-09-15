package semantic

import (
	"strings"
)

func (f Field) Canonical() string {
	if normalized, err := f.Normalized(); err == nil {
		f = normalized
	}
	var b strings.Builder
	b.WriteString("field\t")
	writeCanonicalField(&b, f.ID.String())
	writeCanonicalField(&b, f.Parent.String())
	writeCanonicalField(&b, f.Name)
	for _, alias := range f.Aliases {
		writeCanonicalField(&b, alias)
	}
	b.WriteString(f.TypeRef.Canonical())
	writeCanonicalField(&b, string(f.Presence))
	writeCanonicalField(&b, string(f.Cardinality))
	writeCanonicalSpan(&b, f.Span)
	return b.String()
}
func (f Field) SemanticCanonical() string {
	if normalized, err := f.Normalized(); err == nil {
		f = normalized
	}
	var b strings.Builder
	b.WriteString("field\t")
	writeCanonicalField(&b, f.ID.String())
	writeCanonicalField(&b, f.Parent.String())
	writeCanonicalField(&b, f.TypeRef.SemanticCanonical())
	writeCanonicalField(&b, string(f.Presence))
	writeCanonicalField(&b, string(f.Cardinality))
	return b.String()
}
func (f Field) StableHash() string {
	return StableHashString(f.SemanticCanonical())
}
func (f Field) Hash() string {
	return f.StableHash()
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
