package coupling

import (
	"unicode/utf8"
)

func liveUTF16Length(source string) int {
	units := 0
	for index := 0; index < len(source); {
		value, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		units += liveUTF16Width(value)
		index += size
	}
	return units
}
