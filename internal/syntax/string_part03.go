package syntax

import (
	"strings"
	"unicode/utf8"
)

func (l *Lexer) consumeInvalidUTF8(value *strings.Builder) {
	invalidStart := l.position()
	l.advanceRune()
	value.WriteRune(utf8.RuneError)
	l.addDiagnostic(DiagInvalidUTF8, startSpan(l.filename, invalidStart, l.position()), "invalid UTF-8 byte in string literal")
}

func (l *Lexer) recoverEscape(value *strings.Builder, recovered string, escapeStart Position, detail string) bool {
	l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, escapeStart, l.position()), detail)
	value.WriteString(recovered)
	return false
}
