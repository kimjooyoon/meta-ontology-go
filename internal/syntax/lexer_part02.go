package syntax

import (
	"unicode/utf8"
)

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
		case r == '.':
			l.advanceRune()
			l.emit(TokenDot, start)
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
