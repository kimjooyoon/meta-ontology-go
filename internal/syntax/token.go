// Package syntax contains the surface syntax for the .gooo language.
//
// The syntax package deliberately has no knowledge of the semantic IR. It
// preserves source locations and reports recoverable diagnostics so callers
// can inspect a partial syntax tree after an invalid edit.
package syntax

import "fmt"

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
)

// Stable aliases make the token vocabulary convenient for callers that use
// the conventional lexer terminology.
const (
	InvalidToken          = TokenIllegal
	TokenInvalid          = TokenIllegal
	EndOfFile             = TokenEOF
	TokenEndOfFile        = TokenEOF
	Ident                 = TokenIdentifier
	TokenIdent            = TokenIdentifier
	StringLiteral         = TokenString
	TokenStringLiteral    = TokenString
	TokenKeywordPackage   = TokenPackage
	TokenKeywordNamespace = TokenNamespace
	TokenKeywordEntity    = TokenEntity
	TokenKeywordID        = TokenID
	TokenKeywordActivity  = TokenActivity
	TokenLeftParen        = TokenLParen
	TokenRightParen       = TokenRParen
)

var tokenKindNames = [...]string{
	TokenIllegal:    "illegal token",
	TokenEOF:        "end of file",
	TokenIdentifier: "identifier",
	TokenString:     "string",
	TokenPackage:    "package",
	TokenNamespace:  "namespace",
	TokenEntity:     "entity",
	TokenID:         "id",
	TokenActivity:   "activity",
	TokenLParen:     "(",
	TokenRParen:     ")",
	TokenComma:      ",",
	TokenArrow:      "->",
}

// String returns a stable human-readable token name.
func (k TokenKind) String() string {
	if int(k) < len(tokenKindNames) {
		return tokenKindNames[k]
	}
	return fmt.Sprintf("token(%d)", k)
}

// Token is a lexical token. Text is the exact source spelling. Value is the
// decoded value for string literals and equals Text for all other tokens.
type Token struct {
	Kind TokenKind
	Span Span
	Text string
	// Lexeme is the compatibility spelling for Text. Both fields contain the
	// exact source spelling and are populated together.
	Lexeme string
	Value  string
}

// Is reports whether the token has kind.
func (t Token) Is(kind TokenKind) bool {
	return t.Kind == kind
}

// Tokens is a lexer result. The slice always ends with exactly one EOF token.
type Tokens []Token

var keywordKinds = map[string]TokenKind{
	"package":   TokenPackage,
	"namespace": TokenNamespace,
	"entity":    TokenEntity,
	"id":        TokenID,
	"activity":  TokenActivity,
}
