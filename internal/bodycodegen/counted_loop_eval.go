package bodycodegen

import (
	"fmt"
	"go/ast"
)

func (e *integerBodyEvaluator) evaluateCountedLoop(loop *ast.ForStmt) (any, bool, error) {
	_, bound, err := countedLoopShape(loop)
	if err != nil {
		return nil, false, err
	}
	cleanup, err := e.initializeCountedLoop(loop.Init.(*ast.AssignStmt))
	if err != nil {
		return nil, false, err
	}
	defer cleanup()
	for range bound {
		value, err := e.evaluateExpression(loop.Cond)
		if err != nil {
			return nil, false, err
		}
		truth, ok := value.(bool)
		if !ok || !truth {
			return nil, false, fmt.Errorf("counted-loop condition diverges before its declared bound")
		}
		result, returned, err := e.evaluateBlock(loop.Body)
		if err != nil || returned {
			return result, returned, err
		}
		if err := e.evaluateCountedPost(loop.Post); err != nil {
			return nil, false, err
		}
	}
	value, err := e.evaluateExpression(loop.Cond)
	if err != nil {
		return nil, false, err
	}
	if truth, ok := value.(bool); !ok || truth {
		return nil, false, fmt.Errorf("counted-loop condition exceeds its declared bound")
	}
	return nil, false, nil
}
