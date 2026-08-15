package syntax

import "strings"

func quoteString(value string) string {
	var output strings.Builder
	output.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			output.WriteString(`\"`)
		case '\\':
			output.WriteString(`\\`)
		case '\n':
			output.WriteString(`\n`)
		case '\r':
			output.WriteString(`\r`)
		case '\t':
			output.WriteString(`\t`)
		default:
			if character < 0x20 || character == 0x7f {
				writeUnicodeEscape(&output, character)
			} else {
				output.WriteRune(character)
			}
		}
	}
	output.WriteByte('"')
	return output.String()
}

func writeUnicodeEscape(output *strings.Builder, character rune) {
	const hexadecimal = "0123456789abcdef"
	output.WriteString(`\u`)
	output.WriteByte(hexadecimal[(character>>12)&0xf])
	output.WriteByte(hexadecimal[(character>>8)&0xf])
	output.WriteByte(hexadecimal[(character>>4)&0xf])
	output.WriteByte(hexadecimal[character&0xf])
}
