package syntax

import (
	"fmt"
)

// Position identifies a byte offset in a source file. Lines and columns are
// one-based; the end position of a span is exclusive. Column counts Unicode
// code points while Offset counts UTF-8 bytes.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span identifies a half-open source range: [Start, End).
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// IsEmpty reports whether the span has no source text.
func (s Span) IsEmpty() bool {
	return s.Start.Offset == s.End.Offset
}

// Len returns the span length in UTF-8 bytes.
func (s Span) Len() int {
	return s.End.Offset - s.Start.Offset
}

// Contains reports whether position is inside the span. The end position is
// exclusive, as it is for Go slices.
func (s Span) Contains(position Position) bool {
	return s.Start.Offset <= position.Offset && position.Offset < s.End.Offset
}

// String formats a span for diagnostics without requiring access to source
// text. A filename is included when one was supplied by the caller.
func (s Span) String() string {
	location := fmt.Sprintf("%d:%d", s.Start.Line, s.Start.Column)
	if s.Filename != "" {
		location = s.Filename + ":" + location
	}
	if s.Start.Offset == s.End.Offset {
		return location
	}
	return fmt.Sprintf("%s-%d:%d", location, s.End.Line, s.End.Column)
}

// TokenKind is the lexical category of a token.
type TokenKind uint8

const (
	TokenIllegal TokenKind = iota
	TokenEOF
	TokenIdentifier
	TokenString

	TokenPackage
	TokenNamespace
	TokenEntity
	TokenID
	TokenActivity

	TokenLParen
	TokenRParen
	TokenComma
	TokenArrow
	TokenLBrace
	TokenRBrace
	TokenDot
)
