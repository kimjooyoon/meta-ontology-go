package bodylibrary

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Change typed objects, including keyed literals and selectors. Equal spellings
// in another scope and source string values are never treated as references.
func nameLibraryProjection(source, packageName, sourceDigest string, plan bodyexecution.CompositionPlan, exports map[string]string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "library.go", source, parser.ParseComments)
	if err != nil {
		return "", err
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	pkg, err := new(types.Config).Check("gooo.generated.library", fset, []*ast.File{file}, info)
	if err != nil {
		return "", fmt.Errorf("library bridge typecheck: %w", err)
	}
	names, identities := libraryFunctionNames(file, info, exports), map[types.Object]string{}
	for _, record := range plan.Records {
		object := pkg.Scope().Lookup(record.GoName)
		if object == nil {
			return "", fmt.Errorf("library record %q is absent", record.ID)
		}
		shape, ok := object.Type().Underlying().(*types.Struct)
		if !ok || shape.NumFields() != len(record.Fields) {
			return "", fmt.Errorf("library record %q differs from its typed plan", record.ID)
		}
		names[object] = "R_" + record.Name
		for i, field := range record.Fields {
			names[shape.Field(i)], identities[shape.Field(i)] = "F_"+field.Name, field.ID
		}
	}
	avoidLibraryShadowing(file, info, names)
	file.Name.Name = packageName
	ast.Inspect(file, func(node ast.Node) bool {
		if field, ok := node.(*ast.Field); ok && len(field.Names) == 1 {
			if id := identities[info.Defs[field.Names[0]]]; id != "" {
				tag, _ := strconv.Unquote(field.Tag.Value) // The compiler emitted and typechecked this tag.
				tag += fmt.Sprintf(" gooo:%q", id)
				field.Tag.Value = strconv.Quote(tag)
				if !strings.ContainsAny(tag, "`\r\n") {
					field.Tag.Value = "`" + tag + "`"
				}
			}
		}
		if name, ok := node.(*ast.Ident); ok {
			object := info.Uses[name]
			if object == nil {
				object = info.Defs[name]
			}
			if replacement := names[object]; replacement != "" {
				name.Name = replacement
			}
		}
		return true
	})
	return renderLibraryProjection(fset, file, sourceDigest, plan.TypedPlanSHA256)
}
