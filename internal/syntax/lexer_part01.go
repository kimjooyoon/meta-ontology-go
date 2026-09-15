package syntax

// Lexer converts .gooo source text into tokens while retaining exact source
// spans. It is deterministic and never panics on malformed UTF-8 or input.
type Lexer struct {
	filename string
	source   string
	offset   int
	line     int
	column   int

	tokens      Tokens
	diagnostics Diagnostics
}

// NewLexer creates a lexer for an unnamed source.
func NewLexer(source string) *Lexer {
	return NewLexerFile("", source)
}

// NewLexerFile creates a lexer and associates filename with every token and
// diagnostic span.
func NewLexerFile(filename, source string) *Lexer {
	return &Lexer{
		filename: filename,
		source:   source,
		line:     1,
		column:   1,
	}
}
