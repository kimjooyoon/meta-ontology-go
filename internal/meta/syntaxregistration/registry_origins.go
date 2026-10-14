package syntaxregistration

import (
	"fmt"
	"go/ast"
)

// The Cases field owns the sequence. Follow only its finite source expression;
// unrelated literals and helpers do not supply registration destinations.
func (source *goSource) registryCaseTail(function *ast.FuncDecl) (*ast.CompositeLit, error) {
	if function.Body == nil || len(function.Body.List) != 1 {
		return nil, fmt.Errorf("native registry must have one direct return")
	}
	returned, ok := function.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(returned.Results) != 1 {
		return nil, fmt.Errorf("native registry return is unresolved")
	}
	registry, ok := returned.Results[0].(*ast.CompositeLit)
	if !ok || !namedRegistryType(registry.Type, "Registry") {
		return nil, fmt.Errorf("native registry is not a Registry literal")
	}
	var value ast.Expr
	for _, element := range registry.Elts {
		field, ok := element.(*ast.KeyValueExpr)
		if !ok || !namedRegistryType(field.Key, "Cases") {
			continue
		}
		if value != nil {
			return nil, fmt.Errorf("native Cases field is ambiguous")
		}
		value = field.Value
	}
	if value == nil {
		return nil, fmt.Errorf("native Cases field is missing")
	}
	walk := registryOriginWalk{source: source, functions: map[string]bool{}, literals: map[*ast.CompositeLit]bool{}}
	leaves, err := walk.expression(value)
	if err != nil {
		return nil, err
	}
	if len(leaves) == 0 || len(leaves[len(leaves)-1].Elts) == 0 {
		return nil, fmt.Errorf("native case tail is empty or unresolved")
	}
	return leaves[len(leaves)-1], nil
}

func namedRegistryType(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name
}

func caseRegistrySlice(expression ast.Expr) bool {
	array, ok := expression.(*ast.ArrayType)
	return ok && array.Len == nil && namedRegistryType(array.Elt, "CaseDefinition")
}

type registryOriginWalk struct {
	source    *goSource
	functions map[string]bool
	literals  map[*ast.CompositeLit]bool
}

func (walk *registryOriginWalk) expression(expression ast.Expr) ([]*ast.CompositeLit, error) {
	switch expression := expression.(type) {
	case *ast.CompositeLit:
		if !caseRegistrySlice(expression.Type) || walk.literals[expression] {
			return nil, fmt.Errorf("native case literal is unsupported or repeated")
		}
		walk.literals[expression] = true
		return []*ast.CompositeLit{expression}, nil
	case *ast.CallExpr:
		name, ok := expression.Fun.(*ast.Ident)
		if !ok || len(expression.Args) != 0 || expression.Ellipsis.IsValid() {
			return nil, fmt.Errorf("native case helper must be a local zero-argument call")
		}
		return walk.helper(name.Name)
	default:
		return nil, fmt.Errorf("native case expression has no finite source origin")
	}
}
