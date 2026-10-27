package syntaxregistration

import (
	"fmt"
	"go/ast"
)

// The Cases field owns the sequence. Follow only its finite source expression;
// unrelated literals and helpers do not supply registration destinations.
func (source *goSource) registryCaseTail(function *ast.FuncDecl) (*ast.CompositeLit, error) {
	leaves, err := source.registryCaseLeaves(function)
	if err != nil {
		return nil, err
	}
	if len(leaves) == 0 || len(leaves[len(leaves)-1].Elts) == 0 {
		return nil, fmt.Errorf("native case tail is empty or unresolved")
	}
	return leaves[len(leaves)-1], nil
}

func (source *goSource) registryCaseLeaves(function *ast.FuncDecl) ([]*ast.CompositeLit, error) {
	return source.registryFieldLeaves(function, "Cases", "CaseDefinition", false)
}

func (source *goSource) registryFieldLeaves(function *ast.FuncDecl, fieldName, elementType string,
	optional bool) ([]*ast.CompositeLit, error) {
	origins, err := source.registryFieldOrigins(function, fieldName, elementType, optional)
	return origins.leaves, err
}

type registryOrigins struct {
	leaves []*ast.CompositeLit
	locals map[ast.Expr]*ast.CompositeLit
}

func (origins registryOrigins) literal(element ast.Expr, elementType string) (*ast.CompositeLit, error) {
	literal, ok := element.(*ast.CompositeLit)
	if !ok {
		literal = origins.locals[element]
	}
	if literal == nil || (literal.Type != nil && !namedRegistryType(literal.Type, elementType)) {
		return nil, fmt.Errorf("native %s row has no exact literal origin", elementType)
	}
	return literal, nil
}

func (source *goSource) registryFieldOrigins(function *ast.FuncDecl, fieldName, elementType string,
	optional bool) (registryOrigins, error) {
	registry, err := registryReturnLiteral(function)
	if err != nil {
		return registryOrigins{}, err
	}
	value, err := registryLiteralField(registry, fieldName)
	if err != nil {
		return registryOrigins{}, err
	}
	if value == nil {
		if optional {
			return registryOrigins{}, nil
		}
		return registryOrigins{}, fmt.Errorf("native %s field is missing", fieldName)
	}
	walk := registryOriginWalk{source: source, elementType: elementType,
		functions: map[string]bool{}, literals: map[*ast.CompositeLit]bool{}, locals: map[ast.Expr]*ast.CompositeLit{}}
	leaves, err := walk.expression(value)
	return registryOrigins{leaves, walk.locals}, err
}

func registryReturnLiteral(function *ast.FuncDecl) (*ast.CompositeLit, error) {
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
	return registry, nil
}

func registryLiteralField(registry *ast.CompositeLit, fieldName string) (ast.Expr, error) {
	var value ast.Expr
	for _, element := range registry.Elts {
		field, ok := element.(*ast.KeyValueExpr)
		if !ok || !namedRegistryType(field.Key, fieldName) {
			continue
		}
		if value != nil {
			return nil, fmt.Errorf("native %s field is ambiguous", fieldName)
		}
		value = field.Value
	}
	return value, nil
}

func namedRegistryType(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name
}

func (walk *registryOriginWalk) sliceType(expression ast.Expr) bool {
	array, ok := expression.(*ast.ArrayType)
	return ok && array.Len == nil && namedRegistryType(array.Elt, walk.elementType)
}

type registryOriginWalk struct {
	source      *goSource
	elementType string
	functions   map[string]bool
	literals    map[*ast.CompositeLit]bool
	locals      map[ast.Expr]*ast.CompositeLit
}

func (walk *registryOriginWalk) expression(expression ast.Expr) ([]*ast.CompositeLit, error) {
	switch expression := expression.(type) {
	case *ast.CompositeLit:
		if !walk.sliceType(expression.Type) || walk.literals[expression] {
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
