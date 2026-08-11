package syntax

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

func (l *Lexer) lexString(start Position) {
	var value strings.Builder
	l.advanceRune() // opening quote
	terminated := false
	for l.offset < len(l.source) {
		r, size := l.peekRune()
		switch {
		case r == '"':
			l.advanceRune()
			terminated = true
		case r == '\n' || r == '\r':
			l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, start, l.position()), "unterminated string literal")
			terminated = true // emit the partial token and let whitespace handle EOL
		case r == '\\':
			l.advanceRune()
			if l.lexEscape(&value, start) {
				terminated = true
			}
		case r == utf8.RuneError && size == 1:
			invalidStart := l.position()
			l.advanceRune()
			value.WriteRune(utf8.RuneError)
			l.addDiagnostic(DiagInvalidUTF8, startSpan(l.filename, invalidStart, l.position()), "invalid UTF-8 byte in string literal")
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

// lexEscape returns true when the escape ends the recoverable string token.
func (l *Lexer) lexEscape(value *strings.Builder, stringStart Position) bool {
	if l.offset >= len(l.source) {
		l.addDiagnostic(DiagUnterminatedString, startSpan(l.filename, stringStart, l.position()), "unterminated escape sequence")
		return true
	}
	r, _ := l.peekRune()
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
		for i := 0; i < 4 && l.offset < len(l.source); i++ {
			l.advanceRune()
		}
		raw := l.source[begin:l.offset]
		if len(raw) != 4 {
			l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, stringStart, l.position()), "unicode escape must contain four hexadecimal digits")
			value.WriteString(raw)
			return false
		}
		decoded, err := strconv.ParseUint(raw, 16, 16)
		if err != nil {
			l.addDiagnostic(DiagInvalidEscape, startSpan(l.filename, stringStart, l.position()), "invalid unicode escape")
			value.WriteString(raw)
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
