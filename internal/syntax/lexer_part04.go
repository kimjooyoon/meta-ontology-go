package syntax

import (
	"unicode"
)

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
