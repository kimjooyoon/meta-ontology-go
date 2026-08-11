package formatter

import (
	"strconv"
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
		return "entity " + declaration.Name + " id " + strconv.Quote(declaration.ID)
	}
	return "activity " + declaration.Name + "(" + strings.Join(declaration.Inputs, ", ") + ") -> " + declaration.Output
}
