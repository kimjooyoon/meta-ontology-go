package replay

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func astShape(file *syntax.File) (string, error) {
	if file == nil || file.Package == nil || file.Namespace == nil {
		return "", fmt.Errorf("syntax headers are absent")
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	var shape strings.Builder
	fmt.Fprintf(&shape, "package=%q\nnamespace=%q\n", file.Package.Name, file.Namespace.Name)
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			fmt.Fprintf(&shape, "entity=%q,%q,fields=%t\n", value.Name, value.ID, value.FieldsPresent)
			for _, field := range value.Fields {
				fmt.Fprintf(&shape, "field=%q,%q,%q,%q,%q\n", field.ID, field.Name,
					field.TypeRef.Spelling, field.Presence, field.Cardinality)
			}
		case *syntax.ActivityDecl:
			inputs := value.Inputs
			if inputs == nil {
				inputs = value.Parameters
			}
			fmt.Fprintf(&shape, "activity=%q", value.Name)
			for _, input := range inputs {
				fmt.Fprintf(&shape, ",input=%q", input.Name)
			}
			fmt.Fprintf(&shape, ",output=%q\n", value.Output)
		default:
			return "", fmt.Errorf("unsupported declaration %T", declaration)
		}
	}
	return shape.String(), nil
}
