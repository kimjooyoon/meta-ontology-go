package lsp

import (
	"unicode"
	"unicode/utf8"
)

func identifierAt(source string, offset int) bool {
	if offset < 0 || offset >= len(source) {
		return false
	}
	value, _ := utf8.DecodeRuneInString(source[offset:])
	return isIdentifier(value)
}
func isIdentifier(value rune) bool {
	return value == '_' || unicode.IsLetter(value) || unicode.IsDigit(value)
}
func byteRange(source string, start, end int) (Range, error) {
	startPosition, err := OffsetToPosition(source, start)
	if err != nil {
		return Range{}, err
	}
	endPosition, err := OffsetToPosition(source, end)
	if err != nil {
		return Range{}, err
	}
	return Range{Start: startPosition, End: endPosition}, nil
}
