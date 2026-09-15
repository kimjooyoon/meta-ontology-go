package lsp

import (
	"unicode/utf16"
	"unicode/utf8"
)

func sourceLineStarts(source string) []int {
	starts := []int{0}
	for index := 0; index < len(source); {
		switch source[index] {
		case '\r':
			index++
			if index < len(source) && source[index] == '\n' {
				index++
			}
			starts = append(starts, index)
		case '\n':
			index++
			starts = append(starts, index)
		default:
			_, size := utf8.DecodeRuneInString(source[index:])
			if size == 0 {
				size = 1
			}
			index += size
		}
	}
	return starts
}
func sourceLineEnd(source string, start int) int {
	for index := start; index < len(source); index++ {
		if source[index] == '\r' || source[index] == '\n' {
			return index
		}
	}
	return len(source)
}
func validateByteOffset(source string, offset int) error {
	if offset < 0 || offset > len(source) {
		return ErrInvalidPosition
	}
	for index := 0; index < offset; {
		_, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		if index+size > offset {
			return ErrInvalidPosition
		}
		index += size
	}
	if offset > 0 && offset < len(source) && source[offset-1] == '\r' && source[offset] == '\n' {
		return ErrInvalidPosition
	}
	return nil
}
func utf16Width(value rune) int {
	width := utf16.RuneLen(value)
	if width < 0 {
		return 1
	}
	return width
}
