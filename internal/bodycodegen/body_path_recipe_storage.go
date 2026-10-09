package bodycodegen

import (
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
)

func (b *recipeArena) appendRecipeExpression(node ast.Expr, value bodyplan.Expr) (int, error) {
	if b.expressionCount == len(b.expressions) {
		return 0, fmt.Errorf("recipe exceeds 128 expressions")
	}
	index := b.expressionCount
	b.expressions[index], b.expressionPositions[index] = value, node.Pos()
	b.expressionCount++
	if value.Kind == bodyplan.ExprInput {
		b.inputIndex, b.inputSeen = index, true
	}
	return index, nil
}

func (b *recipeArena) appendRecipeStatement(node ast.Stmt, value bodyplan.Stmt, expression ast.Expr, depth int) (int, error) {
	index, err := b.expression(expression, depth+1)
	if err != nil {
		return 0, err
	}
	value.Expr = index
	if b.statementCount == len(b.statements) {
		return 0, fmt.Errorf("recipe exceeds 128 statements")
	}
	index = b.statementCount
	b.statements[index], b.statementPositions[index] = value, node.Pos()
	b.statementCount++
	return index, nil
}

func (b *recipeArena) recipeElseSequence(node ast.Stmt, depth int) ([]int, error) {
	if node == nil {
		return nil, nil
	}
	var nodes []ast.Stmt
	switch otherwise := node.(type) {
	case *ast.BlockStmt:
		nodes = otherwise.List
	case *ast.IfStmt:
		// An else-if keeps its original positions and source-order selectors.
		nodes = []ast.Stmt{otherwise}
	default:
		return nil, fmt.Errorf("recipe else requires a block or if")
	}
	return b.sequence(nodes, depth+1)
}
