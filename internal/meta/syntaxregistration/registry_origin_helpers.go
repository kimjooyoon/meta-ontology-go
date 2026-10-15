package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (walk *registryOriginWalk) helper(name string) ([]*ast.CompositeLit, error) {
	if walk.functions[name] {
		return nil, fmt.Errorf("native case helper is cyclic or repeated: %s", name)
	}
	walk.functions[name] = true
	function, err := walk.source.function(name)
	if err != nil {
		return nil, err
	}
	if function.Body == nil || function.Type.TypeParams != nil ||
		function.Type.Params.NumFields() != 0 || function.Type.Results.NumFields() != 1 ||
		len(function.Type.Results.List[0].Names) != 0 ||
		!caseRegistrySlice(function.Type.Results.List[0].Type) {
		return nil, fmt.Errorf("native case helper has an unsupported signature: %s", name)
	}
	body := function.Body.List
	if len(body) == 1 {
		returned, ok := body[0].(*ast.ReturnStmt)
		if !ok || len(returned.Results) != 1 {
			return nil, fmt.Errorf("native case helper has no direct return: %s", name)
		}
		return walk.expression(returned.Results[0])
	}
	if len(body) < 3 {
		return nil, fmt.Errorf("native case helper has no bounded assembly: %s", name)
	}
	first, ok := body[0].(*ast.AssignStmt)
	if !ok || first.Tok != token.DEFINE || len(first.Lhs) != 1 || len(first.Rhs) != 1 {
		return nil, fmt.Errorf("native case assembly must declare one local sequence")
	}
	local, ok := first.Lhs[0].(*ast.Ident)
	if !ok || local.Name == "_" || local.Name == "append" {
		return nil, fmt.Errorf("native case assembly local is unresolved")
	}
	leaves, err := walk.expression(first.Rhs[0])
	if err != nil {
		return nil, err
	}
	for _, statement := range body[1 : len(body)-1] {
		extra, err := walk.appended(statement, local.Name)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, extra...)
	}
	last, ok := body[len(body)-1].(*ast.ReturnStmt)
	if !ok || len(last.Results) != 1 || !namedRegistryType(last.Results[0], local.Name) {
		return nil, fmt.Errorf("native case assembly must return its complete sequence")
	}
	return leaves, nil
}

func (walk *registryOriginWalk) appended(statement ast.Stmt, local string) ([]*ast.CompositeLit, error) {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 ||
		!namedRegistryType(assignment.Lhs[0], local) {
		return nil, fmt.Errorf("native case assembly has a conditional or unrelated write")
	}
	call, ok := assignment.Rhs[0].(*ast.CallExpr)
	if !ok || !namedRegistryType(call.Fun, "append") || len(call.Args) != 2 ||
		!call.Ellipsis.IsValid() || !namedRegistryType(call.Args[0], local) {
		return nil, fmt.Errorf("native case assembly requires ordered slice append")
	}
	if extra, ok := call.Args[1].(*ast.CallExpr); ok && namedRegistryType(extra.Fun, local) {
		return nil, fmt.Errorf("native case helper is shadowed by the assembly local")
	}
	for _, declaration := range walk.source.file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Recv == nil && declaration.Name.Name == "append" {
				return nil, fmt.Errorf("native case assembly append builtin is shadowed")
			}
		case *ast.GenDecl:
			for _, spec := range declaration.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range value.Names {
						if name.Name == "append" {
							return nil, fmt.Errorf("native case assembly append builtin is shadowed")
						}
					}
				}
			}
		}
	}
	return walk.expression(call.Args[1])
}
