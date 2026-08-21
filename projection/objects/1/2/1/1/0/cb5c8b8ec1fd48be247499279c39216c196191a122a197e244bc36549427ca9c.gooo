package syntax

import (
	"fmt"
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
	TokenLeftBrace        = TokenLBrace
	TokenRightBrace       = TokenRBrace
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
	TokenLBrace:     "{",
	TokenRBrace:     "}",
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
