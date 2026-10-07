package bodyexecution

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func compositionCallFunctions(generation bodycodegen.Result, node CompositionActivity, index int) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "calls.go", generation.Source, parser.ParseComments)
	if err != nil || len(file.Imports) != 0 {
		return "", fmt.Errorf("composition call closure requires pure declarations")
	}
	if err := checkCompositionRecordDeclarations(file, fset, generation.Report.RecordTypes); err != nil {
		return "", err
	}
	names := map[string]string{node.Name: fmt.Sprintf("GoooComposedActivity%d", index)}
	for i, call := range generation.Report.CallClosure.Activities {
		if _, exists := names[call.Name]; exists {
			return "", fmt.Errorf("duplicate call closure activity")
		}
		names[call.Name] = fmt.Sprintf("GoooComposedActivity%dCall%d", index, i)
	}
	if err := renameCompositionCalls(file, fset, names); err != nil {
		return "", err
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "\n//gooo:generated:start id=%q kind=\"activity\"\n", node.ID)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			function.Doc = nil
			if err := format.Node(&out, fset, function); err != nil {
				return "", err
			}
			out.WriteByte('\n')
		}
	}
	fmt.Fprintf(&out, "//gooo:generated:end id=%q kind=\"activity\"\n", node.ID)
	return out.String(), nil
}

func renameCompositionCalls(file *ast.File, fset *token.FileSet, names map[string]string) error {
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	if _, err := new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, info); err != nil {
		return err
	}
	objects := map[types.Object]string{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if function.Recv != nil || names[function.Name.Name] == "" {
			return fmt.Errorf("composition function is outside the declared call closure")
		}
		objects[info.Defs[function.Name]] = names[function.Name.Name]
	}
	if len(objects) != len(names) {
		return fmt.Errorf("composition call closure is incomplete")
	}
	for name, object := range info.Uses {
		if next := objects[object]; next != "" {
			name.Name = next
		}
	}
	for name, object := range info.Defs {
		if next := objects[object]; next != "" {
			name.Name = next
		}
	}
	return nil
}
