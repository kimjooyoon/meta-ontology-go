package bodycodegen

import (
	"context"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
)

type integerBodyEvaluator struct {
	context     context.Context
	information types.Info
	environment map[types.Object]any
}

// evaluateIntegerCases interprets only the already typechecked, pure integer
// body profile. It is used to score bounded experimental candidates without
// executing arbitrary generated programs.
func evaluateIntegerCases(
	source []byte,
	activity string,
	cases []IRBodyFillTestCase,
) ([]IRBodyFillCaseResult, int, error) {
	return evaluateIntegerCasesContext(context.Background(), source, activity, cases)
}

func evaluateIntegerCasesContext(ctx context.Context, source []byte, activity string,
	cases []IRBodyFillTestCase,
) ([]IRBodyFillCaseResult, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", source, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, 0, fmt.Errorf("parse generated Go: %w", err)
	}
	function, ok := findFunction(file, activity)
	if !ok {
		return nil, 0, fmt.Errorf("generated function %q was not found", activity)
	}
	information := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	configuration := types.Config{}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if _, err := configuration.Check(file.Name.Name, fset, []*ast.File{file}, &information); err != nil {
		return nil, 0, fmt.Errorf("typecheck integer evaluator input: %w", err)
	}
	if function.Type.Params == nil || len(function.Type.Params.List) != 1 ||
		len(function.Type.Params.List[0].Names) != 1 {
		return nil, 0, fmt.Errorf("integer evaluator requires one named input")
	}
	input := information.Defs[function.Type.Params.List[0].Names[0]]
	if input == nil || input.Type() != types.Typ[types.Int64] {
		return nil, 0, fmt.Errorf("integer evaluator input must be int64")
	}
	evaluator := integerBodyEvaluator{context: ctx, information: information}
	results := make([]IRBodyFillCaseResult, 0, len(cases))
	passed := 0
	for _, testCase := range cases {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		evaluator.environment = map[types.Object]any{input: testCase.Input}
		value, returned, err := evaluator.evaluateBlock(function.Body)
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		if err != nil {
			return nil, 0, fmt.Errorf("input %d: %w", testCase.Input, err)
		}
		if !returned {
			return nil, 0, fmt.Errorf("input %d did not return", testCase.Input)
		}
		actual, ok := value.(int64)
		if !ok {
			return nil, 0, fmt.Errorf("input %d returned %T, want int64", testCase.Input, value)
		}
		match := actual == testCase.Expected
		if match {
			passed++
		}
		results = append(results, IRBodyFillCaseResult{
			Input: testCase.Input, Expected: testCase.Expected, Actual: actual, Passed: match,
		})
	}
	return results, passed, nil
}

func (e *integerBodyEvaluator) evaluateBlock(block *ast.BlockStmt) (any, bool, error) {
	for _, statement := range block.List {
		if e.context != nil {
			if err := e.context.Err(); err != nil {
				return nil, false, err
			}
		}
		switch value := statement.(type) {
		case *ast.DeclStmt:
			declaration, ok := value.Decl.(*ast.GenDecl)
			if !ok || declaration.Tok != token.VAR || len(declaration.Specs) != 1 {
				return nil, false, fmt.Errorf("unsupported generated declaration %T", value.Decl)
			}
			spec, ok := declaration.Specs[0].(*ast.ValueSpec)
			if !ok || len(spec.Names) != 1 {
				return nil, false, fmt.Errorf("unsupported generated value declaration")
			}
			object := e.information.Defs[spec.Names[0]]
			if object == nil {
				return nil, false, fmt.Errorf("generated declaration has no typed binding")
			}
			var initial any = int64(0)
			if object.Type().Underlying() == types.Typ[types.Bool] {
				initial = false
			} else if object.Type().Underlying() == types.Typ[types.String] {
				initial = ""
			}
			if len(spec.Values) == 1 {
				var err error
				initial, err = e.evaluateExpression(spec.Values[0])
				if err != nil {
					return nil, false, err
				}
			}
			initial, err := coerceBodyValue(initial, object.Type())
			if err != nil {
				return nil, false, err
			}
			e.environment[object] = initial
		case *ast.AssignStmt:
			if len(value.Lhs) != 1 || len(value.Rhs) != 1 || value.Tok != token.ASSIGN {
				return nil, false, fmt.Errorf("unsupported generated assignment")
			}
			name, ok := value.Lhs[0].(*ast.Ident)
			if !ok {
				return nil, false, fmt.Errorf("unsupported generated assignment target")
			}
			assigned, err := e.evaluateExpression(value.Rhs[0])
			if err != nil {
				return nil, false, err
			}
			if name.Name == "_" {
				continue
			}
			object := e.information.Uses[name]
			if object == nil {
				return nil, false, fmt.Errorf("generated assignment has no typed binding")
			}
			assigned, err = coerceBodyValue(assigned, object.Type())
			if err != nil {
				return nil, false, err
			}
			e.environment[object] = assigned
		case *ast.IfStmt:
			condition, err := e.evaluateExpression(value.Cond)
			if err != nil {
				return nil, false, err
			}
			truth, ok := condition.(bool)
			if !ok {
				return nil, false, fmt.Errorf("generated if condition evaluated to %T", condition)
			}
			if truth {
				result, returned, err := e.evaluateBlock(value.Body)
				if err != nil || returned {
					return result, returned, err
				}
			} else {
				switch otherwise := value.Else.(type) {
				case *ast.BlockStmt:
					result, returned, err := e.evaluateBlock(otherwise)
					if err != nil || returned {
						return result, returned, err
					}
				case *ast.IfStmt:
					result, returned, err := e.evaluateBlock(&ast.BlockStmt{List: []ast.Stmt{otherwise}})
					if err != nil || returned {
						return result, returned, err
					}
				}
			}
		case *ast.ReturnStmt:
			if len(value.Results) != 1 {
				return nil, false, fmt.Errorf("generated return must have one value")
			}
			returned, err := e.evaluateExpression(value.Results[0])
			return returned, true, err
		default:
			return nil, false, fmt.Errorf("unsupported generated statement %T", statement)
		}
	}
	return nil, false, nil
}

func (e *integerBodyEvaluator) evaluateExpression(expression ast.Expr) (any, error) {
	if e.context != nil {
		if err := e.context.Err(); err != nil {
			return nil, err
		}
	}
	typed := e.information.Types[expression]
	if typed.Value != nil {
		// Go folds constants with arbitrary precision before conversion to a
		// runtime type. This also handles MinInt64 and shadowed true/false.
		var value any
		switch typed.Value.Kind() {
		case constant.Bool:
			value = constant.BoolVal(typed.Value)
		case constant.String:
			value = constant.StringVal(typed.Value)
		default:
			integerConstant := constant.ToInt(typed.Value)
			if integerConstant.Kind() != constant.Int {
				return nil, fmt.Errorf("constant %s is outside the integer evaluator profile", typed.Value)
			}
			integer, exact := constant.Int64Val(integerConstant)
			if !exact {
				return nil, fmt.Errorf("constant %s is outside the integer evaluator profile", typed.Value)
			}
			value = integer
		}
		return coerceBodyValue(value, typed.Type)
	}
	switch value := expression.(type) {
	case *ast.Ident:
		result, ok := e.environment[e.information.Uses[value]]
		if !ok {
			return nil, fmt.Errorf("unbound identifier %q", value.Name)
		}
		return result, nil
	case *ast.ParenExpr:
		return e.evaluateExpression(value.X)
	case *ast.UnaryExpr:
		operand, err := e.evaluateExpression(value.X)
		if err != nil {
			return nil, err
		}
		switch value.Op {
		case token.SUB:
			number, ok := operand.(int64)
			if !ok {
				return nil, fmt.Errorf("unary minus operand is %T", operand)
			}
			return coerceBodyValue(-number, typed.Type)
		case token.NOT:
			truth, ok := operand.(bool)
			if !ok {
				return nil, fmt.Errorf("logical not operand is %T", operand)
			}
			return !truth, nil
		default:
			return nil, fmt.Errorf("unsupported generated unary operator %s", value.Op)
		}
	case *ast.BinaryExpr:
		left, err := e.evaluateExpression(value.X)
		if err != nil {
			return nil, err
		}
		if value.Op == token.LAND {
			truth, ok := left.(bool)
			if !ok {
				return nil, fmt.Errorf("left operand of && is %T", left)
			}
			if !truth {
				return false, nil
			}
		}
		if value.Op == token.LOR {
			truth, ok := left.(bool)
			if !ok {
				return nil, fmt.Errorf("left operand of || is %T", left)
			}
			if truth {
				return true, nil
			}
		}
		right, err := e.evaluateExpression(value.Y)
		if err != nil {
			return nil, err
		}
		result, err := evaluateIntegerBinary(value.Op, left, right)
		if err != nil {
			return nil, err
		}
		return coerceBodyValue(result, typed.Type)
	default:
		return nil, fmt.Errorf("unsupported generated expression %T", expression)
	}
}

func coerceBodyValue(value any, valueType types.Type) (any, error) {
	if valueType == nil {
		return nil, fmt.Errorf("generated value has no static type")
	}
	basic, ok := valueType.Underlying().(*types.Basic)
	if !ok {
		return nil, fmt.Errorf("unsupported generated value type %s", valueType)
	}
	switch basic.Kind() {
	case types.Int, types.Int32, types.Int64, types.UntypedInt, types.UntypedRune:
		number, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("generated integer value has type %T", value)
		}
		if basic.Kind() == types.Int32 {
			number = int64(int32(number))
		} else if basic.Kind() == types.Int {
			number = int64(int(number))
		}
		return number, nil
	case types.Bool, types.UntypedBool:
		if truth, ok := value.(bool); ok {
			return truth, nil
		}
	case types.String, types.UntypedString:
		if text, ok := value.(string); ok {
			return text, nil
		}
	}
	return nil, fmt.Errorf("unsupported generated value %T for type %s", value, valueType)
}

func evaluateIntegerBinary(operator token.Token, left, right any) (any, error) {
	if a, ok := left.(int64); ok {
		b, ok := right.(int64)
		if !ok {
			return nil, fmt.Errorf("binary operands have types %T and %T", left, right)
		}
		switch operator {
		case token.ADD:
			return a + b, nil
		case token.SUB:
			return a - b, nil
		case token.MUL:
			return a * b, nil
		case token.EQL:
			return a == b, nil
		case token.NEQ:
			return a != b, nil
		case token.LSS:
			return a < b, nil
		case token.LEQ:
			return a <= b, nil
		case token.GTR:
			return a > b, nil
		case token.GEQ:
			return a >= b, nil
		}
	}
	if a, ok := left.(bool); ok {
		b, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("binary operands have types %T and %T", left, right)
		}
		switch operator {
		case token.LAND:
			return a && b, nil
		case token.LOR:
			return a || b, nil
		case token.EQL:
			return a == b, nil
		case token.NEQ:
			return a != b, nil
		}
	}
	if a, ok := left.(string); ok {
		b, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("binary operands have types %T and %T", left, right)
		}
		switch operator {
		case token.ADD:
			return a + b, nil
		case token.EQL:
			return a == b, nil
		case token.NEQ:
			return a != b, nil
		case token.LSS:
			return a < b, nil
		case token.LEQ:
			return a <= b, nil
		case token.GTR:
			return a > b, nil
		case token.GEQ:
			return a >= b, nil
		}
	}
	return nil, fmt.Errorf("unsupported generated binary operator %s for %T", operator, left)
}
