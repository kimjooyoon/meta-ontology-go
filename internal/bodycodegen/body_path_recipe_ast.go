package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
)

func (b *recipeArena) expression(node ast.Expr, depth int) (int, error) {
	if err := b.bounded(depth); err != nil {
		return 0, err
	}
	value := bodyplan.Expr{}
	switch e := node.(type) {
	case *ast.ParenExpr:
		return b.expression(e.X, depth)
	case *ast.Ident:
		value.Kind, value.Name = bodyplan.ExprLocal, e.Name
		if e.Name == "input" {
			if b.inputSeen {
				return b.inputIndex, nil
			}
			value.Kind = bodyplan.ExprInput
		} else if e.Name == "true" || e.Name == "false" {
			value = bodyplan.Expr{Kind: bodyplan.ExprBool, Bool: e.Name == "true"}
		}
	case *ast.BasicLit:
		integer, err := strconv.ParseInt(e.Value, 0, 64)
		if e.Kind != token.INT || err != nil {
			return 0, fmt.Errorf("recipe literal must fit int64")
		}
		value = bodyplan.Expr{Kind: bodyplan.ExprInt, Int: integer}
	case *ast.UnaryExpr:
		literal, ok := e.X.(*ast.BasicLit)
		if !ok || e.Op != token.SUB || literal.Kind != token.INT {
			return 0, fmt.Errorf("recipe unary expression requires a negative integer literal")
		}
		integer, err := strconv.ParseInt("-"+literal.Value, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("recipe negative literal must fit int64")
		}
		value = bodyplan.Expr{Kind: bodyplan.ExprInt, Int: integer}
	case *ast.BinaryExpr:
		operation := recipeOperation(e.Op)
		if operation == "" {
			return 0, fmt.Errorf("recipe binary operation %s is outside the typed arena", e.Op)
		}
		left, err := b.expression(e.X, depth+1)
		if err != nil {
			return 0, err
		}
		right, err := b.expression(e.Y, depth+1)
		if err != nil {
			return 0, err
		}
		value = bodyplan.Expr{Kind: bodyplan.ExprBinary, Operation: operation, Left: left, Right: right}
	default:
		return 0, fmt.Errorf("recipe expression %T is outside the typed arena", node)
	}
	if b.expressionCount == len(b.expressions) {
		return 0, fmt.Errorf("recipe exceeds 128 expressions")
	}
	index := b.expressionCount
	b.expressions[index], b.expressionPositions[index] = value, node.Pos()
	b.expressionCount++
	if value.Kind == bodyplan.ExprInput {
		b.inputIndex, b.inputSeen = index, true
	}
	return index, nil
}

func recipeOperation(op token.Token) string {
	switch op {
	case token.ADD:
		return "add"
	case token.SUB:
		return "subtract"
	case token.MUL:
		return "multiply"
	case token.LSS:
		return "less_than"
	case token.LEQ:
		return "less_equal"
	case token.EQL:
		return "equal"
	case token.LAND:
		return "and"
	case token.LOR:
		return "or"
	}
	return ""
}

func (b *recipeArena) statement(node ast.Stmt, depth int) (int, error) {
	if err := b.bounded(depth); err != nil {
		return 0, err
	}
	var value bodyplan.Stmt
	var expression ast.Expr
	switch s := node.(type) {
	case *ast.DeclStmt:
		decl, ok := s.Decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.VAR || len(decl.Specs) != 1 {
			return 0, fmt.Errorf("recipe requires a single local declaration")
		}
		spec, ok := decl.Specs[0].(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return 0, fmt.Errorf("recipe local requires one initializer")
		}
		if spec.Type != nil {
			name, ok := spec.Type.(*ast.Ident)
			if !ok || name.Name != "int64" && name.Name != "bool" {
				return 0, fmt.Errorf("recipe explicit local type must be int64 or bool")
			}
		}
		value.Kind, value.Name, expression = bodyplan.StmtLet, spec.Names[0].Name, spec.Values[0]
	case *ast.AssignStmt:
		if s.Tok != token.ASSIGN || len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return 0, fmt.Errorf("recipe requires a single existing-local assignment")
		}
		name, ok := s.Lhs[0].(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("recipe assignment requires a local name")
		}
		value.Kind, value.Name, expression = bodyplan.StmtAssign, name.Name, s.Rhs[0]
	case *ast.ReturnStmt:
		if len(s.Results) != 1 {
			return 0, fmt.Errorf("recipe requires one return value")
		}
		value.Kind, expression = bodyplan.StmtReturn, s.Results[0]
	case *ast.IfStmt:
		if s.Init != nil {
			return 0, fmt.Errorf("recipe if initializer is unsupported")
		}
		value.Kind, expression = bodyplan.StmtIf, s.Cond
		var err error
		value.Then, err = b.sequence(s.Body.List, depth+1)
		if err != nil {
			return 0, err
		}
		if s.Else != nil {
			otherwise, ok := s.Else.(*ast.BlockStmt)
			if !ok {
				return 0, fmt.Errorf("recipe else requires an explicit block")
			}
			value.Else, err = b.sequence(otherwise.List, depth+1)
			if err != nil {
				return 0, err
			}
		}
	default:
		return 0, fmt.Errorf("recipe statement %T is outside the typed arena", node)
	}
	index, err := b.expression(expression, depth+1)
	if err != nil {
		return 0, err
	}
	value.Expr = index
	if b.statementCount == len(b.statements) {
		return 0, fmt.Errorf("recipe exceeds 128 statements")
	}
	index = b.statementCount
	b.statements[index], b.statementPositions[index] = value, node.Pos()
	b.statementCount++
	return index, nil
}
