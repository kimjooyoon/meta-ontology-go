package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (e *integerBodyEvaluator) initializeCountedLoop(initial *ast.AssignStmt) (func(), error) {
	cleanup := func() {}
	if initial.Tok == token.ASSIGN {
		return cleanup, e.evaluateAssignment(initial)
	}
	name, ok := initial.Lhs[0].(*ast.Ident)
	if !ok {
		return cleanup, fmt.Errorf("counted counter has no identifier")
	}
	object := e.information.Defs[name]
	if object == nil {
		return cleanup, fmt.Errorf("counted counter has no typed declaration")
	}
	// countedLoopShape has already restricted this initializer to literal zero.
	value, err := coerceBodyValue(int64(0), object.Type())
	if err != nil {
		return cleanup, err
	}
	e.environment[object] = value
	return func() { delete(e.environment, object) }, nil
}

func (e *integerBodyEvaluator) evaluateCountedPost(statement ast.Stmt) error {
	if assignment, ok := statement.(*ast.AssignStmt); ok {
		return e.evaluateAssignment(assignment)
	}
	update, ok := statement.(*ast.IncDecStmt)
	if !ok || update.Tok != token.INC {
		return fmt.Errorf("counted post must use a unit forward step")
	}
	name, ok := update.X.(*ast.Ident)
	if !ok {
		return fmt.Errorf("counted post has no counter")
	}
	object := e.information.Uses[name]
	value, ok := e.environment[object].(int64)
	if !ok {
		return fmt.Errorf("counted post has no Integer counter value")
	}
	e.environment[object] = value + 1
	return nil
}
