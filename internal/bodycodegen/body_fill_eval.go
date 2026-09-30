package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

// evaluateIntegerCases interprets only the already typechecked, pure integer
// body profile. It is used to score bounded experimental candidates without
// executing arbitrary generated programs.
func evaluateIntegerCases(
	source []byte,
	activity string,
	cases []IRBodyFillTestCase,
) ([]IRBodyFillCaseResult, int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", source, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, 0, fmt.Errorf("parse generated Go: %w", err)
	}
	function, ok := findFunction(file, activity)
	if !ok {
		return nil, 0, fmt.Errorf("generated function %q was not found", activity)
	}
	results := make([]IRBodyFillCaseResult, 0, len(cases))
	passed := 0
	for _, testCase := range cases {
		environment := map[string]any{"input": testCase.Input}
		value, returned, err := evaluateIntBlock(function.Body, environment)
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

func evaluateIntBlock(block *ast.BlockStmt, environment map[string]any) (any, bool, error) {
	for _, statement := range block.List {
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
			var initial any
			if len(spec.Values) == 1 {
				var err error
				initial, err = evaluateIntExpression(spec.Values[0], environment)
				if err != nil {
					return nil, false, err
				}
			}
			environment[spec.Names[0].Name] = initial
		case *ast.AssignStmt:
			if len(value.Lhs) != 1 || len(value.Rhs) != 1 || value.Tok != token.ASSIGN {
				return nil, false, fmt.Errorf("unsupported generated assignment")
			}
			name, ok := value.Lhs[0].(*ast.Ident)
			if !ok {
				return nil, false, fmt.Errorf("unsupported generated assignment target")
			}
			assigned, err := evaluateIntExpression(value.Rhs[0], environment)
			if err != nil {
				return nil, false, err
			}
			environment[name.Name] = assigned
		case *ast.IfStmt:
			condition, err := evaluateIntExpression(value.Cond, environment)
			if err != nil {
				return nil, false, err
			}
			truth, ok := condition.(bool)
			if !ok {
				return nil, false, fmt.Errorf("generated if condition evaluated to %T", condition)
			}
			if truth {
				result, returned, err := evaluateIntBlock(value.Body, environment)
				if err != nil || returned {
					return result, returned, err
				}
			} else {
				switch otherwise := value.Else.(type) {
				case *ast.BlockStmt:
					result, returned, err := evaluateIntBlock(otherwise, environment)
					if err != nil || returned {
						return result, returned, err
					}
				case *ast.IfStmt:
					result, returned, err := evaluateIntBlock(&ast.BlockStmt{List: []ast.Stmt{otherwise}}, environment)
					if err != nil || returned {
						return result, returned, err
					}
				}
			}
		case *ast.ReturnStmt:
			if len(value.Results) != 1 {
				return nil, false, fmt.Errorf("generated return must have one value")
			}
			returned, err := evaluateIntExpression(value.Results[0], environment)
			return returned, true, err
		default:
			return nil, false, fmt.Errorf("unsupported generated statement %T", statement)
		}
	}
	return nil, false, nil
}

func evaluateIntExpression(expression ast.Expr, environment map[string]any) (any, error) {
	switch value := expression.(type) {
	case *ast.Ident:
		if value.Name == "true" {
			return true, nil
		}
		if value.Name == "false" {
			return false, nil
		}
		result, ok := environment[value.Name]
		if !ok {
			return nil, fmt.Errorf("unbound identifier %q", value.Name)
		}
		return result, nil
	case *ast.BasicLit:
		if value.Kind != token.INT {
			return nil, fmt.Errorf("unsupported generated literal %s", value.Kind)
		}
		parsed, err := strconv.ParseInt(value.Value, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("parse generated integer %q: %w", value.Value, err)
		}
		return parsed, nil
	case *ast.ParenExpr:
		return evaluateIntExpression(value.X, environment)
	case *ast.UnaryExpr:
		operand, err := evaluateIntExpression(value.X, environment)
		if err != nil {
			return nil, err
		}
		switch value.Op {
		case token.SUB:
			number, ok := operand.(int64)
			if !ok {
				return nil, fmt.Errorf("unary minus operand is %T", operand)
			}
			return -number, nil
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
		left, err := evaluateIntExpression(value.X, environment)
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
		right, err := evaluateIntExpression(value.Y, environment)
		if err != nil {
			return nil, err
		}
		return evaluateIntegerBinary(value.Op, left, right)
	default:
		return nil, fmt.Errorf("unsupported generated expression %T", expression)
	}
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
	return nil, fmt.Errorf("unsupported generated binary operator %s for %T", operator, left)
}
