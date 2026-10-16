// Package recordbridge reads native scalar field identities from Go source.
package recordbridge

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

type Field struct{ Name, Kind string }

func Fields(filename string, raw []byte, name string) (map[string]Field, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, raw, 0)
	if err != nil {
		return nil, err
	}
	var shape *ast.StructType
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, entry := range group.Specs {
			spec, ok := entry.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			record, ok := spec.Type.(*ast.StructType)
			if !ok || shape != nil || spec.Assign.IsValid() || spec.TypeParams != nil {
				return nil, fmt.Errorf("target type is ambiguous or not a declared record")
			}
			shape = record
		}
	}
	if shape == nil {
		return nil, fmt.Errorf("target type unavailable")
	}
	result := map[string]Field{}
	names := map[string]bool{}
	for _, field := range shape.Fields.List {
		id, target, err := identity(field)
		if err != nil {
			return nil, err
		}
		if _, duplicate := result[id]; duplicate || names[target.Name] {
			return nil, fmt.Errorf("target field identity is ambiguous: %q", id)
		}
		result[id], names[target.Name] = target, true
	}
	return result, nil
}

func identity(field *ast.Field) (string, Field, error) {
	kind, scalar := field.Type.(*ast.Ident)
	if !scalar || len(field.Names) != 1 || field.Tag == nil {
		return "", Field{}, fmt.Errorf("target requires one named scalar and a gooo ID tag")
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", Field{}, err
	}
	id, present, err := TagID(raw)
	if err != nil {
		return "", Field{}, fmt.Errorf("target field %s tag: %w", field.Names[0].Name, err)
	}
	if !present || id == "" {
		return "", Field{}, fmt.Errorf("target field %s requires a gooo ID tag", field.Names[0].Name)
	}
	return id, Field{Name: field.Names[0].Name, Kind: kind.Name}, nil
}
