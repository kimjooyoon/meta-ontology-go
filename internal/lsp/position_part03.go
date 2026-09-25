package lsp

import (
	"unicode/utf8"
)

func utf16Length(source string) int {
	units := 0
	for index := 0; index < len(source); {
		value, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		units += utf16Width(value)
		index += size
	}
	return units
}
