package coupling

import (
	"unicode/utf16"
	"unicode/utf8"
)

func liveLineStarts(source string) []int {
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
func liveLineEnd(source string, start int) int {
	for index := start; index < len(source); index++ {
		if source[index] == '\r' || source[index] == '\n' {
			return index
		}
	}
	return len(source)
}
func validateLiveOffset(source string, offset int) error {
	if offset < 0 || offset > len(source) {
		return ErrInvalidLivePosition
	}
	for index := 0; index < offset; {
		_, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		if index+size > offset {
			return ErrInvalidLivePosition
		}
		index += size
	}
	if offset > 0 && offset < len(source) && source[offset-1] == '\r' && source[offset] == '\n' {
		return ErrInvalidLivePosition
	}
	return nil
}
func liveUTF16Width(value rune) int {
	width := utf16.RuneLen(value)
	if width < 0 {
		return 1
	}
	return width
}
