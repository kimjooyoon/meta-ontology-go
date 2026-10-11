package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

const countedLoopLimit = 1024

func countedLoopShape(loop *ast.ForStmt) (string, int, error) {
	initial, initialOK := loop.Init.(*ast.AssignStmt)
	post, postOK := loop.Post.(*ast.AssignStmt)
	condition, conditionOK := loop.Cond.(*ast.BinaryExpr)
	if !initialOK || !postOK || !conditionOK || initial.Tok != token.ASSIGN ||
		post.Tok != token.ASSIGN || len(initial.Lhs) != 1 || len(initial.Rhs) != 1 ||
		len(post.Lhs) != 1 || len(post.Rhs) != 1 || condition.Op != token.LSS {
		return "", 0, fmt.Errorf("for requires an existing counter initialized to 0, counter < literal bound, and counter = counter + 1")
	}
	counter, ok := initial.Lhs[0].(*ast.Ident)
	if !ok || counter.Name == "_" || !loopIntegerLiteral(initial.Rhs[0], 0) ||
		!loopCounter(condition.X, counter.Name) || !loopCounter(post.Lhs[0], counter.Name) {
		return "", 0, fmt.Errorf("for must use the same named counter from zero")
	}
	bound, ok := condition.Y.(*ast.BasicLit)
	if !ok || bound.Kind != token.INT {
		return "", 0, fmt.Errorf("for bound must be an integer literal from 1 to %d", countedLoopLimit)
	}
	limit, err := strconv.ParseInt(bound.Value, 0, 64)
	step, stepOK := post.Rhs[0].(*ast.BinaryExpr)
	if err != nil || limit < 1 || limit > countedLoopLimit || !stepOK || step.Op != token.ADD ||
		!loopCounter(step.X, counter.Name) || !loopIntegerLiteral(step.Y, 1) {
		return "", 0, fmt.Errorf("for requires a 1..%d bound and a unit forward step", countedLoopLimit)
	}
	if err := validateLoopCounterBody(loop.Body, counter.Name); err != nil {
		return "", 0, err
	}
	return counter.Name, int(limit), nil
}

func loopCounter(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name
}

func loopIntegerLiteral(expression ast.Expr, want int64) bool {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return false
	}
	value, err := strconv.ParseInt(literal.Value, 0, 64)
	return err == nil && value == want
}

func validateLoopCounterBody(body *ast.BlockStmt, counter string) error {
	var err error
	ast.Inspect(body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			err = fmt.Errorf("nested loops are outside the counted-loop profile")
		case *ast.AssignStmt:
			for _, target := range value.Lhs {
				if loopCounter(target, counter) {
					err = fmt.Errorf("for body cannot assign its counter %q", counter)
				}
			}
		case *ast.IncDecStmt:
			if loopCounter(value.X, counter) {
				err = fmt.Errorf("for body cannot update its counter %q", counter)
			}
		case *ast.ValueSpec:
			for _, name := range value.Names {
				if name.Name == counter {
					err = fmt.Errorf("for body cannot shadow its counter %q", counter)
				}
			}
		}
		return err == nil
	})
	return err
}

func validateCountedLoop(loop *ast.ForStmt, readonly, locals map[string]bool,
	allowGuardReturn bool, records []RecordType) (int, error) {
	counter, _, err := countedLoopShape(loop)
	if err != nil {
		return 0, err
	}
	if readonly[counter] || !locals[counter] {
		return 0, fmt.Errorf("for counter must be an existing mutable local")
	}
	protected := cloneNames(readonly)
	protected[counter] = true
	count, err := validateBlockInputs(loop.Body, protected, locals, allowGuardReturn, records...)
	return count + 2, err
}
