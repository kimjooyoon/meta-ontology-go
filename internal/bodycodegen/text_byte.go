package bodycodegen

import (
	"fmt"
	"go/ast"
)

func (e *integerBodyEvaluator) evaluateTextByte(index *ast.IndexExpr) (any, error) {
	value, err := e.evaluateExpression(index.X)
	if err != nil {
		return nil, err
	}
	text, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("byte index requires Text, got %T", value)
	}
	offset, err := e.textOffset(index.Index, 0)
	if err != nil {
		return nil, err
	}
	if offset < 0 || offset >= int64(len(text)) {
		return nil, fmt.Errorf("Text byte offset %d exceeds length %d", offset, len(text))
	}
	return int64(text[offset]), nil
}
