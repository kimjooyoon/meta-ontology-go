package policycompilation

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// V3 selects an io.Writer human-first branch and its exact number of writes.
// It neither widens v1/v2 matching nor authorizes source application.
func parseGoHumanGuardProgram(parts []string) (goErrorGuardProgram, error) {
	if len(parts) != 8 {
		return goErrorGuardProgram{}, fmt.Errorf("human guard requires seven fields")
	}
	values := map[string]string{}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || value == "" || values[key] != "" {
			return goErrorGuardProgram{}, fmt.Errorf("missing or duplicate human guard field")
		}
		switch key {
		case "function", "writer", "mode", "handler-call":
			if !token.IsIdentifier(value) || value == "_" {
				return goErrorGuardProgram{}, fmt.Errorf("human guard selector is not an identifier")
			}
		case "source", "handler":
			if !ValidDigest(value) {
				return goErrorGuardProgram{}, fmt.Errorf("human guard pin is not a digest")
			}
		case "writes":
			count, err := strconv.Atoi(value)
			if err != nil || count <= 0 || strconv.Itoa(count) != value {
				return goErrorGuardProgram{}, fmt.Errorf("human guard writes must be a canonical positive integer")
			}
		default:
			return goErrorGuardProgram{}, fmt.Errorf("unknown human guard field %q", key)
		}
		values[key] = value
	}
	count, _ := strconv.Atoi(values["writes"])
	if values["writer"] == values["mode"] {
		return goErrorGuardProgram{}, fmt.Errorf("human guard writer and mode must differ")
	}
	return goErrorGuardProgram{
		function: values["function"], writer: values["writer"], mode: values["mode"],
		handlerCall: values["handler-call"], source: values["source"], handler: values["handler"],
		writeCount: count, humanBranch: true,
	}, nil
}

func goHumanGuardMode(function *ast.FuncDecl, name string) *ast.Object {
	for _, field := range function.Type.Params.List {
		if !goGuardBuiltin(field.Type, "bool") {
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

func goHumanGuardSameReturn(left, right ast.Stmt) bool {
	a, ok := left.(*ast.ReturnStmt)
	b, other := right.(*ast.ReturnStmt)
	if !ok || !other || len(a.Results) != 1 || len(b.Results) != 1 {
		return false
	}
	switch value := a.Results[0].(type) {
	case *ast.Ident:
		other, ok := b.Results[0].(*ast.Ident)
		return ok && value.Name == other.Name && value.Obj == other.Obj
	case *ast.BasicLit:
		other, ok := b.Results[0].(*ast.BasicLit)
		return ok && value.Kind == token.INT && value.Kind == other.Kind && value.Value == other.Value
	}
	return false
}
