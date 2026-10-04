package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func (e *integerBodyEvaluator) evaluateStatement(statement ast.Stmt) (any, bool, error) {
	switch value := statement.(type) {
	case *ast.DeclStmt:
		return nil, false, e.evaluateDeclaration(value)
	case *ast.AssignStmt:
		return nil, false, e.evaluateAssignment(value)
	case *ast.IfStmt:
		return e.evaluateConditional(value)
	case *ast.ReturnStmt:
		if len(value.Results) != 1 {
			return nil, false, fmt.Errorf("generated return must have one value")
		}
		returned, err := e.evaluateExpression(value.Results[0])
		return returned, true, err
	default:
		return nil, false, fmt.Errorf("unsupported generated statement %T", statement)
	}
}

func (e *integerBodyEvaluator) evaluateDeclaration(value *ast.DeclStmt) error {
	declaration, ok := value.Decl.(*ast.GenDecl)
	if !ok || declaration.Tok != token.VAR || len(declaration.Specs) != 1 {
		return fmt.Errorf("unsupported generated declaration %T", value.Decl)
	}
	spec, ok := declaration.Specs[0].(*ast.ValueSpec)
	if !ok || len(spec.Names) != 1 {
		return fmt.Errorf("unsupported generated value declaration")
	}
	object := e.information.Defs[spec.Names[0]]
	if object == nil {
		return fmt.Errorf("generated declaration has no typed binding")
	}
	initial, err := zeroBodyValue(object.Type())
	if err != nil {
		return err
	}
	if len(spec.Values) == 1 {
		initial, err = e.evaluateExpression(spec.Values[0])
		if err != nil {
			return err
		}
	}
	initial, err = coerceBodyValue(initial, object.Type())
	if err == nil {
		e.environment[object] = initial
	}
	return err
}

func zeroBodyValue(t types.Type) (any, error) {
	switch t.Underlying() {
	case types.Typ[types.Bool]:
		return false, nil
	case types.Typ[types.String]:
		return "", nil
	}
	if _, ok := t.Underlying().(*types.Struct); ok {
		return zeroRecordBodyValue(t)
	}
	return int64(0), nil
}

func (e *integerBodyEvaluator) evaluateAssignment(value *ast.AssignStmt) error {
	if len(value.Lhs) != 1 || len(value.Rhs) != 1 || value.Tok != token.ASSIGN {
		return fmt.Errorf("unsupported generated assignment")
	}
	name, ok := value.Lhs[0].(*ast.Ident)
	if !ok {
		return fmt.Errorf("unsupported generated assignment target")
	}
	assigned, err := e.evaluateExpression(value.Rhs[0])
	if err != nil || name.Name == "_" {
		return err
	}
	object := e.information.Uses[name]
	if object == nil {
		return fmt.Errorf("generated assignment has no typed binding")
	}
	assigned, err = coerceBodyValue(assigned, object.Type())
	if err == nil {
		e.environment[object] = assigned
	}
	return err
}

func (e *integerBodyEvaluator) evaluateConditional(value *ast.IfStmt) (any, bool, error) {
	condition, err := e.evaluateExpression(value.Cond)
	if err != nil {
		return nil, false, err
	}
	truth, ok := condition.(bool)
	if !ok {
		return nil, false, fmt.Errorf("generated if condition evaluated to %T", condition)
	}
	if truth {
		return e.evaluateBlock(value.Body)
	}
	switch otherwise := value.Else.(type) {
	case *ast.BlockStmt:
		return e.evaluateBlock(otherwise)
	case *ast.IfStmt:
		return e.evaluateConditional(otherwise)
	case nil:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("unsupported generated else %T", value.Else)
	}
}
