package syntax

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// lexEscape returns true when the escape ends the recoverable string token.
func (l *Lexer) lexEscape(value *strings.Builder, escapeStart Position) bool {
	if l.offset >= len(l.source) {
		l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, escapeStart, l.position()), "unterminated escape sequence")
		return true
	}
	r, _ := l.peekRune()
	if r == utf8.RuneError {
		_, size := l.peekRune()
		if size == 1 {
			l.consumeInvalidUTF8(value)
			return false
		}
	}
	if r == '\n' || r == '\r' {
		l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, escapeStart, l.position()), "unterminated escape sequence")
		return true
	}
	switch r {
	case '"', '\\':
		value.WriteRune(l.advanceRune())
	case 'n':
		l.advanceRune()
		value.WriteByte('\n')
	case 'r':
		l.advanceRune()
		value.WriteByte('\r')
	case 't':
		l.advanceRune()
		value.WriteByte('\t')
	case 'u':
		l.advanceRune()
		begin := l.offset
		var recovered strings.Builder
		for i := 0; i < 4 && l.offset < len(l.source); i++ {
			r, size := l.peekRune()
			if r == '\n' || r == '\r' || r == '"' {
				break
			}
			if r == utf8.RuneError && size == 1 {
				l.consumeInvalidUTF8(&recovered)
				continue
			}
			recovered.WriteRune(l.advanceRune())
		}
		raw := l.source[begin:l.offset]
		if len(raw) != 4 {
			l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, escapeStart, l.position()), "unicode escape must contain four hexadecimal digits")
			value.WriteString(recovered.String())
			return false
		}
		decoded, err := strconv.ParseUint(raw, 16, 16)
		if err != nil {
			l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, escapeStart, l.position()), "invalid unicode escape")
			value.WriteString(recovered.String())
			return false
		}
		if decoded >= 0xd800 && decoded <= 0xdfff {
			l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, escapeStart, l.position()), "unicode escape cannot encode a surrogate code point")
			value.WriteString(recovered.String())
			return false
		}
		value.WriteRune(rune(decoded))
	default:
		badStart := l.position()
		l.advanceRune()
		l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, badStart, l.position()), "invalid escape sequence")
		value.WriteRune(r)
	}
	return false
}
