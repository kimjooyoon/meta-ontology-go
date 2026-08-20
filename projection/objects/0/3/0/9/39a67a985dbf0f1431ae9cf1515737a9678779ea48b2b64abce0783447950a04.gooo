package syntax

import (
	"strings"
	"unicode/utf8"
)

// Invalid UTF-8 in a quoted string is recovered byte by byte: each malformed
// byte emits one DiagInvalidUTF8 over its one-byte source span and contributes
// U+FFFD to the recovered string value. Escape syntax remains independent:
// malformed ASCII escapes still emit DiagInvalidEscape, and malformed \u
// escapes may emit both diagnostics when they contain malformed bytes.
func (l *Lexer) lexString(start Position) {
	var value strings.Builder
	l.advanceRune()
	terminated := false
	for l.offset < len(l.source) {
		r, size := l.peekRune()
		switch {
		case r == '"':
			l.advanceRune()
			terminated = true
		case r == '\n' || r == '\r':
			l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, start, l.position()), "unterminated string literal")
			terminated = true
		case r == '\\':
			escapeStart := l.position()
			l.advanceRune()
			if l.lexEscape(&value, escapeStart) {
				terminated = true
			}
		case r == utf8.RuneError && size == 1:
			l.consumeInvalidUTF8(&value)
		default:
			value.WriteRune(l.advanceRune())
		}
		if terminated {
			break
		}
	}
	if !terminated {
		l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, start, l.position()), "unterminated string literal")
	}
	end := l.position()
	raw := l.source[start.Offset:end.Offset]
	l.emitText(TokenString, start, end, raw, value.String())
}
