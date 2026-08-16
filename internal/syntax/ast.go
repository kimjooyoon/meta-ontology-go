package syntax

// Node is implemented by every AST node.
type Node interface {
	SourceSpan() Span
}

// Declaration is implemented by entity and activity declarations.
type Declaration interface {
	Node
	declarationNode()
}

// File is the root of a parsed .gooo source file.
type File struct {
	Span Span

	// Package and Namespace retain declaration-level source spans and semantic
	// names for the two required headers.
	Package   *PackageDecl
	Namespace *NamespaceDecl

	// Decls is the canonical declaration list. Declarations is populated with
	// the same ordered values for callers that prefer the longer name.
	Decls        []Declaration
	Declarations []Declaration
}

// SourceFile and AST are descriptive aliases for callers that prefer those
// names when embedding the syntax tree in a compiler pipeline.
type SourceFile = File
type AST = File

// SourceSpan implements Node.
func (f *File) SourceSpan() Span { return f.Span }

// PackageDecl declares the package name.
type PackageDecl struct {
	Span     Span
	Name     string
	NameSpan Span
}

func (d *PackageDecl) SourceSpan() Span { return d.Span }

// NamespaceDecl declares the semantic namespace in which names are scoped.
type NamespaceDecl struct {
	Span     Span
	Name     string
	NameSpan Span
}

func (d *NamespaceDecl) SourceSpan() Span { return d.Span }

// NameRef is an identifier occurrence in an activity signature.
type NameRef struct {
	Span Span
	Name string
}

func (r NameRef) SourceSpan() Span { return r.Span }

// FieldPresence records the technology-independent presence contract for a
// latent entity field. The public parser does not currently construct these
// values; they are reserved for later lowering from an explicit AST carrier.
type FieldPresence string

const (
	FieldPresenceRequired FieldPresence = "required"
	FieldPresenceOptional FieldPresence = "optional"
)

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

// EntityDecl declares a named entity and its stable semantic identifier.
type EntityDecl struct {
	Span     Span
	Name     string
	ID       string
	NameSpan Span
	IDSpan   Span
	Fields   []FieldDecl
	// FieldsPresent distinguishes an explicit empty fields block from no block.
	FieldsPresent bool
}

func (*EntityDecl) declarationNode()   {}
func (d *EntityDecl) SourceSpan() Span { return d.Span }

// Clone returns a detached copy of the entity declaration, preserving field
// source order and preventing callers from sharing the Fields backing array.
func (d EntityDecl) Clone() EntityDecl {
	clone := d
	if d.Fields != nil {
		clone.Fields = make([]FieldDecl, len(d.Fields))
		for index, field := range d.Fields {
			clone.Fields[index] = field.Clone()
		}
	}
	return clone
}

// ActivityDecl declares an activity, its entity inputs, and its entity result.
type ActivityDecl struct {
	Span     Span
	Name     string
	NameSpan Span

	// Inputs and Output are the compact grammar-facing names. Parameters and
	// Result retain descriptive names for newer consumers.
	Inputs     []NameRef
	Output     string
	Parameters []NameRef
	Result     NameRef
}

func (*ActivityDecl) declarationNode()   {}
func (d *ActivityDecl) SourceSpan() Span { return d.Span }

// Clone returns a detached copy of the parsed syntax tree. The declaration
// aliases retain their existing same-order relationship while every mutable
// slice and declaration node is copied.
func (f *File) Clone() *File {
	if f == nil {
		return nil
	}
	clone := *f
	if f.Package != nil {
		packageDecl := *f.Package
		clone.Package = &packageDecl
	}
	if f.Namespace != nil {
		namespaceDecl := *f.Namespace
		clone.Namespace = &namespaceDecl
	}
	if f.Decls == nil && f.Declarations == nil {
		return &clone
	}
	declarations := f.Decls
	if declarations == nil {
		declarations = f.Declarations
	}
	clonedDeclarations := make([]Declaration, len(declarations))
	for index, declaration := range declarations {
		clonedDeclarations[index] = cloneDeclaration(declaration)
	}
	clone.Decls = clonedDeclarations
	clone.Declarations = clonedDeclarations
	return &clone
}

func cloneDeclaration(declaration Declaration) Declaration {
	switch value := declaration.(type) {
	case *EntityDecl:
		if value == nil {
			return (*EntityDecl)(nil)
		}
		clone := value.Clone()
		return &clone
	case *ActivityDecl:
		if value == nil {
			return (*ActivityDecl)(nil)
		}
		clone := *value
		clone.Inputs = append([]NameRef(nil), value.Inputs...)
		clone.Parameters = append([]NameRef(nil), value.Parameters...)
		return &clone
	default:
		return declaration
	}
}
