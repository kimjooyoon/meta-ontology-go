package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
)

func pureEvaluatorFunctions(file *ast.File, information types.Info) map[types.Object]*ast.FuncDecl {
	functions := map[types.Object]*ast.FuncDecl{}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			functions[information.Defs[function.Name]] = function
		}
	}
	return functions
}

func (e *integerBodyEvaluator) evaluatePureCall(call *ast.CallExpr) (any, error) {
	name, ok := call.Fun.(*ast.Ident)
	if !ok || e.callDepth >= 16 {
		return nil, fmt.Errorf("pure call exceeds its named-function or depth bound")
	}
	function := e.functions[e.information.Uses[name]]
	if function == nil || len(function.Type.Params.List) != len(call.Args) {
		return nil, fmt.Errorf("pure call %q has no matching typed function", name.Name)
	}
	if e.callCount == nil {
		e.callCount = new(int)
	}
	*e.callCount++
	if *e.callCount > pureCallLimit {
		return nil, fmt.Errorf("pure call evaluation exceeds %d calls", pureCallLimit)
	}
	frame := integerBodyEvaluator{context: e.context, information: e.information,
		functions: e.functions, callDepth: e.callDepth + 1, callCount: e.callCount, environment: map[types.Object]any{}}
	if err := e.bindPureCallArguments(call, function, &frame); err != nil {
		return nil, err
	}
	value, returned, err := frame.evaluateBlock(function.Body)
	if err != nil {
		return nil, fmt.Errorf("pure call %q: %w", name.Name, err)
	}
	if !returned {
		return nil, fmt.Errorf("pure call %q did not return", name.Name)
	}
	output := e.information.Types[function.Type.Results.List[0].Type].Type
	return coerceBodyValue(value, output)
}

func (e *integerBodyEvaluator) bindPureCallArguments(call *ast.CallExpr, function *ast.FuncDecl, frame *integerBodyEvaluator) error {
	for i, argument := range call.Args {
		value, err := e.evaluateExpression(argument)
		if err != nil {
			return err
		}
		parameter := e.information.Defs[function.Type.Params.List[i].Names[0]]
		value, err = coerceBodyValue(value, parameter.Type())
		if err != nil {
			return err
		}
		frame.environment[parameter] = value
	}
	return nil
}
