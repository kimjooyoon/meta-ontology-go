package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const RepeatedFieldLimit = syntax.RepeatedFieldLimit

func collectionLoopShape(loop *ast.RangeStmt) (string, error) {
	key, keyOK := loop.Key.(*ast.Ident)
	value, valueOK := loop.Value.(*ast.Ident)
	if loop.Tok != token.DEFINE || !keyOK || key.Name != "_" || !valueOK || value.Name == "_" {
		return "", fmt.Errorf("collection iteration requires a new named value and an omitted index")
	}
	if err := validateCollectionNesting(loop); err != nil {
		return "", err
	}
	return value.Name, validateExpression(loop.X)
}

func validateCollectionLoop(loop *ast.RangeStmt, readonly, locals map[string]bool,
	allowGuardReturn bool, records []RecordType) (int, error) {
	name, err := collectionLoopShape(loop)
	if err != nil {
		return 0, err
	}
	if readonly[name] || locals[name] {
		return 0, fmt.Errorf("collection element %q is already bound", name)
	}
	protected := cloneNames(readonly)
	protected[name] = true
	return validateBlockInputs(loop.Body, protected, locals, allowGuardReturn, records...)
}

func validateCollectionLoopTypes(file *ast.File, info *types.Info) error {
	var err error
	ast.Inspect(file, func(node ast.Node) bool {
		loop, ok := node.(*ast.RangeStmt)
		if !ok || err != nil {
			return err == nil
		}
		value := info.Types[loop.X].Type
		if value == nil {
			err = fmt.Errorf("collection iteration has no source type")
			return false
		}
		sequence, ok := value.Underlying().(*types.Slice)
		if !ok {
			err = fmt.Errorf("collection iteration requires a declared repeated scalar field")
			return false
		}
		_, recordElement := sequence.Elem().Underlying().(*types.Struct)
		if _, optional, kindErr := recordBodyScalarKind(sequence.Elem()); (kindErr != nil && !recordElement) || optional {
			err = fmt.Errorf("collection elements must be Integer, Text or Boolean values")
		}
		return err == nil
	})
	return err
}

func (e *integerBodyEvaluator) evaluateCollectionLoop(loop *ast.RangeStmt) (any, bool, error) {
	if _, err := collectionLoopShape(loop); err != nil {
		return nil, false, err
	}
	raw, err := e.evaluateExpression(loop.X)
	if err != nil {
		return nil, false, err
	}
	sequence, ok := raw.(*recordBodySequence)
	if !ok || sequence == nil || len(sequence.Values) > RepeatedFieldLimit {
		return nil, false, fmt.Errorf("collection iteration exceeds its value profile")
	}
	name := loop.Value.(*ast.Ident)
	object := e.information.Defs[name]
	if object == nil {
		return nil, false, fmt.Errorf("collection element has no typed binding")
	}
	defer delete(e.environment, object)
	for _, value := range sequence.Values {
		e.environment[object] = recordBodyScalarValue(value)
		result, returned, err := e.evaluateBlock(loop.Body)
		if returned || err != nil {
			return result, returned, err
		}
	}
	return nil, false, nil
}
