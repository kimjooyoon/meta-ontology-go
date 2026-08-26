package syntax

// Valid reports whether the carrier contains a known presence value. It does
// not perform semantic validation or registry resolution.
func (p FieldPresence) Valid() bool {
	return p == FieldPresenceRequired || p == FieldPresenceOptional
}

// FieldCardinality records the technology-independent cardinality contract
// for a latent entity field.
type FieldCardinality string

const (
	FieldCardinalityOne  FieldCardinality = "one"
	FieldCardinalityMany FieldCardinality = "many"
)

// Valid reports whether the carrier contains a known cardinality value. It
// does not perform semantic validation.
func (c FieldCardinality) Valid() bool {
	return c == FieldCardinalityOne || c == FieldCardinalityMany
}

// TypeRefDecl is a technology-independent nominal type reference spelling.
// Spelling is intentionally not a Go type name; later lowering resolves it
// through the semantic type registry.
type TypeRefDecl struct {
	Span     Span
	Spelling string
}

func (r TypeRefDecl) SourceSpan() Span { return r.Span }

// FieldDecl is a latent ordered structural member of one EntityDecl. ID is an
// explicit textual stable identifier and is never derived from Name or a
// source path. The ordinary parser keeps fields deferred; the explicit
// supported parser populates this carrier without resolving types or deriving
// identities.
type FieldDecl struct {
	Span        Span
	ID          string
	Name        string
	TypeRef     TypeRefDecl
	Presence    FieldPresence
	Cardinality FieldCardinality

	IDSpan          Span
	NameSpan        Span
	PresenceSpan    Span
	CardinalitySpan Span
}

func (d *FieldDecl) SourceSpan() Span {
	if d == nil {
		return Span{}
	}
	return d.Span
}

// Clone returns a detached copy of the field declaration.
func (d FieldDecl) Clone() FieldDecl {
	d.TypeRef = TypeRefDecl{Span: d.TypeRef.Span, Spelling: d.TypeRef.Spelling}
	return d
}
