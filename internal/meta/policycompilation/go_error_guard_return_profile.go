package policycompilation

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// V2 explicitly selects a local writer interface and a return-only handler.
// It does not change the v1 diagnostic-handler contract.
func parseGoReturnGuardProgram(parts []string) (goErrorGuardProgram, error) {
	if len(parts) != 6 {
		return goErrorGuardProgram{}, fmt.Errorf("return guard requires five fields")
	}
	values := map[string]string{}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || value == "" || values[key] != "" {
			return goErrorGuardProgram{}, fmt.Errorf("missing or duplicate return guard field")
		}
		switch key {
		case "function", "writer", "writer-type":
			if !token.IsIdentifier(value) || value == "_" {
				return goErrorGuardProgram{}, fmt.Errorf("return guard selector is not an identifier")
			}
		case "source", "handler":
			if !ValidDigest(value) {
				return goErrorGuardProgram{}, fmt.Errorf("return guard pin is not a digest")
			}
		default:
			return goErrorGuardProgram{}, fmt.Errorf("unknown return guard field %q", key)
		}
		values[key] = value
	}
	return goErrorGuardProgram{
		function: values["function"], writer: values["writer"],
		writerType: values["writer-type"], source: values["source"],
		handler: values["handler"], returnOnly: true,
	}, nil
}

func goReturnGuardWriterType(file *ast.File, name string) *ast.Object {
	var found *ast.Object
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}
		for _, raw := range group.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			if found != nil || spec.Assign.IsValid() || spec.TypeParams != nil || !goReturnGuardInterface(spec.Type) {
				return nil
			}
			found = spec.Name.Obj
		}
	}
	return found
}

func goReturnGuardInterface(expression ast.Expr) bool {
	value, ok := expression.(*ast.InterfaceType)
	if !ok || value.Methods == nil || len(value.Methods.List) != 1 {
		return false
	}
	method := value.Methods.List[0]
	signature, ok := method.Type.(*ast.FuncType)
	if len(method.Names) != 1 || method.Names[0].Name != "Write" || !ok ||
		signature.TypeParams != nil || signature.Params == nil || len(signature.Params.List) != 1 ||
		signature.Results == nil || len(signature.Results.List) != 2 {
		return false
	}
	input := signature.Params.List[0]
	slice, ok := input.Type.(*ast.ArrayType)
	if !ok || slice.Len != nil || len(input.Names) > 1 || !goGuardBuiltin(slice.Elt, "byte") {
		return false
	}
	for index, name := range []string{"int", "error"} {
		result := signature.Results.List[index]
		if len(result.Names) > 1 || !goGuardBuiltin(result.Type, name) {
			return false
		}
	}
	return true
}

func goReturnGuardParameter(function *ast.FuncDecl, writerType *ast.Object, name string) *ast.Object {
	for _, field := range function.Type.Params.List {
		if !goGuardObject(field.Type, writerType) {
			continue
		}
		for _, identifier := range field.Names {
			if identifier.Name == name {
				return identifier.Obj
			}
		}
	}
	return nil
}
