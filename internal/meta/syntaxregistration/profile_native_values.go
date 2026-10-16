package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"reflect"
)

func profileConstantValue(source *goSource, expression ast.Expr, visiting map[string]bool) (constant.Value, error) {
	switch value := expression.(type) {
	case *ast.BasicLit:
		if value.Kind == token.STRING {
			return constant.MakeFromLiteral(value.Value, token.STRING, 0), nil
		}
	case *ast.Ident:
		if visiting[value.Name] || len(visiting) >= 16 {
			return nil, fmt.Errorf("PROFILE_NATIVE_CONSTANT_CYCLE_OR_DEPTH")
		}
		bound, found, err := profileConstantSource(source, value.Name)
		if err != nil {
			return nil, err
		}
		if found {
			visiting[value.Name] = true
			defer delete(visiting, value.Name)
			return profileConstantValue(source, bound, visiting)
		}
		if value.Name == "true" || value.Name == "false" {
			return constant.MakeBool(value.Name == "true"), nil
		}
	}
	return nil, fmt.Errorf("PROFILE_NATIVE_VALUE_UNRESOLVED")
}

func profileConstantSource(source *goSource, name string) (ast.Expr, bool, error) {
	var found ast.Expr
	for _, declaration := range source.file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && function.Name.Name == name {
			return nil, false, fmt.Errorf("PROFILE_NATIVE_CONSTANT_SHADOWED")
		}
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			if kind, ok := spec.(*ast.TypeSpec); ok && kind.Name.Name == name {
				return nil, false, fmt.Errorf("PROFILE_NATIVE_CONSTANT_SHADOWED")
			}
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, identifier := range value.Names {
				if identifier.Name != name {
					continue
				}
				if group.Tok != token.CONST || len(value.Names) != len(value.Values) || found != nil {
					return nil, false, fmt.Errorf("PROFILE_NATIVE_CONSTANT_UNBOUND_OR_AMBIGUOUS")
				}
				found = value.Values[index]
			}
		}
	}
	return found, found != nil, nil
}

func assignProfileCaseValue(field reflect.Value, value constant.Value) error {
	if value == nil {
		return fmt.Errorf("PROFILE_NATIVE_VALUE_UNRESOLVED")
	}
	if field.Kind() == reflect.String && value.Kind() == constant.String {
		field.SetString(constant.StringVal(value))
		return nil
	}
	if field.Kind() == reflect.Bool && value.Kind() == constant.Bool {
		field.SetBool(constant.BoolVal(value))
		return nil
	}
	return fmt.Errorf("PROFILE_NATIVE_VALUE_TYPE_MISMATCH")
}
