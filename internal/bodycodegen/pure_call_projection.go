package bodycodegen

import (
	"fmt"
	"go/ast"
)

func (p *bodyProjection) validatePureCalls(records []RecordType) error {
	allowed := map[string]bool{}
	for _, call := range p.calls {
		allowed[call.identity.Name] = true
		function, ok := findFunction(p.file, call.identity.Name)
		if !ok || !blockTerminates(function.Body) {
			return fmt.Errorf("pure call %q must return on every path", call.identity.Name)
		}
		readonly := parameterNames(call.parameters)
		count, err := validateBlockInputs(function.Body, readonly, readonly, true, records...)
		if err != nil {
			return fmt.Errorf("pure call %s: %w", call.identity.Name, err)
		}
		if err := validateRecordLiterals(function.Body, records); err != nil {
			return err
		}
		p.constructs += count
		p.units += semanticUnitCount(function.Body)
	}
	var err error
	ast.Inspect(p.file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			name, ok := call.Fun.(*ast.Ident)
			if !ok || (!allowed[name.Name] && !bodyPrimitiveName(name.Name)) {
				err = fmt.Errorf("unsupported call: callee must be a source-declared pure activity or a supported value primitive")
			}
		}
		return err == nil
	})
	return err
}
