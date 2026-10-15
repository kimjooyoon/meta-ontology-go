package bodycodegen

import (
	"go/ast"
	"strconv"
)

func (r *valueOriginReader) inputs(expr ast.Expr, env valueOriginEnv) []string {
	if expr == nil || !r.step() {
		return nil
	}
	var result []string
	add := func(value ast.Expr) { result = joinValueOriginIDs(result, r.inputs(value, env)) }
	switch value := expr.(type) {
	case *ast.Ident:
		if value.Obj == nil && (value.Name == "true" || value.Name == "false" || value.Name == "nil") {
			return nil
		}
		if ids, found := env[value.Obj]; found {
			return append([]string(nil), ids...)
		}
		if value.Obj != nil && r.set.Position(value.Obj.Pos()).Offset < r.header {
			return []string{r.node("input_reference", value.Name, value, nil, nil)}
		}
		r.reasons["VARIABLE_DEFINITION_UNAVAILABLE"] = true
		return []string{r.node("unresolved_reference", value.Name, value, nil, nil)}
	case *ast.BasicLit:
	case *ast.BinaryExpr:
		add(value.X)
		add(value.Y)
	case *ast.UnaryExpr:
		add(value.X)
		if value.Op.String() == "&" {
			r.reasons["ALIASES_NOT_EXPANDED"] = true
		}
	case *ast.ParenExpr:
		add(value.X)
	case *ast.SelectorExpr:
		if r.calls != nil {
			return r.access("field_access", value.Sel.Name, value, value.X, env)
		}
		add(value.X)
	case *ast.IndexExpr:
		if r.calls != nil {
			return r.access("index_access", "", value, value.X, env,
				valueAccessOperand{"index_operand", value.Index})
		}
		add(value.X)
		add(value.Index)
	case *ast.SliceExpr:
		if r.calls != nil {
			return r.access("slice_access", "", value, value.X, env,
				valueAccessOperand{"slice_low", value.Low},
				valueAccessOperand{"slice_high", value.High},
				valueAccessOperand{"slice_max", value.Max})
		}
		add(value.X)
		add(value.Low)
		add(value.High)
		add(value.Max)
	case *ast.CallExpr:
		var arguments []string
		for i, argument := range value.Args {
			dependencies := r.inputs(argument, env)
			result = joinValueOriginIDs(result, dependencies)
			if r.calls != nil {
				arguments = append(arguments, r.node("call_argument", strconv.Itoa(i), argument, dependencies, nil))
			}
		}
		if receiver, ok := value.Fun.(*ast.SelectorExpr); ok {
			add(receiver.X)
		}
		if function, ok := value.Fun.(*ast.Ident); ok && function.Obj != nil {
			add(function)
		}
		// The source call remains located, while another activity's body is a
		// separate origin scope and is not inferred from the returned value.
		id := r.node("call_result", "", value, result, nil)
		if r.calls != nil && id != "" {
			name := ""
			if function, ok := value.Fun.(*ast.Ident); ok && function.Obj == nil {
				name = function.Name
			}
			r.calls[id] = BodyValueCall{NodeID: id, Callee: name, Arguments: arguments}
		}
		result = []string{id}
		r.reasons["CALLEE_BODY_NOT_EXPANDED"] = true
	case *ast.CompositeLit:
		for _, element := range value.Elts {
			if pair, ok := element.(*ast.KeyValueExpr); ok {
				add(pair.Value)
			} else {
				add(element)
			}
		}
	default:
		r.reasons["EXPRESSION_FORM_UNRESOLVED"] = true
	}
	return result
}

func valueOriginRoot(expr ast.Expr) *ast.Ident {
	switch value := expr.(type) {
	case *ast.Ident:
		return value
	case *ast.SelectorExpr:
		return valueOriginRoot(value.X)
	case *ast.IndexExpr:
		return valueOriginRoot(value.X)
	case *ast.ParenExpr:
		return valueOriginRoot(value.X)
	default:
		return nil
	}
}
