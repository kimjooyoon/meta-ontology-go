package formatter

import (
	"strings"
)

func render(document Document, options Options) string {
	lines := make([]string, 0, len(document.Declarations)+3)
	lines = append(lines, "package "+document.Package, "namespace "+document.Namespace)
	if len(document.Declarations) > 0 {
		lines = append(lines, "")
	}
	previousKind := DeclarationKind("")
	for _, declaration := range document.Declarations {
		if previousKind != "" && previousKind != declaration.Kind {
			lines = append(lines, "")
		}
		lines = append(lines, renderDeclaration(declaration))
		previousKind = declaration.Kind
	}
	result := strings.Join(lines, "\n")
	if options.FinalNewline {
		result += "\n"
	}
	return result
}

func renderDeclaration(declaration Declaration) string {
	if declaration.Kind == EntityDeclaration {
		return "entity " + declaration.Name + " id " + quoteString(declaration.ID)
	}
	return "activity " + declaration.Name + "(" + strings.Join(declaration.Inputs, ", ") + ") -> " + declaration.Output
}

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
