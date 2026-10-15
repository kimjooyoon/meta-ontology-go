package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

func readHistoryCases(source *goSource, selection *ast.SwitchStmt, targets map[int]int, returns bool) error {
	for _, statement := range selection.Body.List {
		clause, ok := statement.(*ast.CaseClause)
		if !ok || len(clause.List) == 0 || len(clause.Body) != 1 {
			return fmt.Errorf("historical syntax target has an unsupported clause")
		}
		target, err := historyCaseTarget(source, clause.Body[0], returns)
		if err != nil {
			return err
		}
		for _, expression := range clause.List {
			version, literal := integer(expression)
			if !literal || version < 1 {
				return fmt.Errorf("historical syntax version is not a positive literal")
			}
			if _, exists := targets[version]; exists {
				return fmt.Errorf("historical syntax version is repeated")
			}
			targets[version] = target
		}
	}
	return nil
}

func historyCaseTarget(source *goSource, statement ast.Stmt, returns bool) (int, error) {
	var expression ast.Expr
	if returns {
		result, ok := statement.(*ast.ReturnStmt)
		if !ok || len(result.Results) != 2 || source.text(result.Results[1]) != "true" {
			return 0, fmt.Errorf("historical syntax target return differs")
		}
		expression = result.Results[0]
	} else {
		assignment, ok := statement.(*ast.AssignStmt)
		if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 || source.text(assignment.Lhs[0]) != "expected[0].Target" {
			return 0, fmt.Errorf("historical syntax target ownership differs")
		}
		expression = assignment.Rhs[0]
	}
	target, literal := integer(expression)
	if !literal || target < 1 {
		return 0, fmt.Errorf("historical syntax target is not a positive literal")
	}
	return target, nil
}
