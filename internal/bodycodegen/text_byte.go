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
	if sequence, ok := value.(*recordBodySequence); ok {
		if sequence == nil {
			return nil, fmt.Errorf("collection index requires a present array")
		}
		offset, err := e.textOffset(index.Index, 0)
		if err != nil {
			return nil, err
		}
		if offset < 0 || offset >= int64(len(sequence.Values)) {
			return nil, fmt.Errorf("collection offset %d exceeds length %d", offset, len(sequence.Values))
		}
		return recordBodyScalarValue(sequence.Values[offset]), nil
	}
	text, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("index requires Text or an ordered collection, got %T", value)
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
