package syntax

import (
	"unicode/utf8"
)

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
