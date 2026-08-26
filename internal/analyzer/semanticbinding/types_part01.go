package semanticbinding

// Role is the implementation role of a Go declaration.
type Role string

const (
	RoleHandwrittenImpl Role = "HANDWRITTEN_IMPL"
	RoleGeneratedImpl   Role = "GENERATED_IMPL"
	RoleAdapter         Role = "ADAPTER"

	HandwrittenImpl = RoleHandwrittenImpl
	GeneratedImpl   = RoleGeneratedImpl
	Adapter         = RoleAdapter
)

// Status describes whether extraction produced a complete result.
type Status string

const (
	StatusBound   Status = "BOUND"
	StatusUnknown Status = "UNKNOWN"
)

// SourceFile is one explicit Go source input. PackagePath is required unless
// Input.PackagePath supplies the same path for every source.
type SourceFile struct {
	Filename    string
	PackagePath string
	Source      []byte
}

// Source is a concise alias for SourceFile.
type Source = SourceFile

// Input supplies one Go package and, optionally, an explicit ID registry.
// A non-nil RegisteredIDs slice makes registry membership part of validation;
// an ID absent from it produces UNKNOWN rather than an inferred binding.
type Input struct {
	Sources       []SourceFile
	Files         []SourceFile
	PackagePath   string
	RegisteredIDs []string
}

// Position is a token.FileSet position with a zero-based byte offset and
// one-based line and column numbers.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is a half-open source range.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// SourceSpan is a vocabulary alias for Span.
type SourceSpan = Span

// Binding is one explicit gooo:bind declaration binding.
type Binding struct {
	ID              string
	Role            Role
	PackagePath     string
	DeclarationKey  string
	Span            Span
	DirectiveSpan   Span
	Digest          string
	CanonicalDigest string
}
