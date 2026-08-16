package syntax

import (
	"unicode"
	"unicode/utf8"
)

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

// Lex tokenizes the source. Repeated calls return the same deterministic
// result without appending duplicate EOF tokens or diagnostics.
func (l *Lexer) Lex() (Tokens, Diagnostics) {
	if l.tokens != nil {
		return append(Tokens(nil), l.tokens...), append(Diagnostics(nil), l.diagnostics...)
	}

	for l.offset < len(l.source) {
		if l.skipWhitespaceOrComment() {
			continue
		}

		start := l.position()
		r, size := l.peekRune()
		switch {
		case isIdentifierStart(r):
			l.lexIdentifier(start)
		case r == '"':
			l.lexString(start)
		case r == '(':
			l.advanceRune()
			l.emit(TokenLParen, start)
		case r == ')':
			l.advanceRune()
			l.emit(TokenRParen, start)
		case r == ',':
			l.advanceRune()
			l.emit(TokenComma, start)
		case r == '{':
			l.advanceRune()
			l.emit(TokenLBrace, start)
		case r == '}':
			l.advanceRune()
			l.emit(TokenRBrace, start)
		case r == '-' && l.peekByte(1) == '>':
			l.advanceRune()
			l.advanceRune()
			l.emitText(TokenArrow, start, l.position(), "->", "->")
		default:
			if r == utf8.RuneError && size == 1 {
				l.advanceRune()
			} else {
				l.advanceRune()
			}
			end := l.position()
			l.emitText(TokenIllegal, start, end, l.source[start.Offset:end.Offset], l.source[start.Offset:end.Offset])
			l.addDiagnostic(DiagUnexpectedCharacter, startSpan(l.filename, start, end), "unexpected character "+quoteSource(l.source[start.Offset:end.Offset]))
		}
	}

	position := l.position()
	l.tokens = append(l.tokens, Token{
		Kind:   TokenEOF,
		Span:   startSpan(l.filename, position, position),
		Text:   "",
		Lexeme: "",
		Value:  "",
	})
	return append(Tokens(nil), l.tokens...), append(Diagnostics(nil), l.diagnostics...)
}

// Lex is the convenience API for an unnamed source.
func Lex(source string) (Tokens, Diagnostics) {
	return NewLexer(source).Lex()
}

// LexFile is the convenience API for a named source.
func LexFile(filename, source string) (Tokens, Diagnostics) {
	return NewLexerFile(filename, source).Lex()
}

func (l *Lexer) position() Position {
	return Position{Offset: l.offset, Line: l.line, Column: l.column}
}

func (l *Lexer) peekByte(relative int) byte {
	index := l.offset + relative
	if index < 0 || index >= len(l.source) {
		return 0
	}
	return l.source[index]
}

func (l *Lexer) peekRune() (rune, int) {
	if l.offset >= len(l.source) {
		return 0, 0
	}
	r, size := utf8.DecodeRuneInString(l.source[l.offset:])
	return r, size
}

func (l *Lexer) advanceRune() rune {
	r, size := l.peekRune()
	if size == 0 {
		return 0
	}
	l.offset += size
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

func (l *Lexer) advanceNewline() {
	if l.peekByte(0) == '\r' {
		l.offset++
		if l.peekByte(0) == '\n' {
			l.offset++
		}
	} else {
		l.offset++
	}
	l.line++
	l.column = 1
}

func (l *Lexer) skipWhitespaceOrComment() bool {
	if l.offset >= len(l.source) {
		return false
	}
	r, _ := l.peekRune()
	if r == '\n' || r == '\r' {
		l.advanceNewline()
		return true
	}
	if unicode.IsSpace(r) {
		l.advanceRune()
		return true
	}

	if l.peekByte(0) != '/' {
		if l.peekByte(0) == '#' {
			l.advanceRune()
			for l.offset < len(l.source) {
				r, _ := l.peekRune()
				if r == '\n' || r == '\r' {
					break
				}
				l.advanceRune()
			}
			return true
		}
		return false
	}
	if l.peekByte(1) == '/' {
		l.advanceRune()
		l.advanceRune()
		for l.offset < len(l.source) {
			r, _ := l.peekRune()
			if r == '\n' || r == '\r' {
				break
			}
			l.advanceRune()
		}
		return true
	}
	if l.peekByte(1) != '*' {
		return false
	}

	start := l.position()
	l.advanceRune()
	l.advanceRune()
	for l.offset < len(l.source) {
		if l.peekByte(0) == '*' && l.peekByte(1) == '/' {
			l.advanceRune()
			l.advanceRune()
			return true
		}
		if l.peekByte(0) == '\n' || l.peekByte(0) == '\r' {
			l.advanceNewline()
		} else {
			l.advanceRune()
		}
	}
	l.addDiagnostic(DiagUnterminatedComment, startSpan(l.filename, start, l.position()), "unterminated block comment")
	return true
}

func (l *Lexer) lexIdentifier(start Position) {
	for l.offset < len(l.source) {
		r, _ := l.peekRune()
		if !isIdentifierContinue(r) {
			break
		}
		l.advanceRune()
	}
	end := l.position()
	text := l.source[start.Offset:end.Offset]
	kind, ok := keywordKinds[text]
	if !ok {
		kind = TokenIdentifier
	}
	l.emitText(kind, start, end, text, text)
}

func (l *Lexer) emit(kind TokenKind, start Position) {
	l.emitText(kind, start, l.position(), l.source[start.Offset:l.offset], l.source[start.Offset:l.offset])
}

func (l *Lexer) emitText(kind TokenKind, start, end Position, text, value string) {
	l.tokens = append(l.tokens, Token{
		Kind:   kind,
		Span:   startSpan(l.filename, start, end),
		Text:   text,
		Lexeme: text,
		Value:  value,
	})
}

func quoteSource(value string) string {
	return "'" + value + "'"
}

func (l *Lexer) addDiagnostic(code DiagnosticCode, span Span, message string) {
	l.diagnostics = append(l.diagnostics, Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		Span:     span,
	})
}

func startSpan(filename string, start, end Position) Span {
	return Span{Filename: filename, Start: start, End: end}
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierContinue(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}
