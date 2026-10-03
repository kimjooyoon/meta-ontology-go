package syntax

import (
	"strings"
	"unicode/utf8"
)

// Raw strings preserve line endings and backslashes. Invalid UTF-8 retains the
// original byte span and diagnostic, using the existing quoted-string recovery.
func (l *Lexer) lexRawString(start Position) {
	var value strings.Builder
	l.advanceRune()
	terminated := false
	for l.offset < len(l.source) {
		r, size := l.peekRune()
		if r == '`' {
			l.advanceRune()
			terminated = true
			break
		}
		if r == utf8.RuneError && size == 1 {
			l.consumeInvalidUTF8(&value)
			continue
		}
		if r == '\n' || r == '\r' {
			begin := l.offset
			l.advanceNewline()
			value.WriteString(l.source[begin:l.offset])
			continue
		}
		value.WriteRune(l.advanceRune())
	}
	if !terminated {
		l.addDiagnostic(DiagUnterminatedString,
			startSpan(l.filename, start, l.position()), "unterminated raw string literal")
	}
	end := l.position()
	l.emitText(TokenString, start, end, l.source[start.Offset:end.Offset], value.String())
}
