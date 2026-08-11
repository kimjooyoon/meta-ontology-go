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

// EntityDecl declares a named entity and its stable semantic identifier.
type EntityDecl struct {
	Span     Span
	Name     string
	ID       string
	NameSpan Span
	IDSpan   Span
}

func (*EntityDecl) declarationNode()   {}
func (d *EntityDecl) SourceSpan() Span { return d.Span }

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
