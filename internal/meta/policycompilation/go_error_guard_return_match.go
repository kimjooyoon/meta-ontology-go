package policycompilation

import (
	"go/ast"
	"go/token"
)

func findGoReturnGuards(file *ast.File, profile goErrorGuardProgram) ([]goErrorGuardMatch, int) {
	matches := []goErrorGuardMatch{}
	functions := 0
	imported := goGuardImport(file, "fmt")
	writerType := goReturnGuardWriterType(file, profile.writerType)
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
		writer := goReturnGuardParameter(function, writerType, profile.writer)
		if imported == nil || writer == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}
			block, ok := node.(*ast.BlockStmt)
			if ok && len(block.List) >= 3 {
				tail := block.List[len(block.List)-3:]
				if match, found := matchGoReturnGuard(tail, imported, writer); found {
					matches = append(matches, match)
				}
			}
			return true
		})
	}
	return matches, functions
}

func matchGoReturnGuard(tail []ast.Stmt, imported *ast.ImportSpec, writer *ast.Object) (goErrorGuardMatch, bool) {
	empty := goErrorGuardMatch{}
	parent, ok := tail[0].(*ast.IfStmt)
	if !ok || parent.Init != nil || parent.Else != nil || len(parent.Body.List) != 2 {
		return empty, false
	}
	guard, ok := parent.Body.List[0].(*ast.IfStmt)
	if !ok || guard.Else != nil || len(guard.Body.List) != 1 {
		return empty, false
	}
	errorID, ok := goReturnGuardError(guard, writer)
	if !ok {
		return empty, false
	}
	for _, statement := range []ast.Stmt{guard.Body.List[0], parent.Body.List[1], tail[2]} {
		result, ok := statement.(*ast.ReturnStmt)
		if !ok || len(result.Results) != 1 || !goGuardReturnValue(result.Results[0], errorID.Obj) {
			return empty, false
		}
	}
	call, ok := goGuardCall(tail[1])
	if !ok || !goGuardImported(call.Fun, imported, "Fprintf") ||
		len(call.Args) < 2 || !goGuardObject(call.Args[0], writer) {
		return empty, false
	}
	return goErrorGuardMatch{call: call, handler: guard.Body, errorName: errorID.Name}, true
}

func goReturnGuardError(guard *ast.IfStmt, writer *ast.Object) (*ast.Ident, bool) {
	assignment, ok := guard.Init.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 2 || len(assignment.Rhs) != 1 {
		return nil, false
	}
	blank, blankOK := assignment.Lhs[0].(*ast.Ident)
	errorID, errorOK := assignment.Lhs[1].(*ast.Ident)
	write, writeOK := assignment.Rhs[0].(*ast.CallExpr)
	if !blankOK || blank.Name != "_" || !errorOK || errorID.Obj == nil ||
		!writeOK || len(write.Args) != 1 || write.Ellipsis.IsValid() {
		return nil, false
	}
	method, methodOK := write.Fun.(*ast.SelectorExpr)
	condition, conditionOK := guard.Cond.(*ast.BinaryExpr)
	if !methodOK || method.Sel.Name != "Write" || !goGuardObject(method.X, writer) ||
		!conditionOK || condition.Op != token.NEQ || !goGuardObject(condition.X, errorID.Obj) ||
		!goGuardBuiltin(condition.Y, "nil") {
		return nil, false
	}
	return errorID, true
}
