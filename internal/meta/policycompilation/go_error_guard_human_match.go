package policycompilation

import (
	"go/ast"
	"go/token"
)

type goHumanGuardMatch struct {
	calls     []*ast.CallExpr
	handler   *ast.BlockStmt
	errorName string
}

func findGoHumanOutputGuards(file *ast.File, profile goErrorGuardProgram) ([]goHumanGuardMatch, int) {
	matches := []goHumanGuardMatch{}
	functions := 0
	formatImport, writerImport := goGuardImport(file, "fmt"), goGuardImport(file, "io")
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != profile.function {
			continue
		}
		functions++
		if function.Recv != nil || function.Type.TypeParams != nil || function.Body == nil ||
			function.Type.Params == nil || function.Type.Results == nil ||
			len(function.Type.Results.List) != 1 || len(function.Type.Results.List[0].Names) > 1 ||
			!goGuardBuiltin(function.Type.Results.List[0].Type, "int") {
			continue
		}
		writer := goGuardWriterParameter(function, writerImport, profile.writer)
		mode := goHumanGuardMode(function, profile.mode)
		if writer == nil || mode == nil || formatImport == nil {
			continue
		}
		if match, found := matchGoHumanOutputGuard(function.Body, profile, formatImport, writer, mode); found {
			matches = append(matches, match)
		}
	}
	return matches, functions
}

func matchGoHumanOutputGuard(body *ast.BlockStmt, profile goErrorGuardProgram, imported *ast.ImportSpec, writer, mode *ast.Object) (goHumanGuardMatch, bool) {
	empty := goHumanGuardMatch{}
	if len(body.List) < 3 {
		return empty, false
	}
	human, ok := body.List[0].(*ast.IfStmt)
	if !ok || human.Init != nil || human.Else != nil || len(human.Body.List) < 2 {
		return empty, false
	}
	condition, ok := human.Cond.(*ast.UnaryExpr)
	if !ok || condition.Op != token.NOT || !goGuardObject(condition.X, mode) ||
		!goHumanGuardSameReturn(human.Body.List[len(human.Body.List)-1], body.List[len(body.List)-1]) {
		return empty, false
	}
	guard, ok := body.List[len(body.List)-2].(*ast.IfStmt)
	if !ok {
		return empty, false
	}
	errorID, ok := goHumanGuardHandler(guard, profile.handlerCall, writer)
	if !ok {
		return empty, false
	}
	calls, ok := collectGoHumanGuardCalls(human.Body.List[:len(human.Body.List)-1], imported, writer)
	if !ok || len(calls) == 0 {
		return empty, false
	}
	return goHumanGuardMatch{calls: calls, handler: guard.Body, errorName: errorID.Name}, true
}

func goHumanGuardHandler(guard *ast.IfStmt, callName string, writer *ast.Object) (*ast.Ident, bool) {
	if guard.Else != nil || len(guard.Body.List) != 1 {
		return nil, false
	}
	assignment, ok := guard.Init.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return nil, false
	}
	errorID, ok := assignment.Lhs[0].(*ast.Ident)
	call, callOK := assignment.Rhs[0].(*ast.CallExpr)
	if !ok || errorID.Name == "_" || errorID.Obj == nil || !callOK ||
		len(call.Args) != 2 || call.Ellipsis.IsValid() || !goGuardObject(call.Args[0], writer) {
		return nil, false
	}
	target, ok := call.Fun.(*ast.Ident)
	if !ok || target.Name != callName || target.Obj != nil && target.Obj.Kind != ast.Fun {
		return nil, false
	}
	condition, ok := guard.Cond.(*ast.BinaryExpr)
	result, resultOK := guard.Body.List[0].(*ast.ReturnStmt)
	if !ok || condition.Op != token.NEQ || !goGuardObject(condition.X, errorID.Obj) ||
		!goGuardBuiltin(condition.Y, "nil") || !resultOK || len(result.Results) != 1 ||
		!goGuardReturnValue(result.Results[0], errorID.Obj) {
		return nil, false
	}
	return errorID, true
}

func collectGoHumanGuardCalls(statements []ast.Stmt, imported *ast.ImportSpec, writer *ast.Object) ([]*ast.CallExpr, bool) {
	calls := []*ast.CallExpr{}
	for _, statement := range statements {
		if branch, ok := statement.(*ast.IfStmt); ok {
			if branch.Init != nil || branch.Else != nil || !goHumanGuardCondition(branch.Cond, writer) {
				return nil, false
			}
			nested, valid := collectGoHumanGuardCalls(branch.Body.List, imported, writer)
			if !valid {
				return nil, false
			}
			calls = append(calls, nested...)
			continue
		}
		call, ok := goGuardCall(statement)
		if !ok || !goGuardImported(call.Fun, imported, "Fprintf") || len(call.Args) < 2 ||
			!goGuardObject(call.Args[0], writer) {
			return nil, false
		}
		for _, argument := range call.Args[1:] {
			if goHumanGuardReferencesWriter(argument, writer) {
				return nil, false
			}
		}
		calls = append(calls, call)
	}
	return calls, true
}

func goHumanGuardReferencesWriter(node ast.Node, writer *ast.Object) bool {
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		if identifier, ok := node.(*ast.Ident); ok && identifier.Obj == writer {
			found = true
			return false
		}
		return true
	})
	return found
}

func goHumanGuardCondition(expression ast.Expr, writer *ast.Object) bool {
	safe := !goHumanGuardReferencesWriter(expression, writer)
	ast.Inspect(expression, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.CallExpr, *ast.FuncLit:
			safe = false
		case *ast.UnaryExpr:
			if value.Op == token.ARROW {
				safe = false
			}
		}
		return safe
	})
	return safe
}
