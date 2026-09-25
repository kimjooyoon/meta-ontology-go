package syntax

import (
	"unicode"
)

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
