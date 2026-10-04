package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
)

func validateBlockInputs(block *ast.BlockStmt, readonly, inherited map[string]bool,
	allowGuardReturn bool, records ...RecordType) (int, error) {
	if block == nil {
		return 0, fmt.Errorf("activity body has no block")
	}
	count, locals := 0, cloneNames(inherited)
	for _, statement := range block.List {
		nested, err := validateBodyStatement(statement, readonly, locals, allowGuardReturn, records)
		if err != nil {
			return 0, err
		}
		count += 1 + nested
	}
	return count, nil
}

func validateBodyStatement(statement ast.Stmt, readonly, locals map[string]bool,
	allowGuardReturn bool, records []RecordType) (int, error) {
	switch value := statement.(type) {
	case *ast.DeclStmt:
		name, err := validateBodyLocal(value, readonly, locals, records)
		if err == nil {
			locals[name] = true
		}
		return 0, err
	case *ast.AssignStmt:
		if value.Tok != token.ASSIGN || len(value.Lhs) != 1 || len(value.Rhs) != 1 {
			return 0, fmt.Errorf("assignment requires one existing local and one value")
		}
		name, ok := assignmentLocalName(value.Lhs[0])
		if !ok || readonly[name] || !locals[name] {
			return 0, fmt.Errorf("assignment target must be an existing local")
		}
		return 0, validateExpression(value.Rhs[0])
	case *ast.IfStmt:
		return validateBodyIf(value, readonly, locals, allowGuardReturn, records)
	case *ast.ReturnStmt:
		if len(value.Results) != 1 {
			return 0, fmt.Errorf("return requires exactly one value")
		}
		return 0, validateExpression(value.Results[0])
	default:
		return 0, fmt.Errorf("unsupported activity statement %T", statement)
	}
}

func validateBodyIf(value *ast.IfStmt, readonly, locals map[string]bool,
	allowGuardReturn bool, records []RecordType) (int, error) {
	if value.Init != nil || (value.Else == nil && !allowGuardReturn) {
		return 0, fmt.Errorf("if requires a condition and an explicit else branch")
	}
	if err := validateExpression(value.Cond); err != nil {
		return 0, err
	}
	thenCount, err := validateBlockInputs(value.Body, readonly, locals, allowGuardReturn, records...)
	if err != nil {
		return 0, err
	}
	if value.Else == nil {
		return thenCount, nil
	}
	var otherwise *ast.BlockStmt
	switch branch := value.Else.(type) {
	case *ast.BlockStmt:
		otherwise = branch
	case *ast.IfStmt:
		otherwise = &ast.BlockStmt{List: []ast.Stmt{branch}}
	default:
		return 0, fmt.Errorf("else branch must be a block or if")
	}
	elseCount, err := validateBlockInputs(otherwise, readonly, locals, allowGuardReturn, records...)
	return thenCount + elseCount, err
}
