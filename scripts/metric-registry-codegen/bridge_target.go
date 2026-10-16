package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
)

type targetField struct{ name, kind string }

func targetFields(path, name string) (map[string]targetField, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
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
	result := map[string]targetField{}
	names := map[string]bool{}
	for _, field := range shape.Fields.List {
		id, target, err := targetFieldIdentity(field)
		if err != nil {
			return nil, err
		}
		if _, duplicate := result[id]; duplicate || names[target.name] {
			return nil, fmt.Errorf("target field identity is ambiguous: %q", id)
		}
		result[id], names[target.name] = target, true
	}
	return result, nil
}

func targetFieldIdentity(field *ast.Field) (string, targetField, error) {
	kind, scalar := field.Type.(*ast.Ident)
	if !scalar || len(field.Names) != 1 || field.Tag == nil {
		return "", targetField{}, fmt.Errorf("target requires one named scalar and a gooo ID tag")
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", targetField{}, err
	}
	id, present := reflect.StructTag(raw).Lookup("gooo")
	if !present || id == "" {
		return "", targetField{}, fmt.Errorf("target field %s requires a gooo ID tag", field.Names[0].Name)
	}
	return id, targetField{name: field.Names[0].Name, kind: kind.Name}, nil
}
