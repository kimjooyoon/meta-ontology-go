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

	// Freshness carries formal meta-language declarations. They are parsed and
	// compiled by the feature that owns them, while ordinary entity/activity
	// lowering remains unaware of this extension.
	Freshness []*FreshnessDecl
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
