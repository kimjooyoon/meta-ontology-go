package policycompilation

import (
	"go/ast"
	"go/token"
	"strconv"
)

func findGoErrorGuards(file *ast.File, profile goErrorGuardProgram) ([]goErrorGuardMatch, int) {
	if profile.returnOnly {
		return findGoReturnGuards(file, profile)
	}
	matches := []goErrorGuardMatch{}
	functions := 0
	formatImport, writerImport := goGuardImport(file, "fmt"), goGuardImport(file, "io")
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != profile.function {
			continue
		}
		functions++
		if function.Recv != nil || function.Type.TypeParams != nil || function.Body == nil ||
			function.Type.Results == nil || len(function.Type.Results.List) != 1 ||
			!goGuardBuiltin(function.Type.Results.List[0].Type, "int") {
			continue
		}
		writer := goGuardWriterParameter(function, writerImport, profile.writer)
		diagnostic := goGuardWriterParameter(function, writerImport, profile.diagnostic)
		if writer == nil || diagnostic == nil || writer == diagnostic || formatImport == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}
			parent, ok := node.(*ast.IfStmt)
			if ok {
				if match, found := matchGoErrorGuard(parent, formatImport, writer, diagnostic); found {
					matches = append(matches, match)
				}
			}
			return true
		})
	}
	return matches, functions
}

func matchGoErrorGuard(parent *ast.IfStmt, imported *ast.ImportSpec, writer, diagnostic *ast.Object) (goErrorGuardMatch, bool) {
	empty := goErrorGuardMatch{}
	target, ok := parent.Else.(*ast.IfStmt)
	if !ok || parent.Init != nil || target.Init != nil || len(parent.Body.List) != 1 || len(target.Body.List) != 1 {
		return empty, false
	}
	guard, ok := parent.Body.List[0].(*ast.IfStmt)
	if !ok || guard.Else != nil || len(guard.Body.List) != 2 {
		return empty, false
	}
	assignment, ok := guard.Init.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 2 || len(assignment.Rhs) != 1 {
		return empty, false
	}
	blank, blankOK := assignment.Lhs[0].(*ast.Ident)
	errorID, errorOK := assignment.Lhs[1].(*ast.Ident)
	write, writeOK := assignment.Rhs[0].(*ast.CallExpr)
	if !blankOK || blank.Name != "_" || !errorOK || errorID.Obj == nil || !writeOK {
		return empty, false
	}
	method, methodOK := write.Fun.(*ast.SelectorExpr)
	condition, conditionOK := guard.Cond.(*ast.BinaryExpr)
	if !methodOK || method.Sel.Name != "Write" || !goGuardObject(method.X, writer) ||
		!conditionOK || condition.Op != token.NEQ || !goGuardObject(condition.X, errorID.Obj) ||
		!goGuardBuiltin(condition.Y, "nil") {
		return empty, false
	}
	diagnosticCall, diagnosticOK := goGuardCall(guard.Body.List[0])
	result, resultOK := guard.Body.List[1].(*ast.ReturnStmt)
	if !diagnosticOK || !goGuardImported(diagnosticCall.Fun, imported, "Fprintf") ||
		len(diagnosticCall.Args) != 3 || !goGuardObject(diagnosticCall.Args[0], diagnostic) ||
		!goGuardObject(diagnosticCall.Args[2], errorID.Obj) || !resultOK || len(result.Results) != 1 {
		return empty, false
	}
	format, formatOK := diagnosticCall.Args[1].(*ast.BasicLit)
	if !formatOK || format.Kind != token.STRING || !goGuardReturnValue(result.Results[0], errorID.Obj) {
		return empty, false
	}
	call, ok := goGuardCall(target.Body.List[0])
	if !ok || !goGuardImported(call.Fun, imported, "Fprintf") || len(call.Args) < 2 ||
		!goGuardObject(call.Args[0], writer) {
		return empty, false
	}
	return goErrorGuardMatch{call: call, handler: guard.Body, errorName: errorID.Name}, true
}

func goGuardImport(file *ast.File, path string) *ast.ImportSpec {
	var found *ast.ImportSpec
	for _, imported := range file.Imports {
		value, err := strconv.Unquote(imported.Path.Value)
		if err != nil || value != path {
			continue
		}
		if found != nil || imported.Name != nil && (imported.Name.Name == "." || imported.Name.Name == "_") {
			return nil
		}
		found = imported
	}
	return found
}

func goGuardImported(expression ast.Expr, imported *ast.ImportSpec, member string) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || imported == nil || selector.Sel.Name != member {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	name, err := strconv.Unquote(imported.Path.Value)
	if imported.Name != nil {
		name = imported.Name.Name
	}
	return ok && err == nil && identifier.Name == name &&
		(identifier.Obj == nil || identifier.Obj.Kind == ast.Pkg && identifier.Obj.Decl == imported)
}

func goGuardWriterParameter(function *ast.FuncDecl, imported *ast.ImportSpec, name string) *ast.Object {
	for _, field := range function.Type.Params.List {
		if !goGuardImported(field.Type, imported, "Writer") {
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

func goGuardObject(expression ast.Expr, object *ast.Object) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && object != nil && identifier.Obj == object
}

func goGuardBuiltin(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name && identifier.Obj == nil
}

func goGuardCall(statement ast.Stmt) (*ast.CallExpr, bool) {
	expression, ok := statement.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	call, ok := expression.X.(*ast.CallExpr)
	return call, ok
}

func goGuardReturnValue(expression ast.Expr, errorObject *ast.Object) bool {
	switch value := expression.(type) {
	case *ast.BasicLit:
		return value.Kind == token.INT
	case *ast.Ident:
		return value.Name != "_" && value.Obj != errorObject
	}
	return false
}
