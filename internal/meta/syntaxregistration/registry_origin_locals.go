package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Local row declarations retain their physical literal origins. Only a finite
// sequence of single assignments and one explicit slice return is admitted.
func (walk *registryOriginWalk) localElements(body []ast.Stmt) ([]*ast.CompositeLit, error) {
	locals := map[string]*ast.CompositeLit{}
	for _, statement := range body[:len(body)-1] {
		assignment, ok := statement.(*ast.AssignStmt)
		if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			return nil, fmt.Errorf("native registry rows require immutable local literals")
		}
		name, named := assignment.Lhs[0].(*ast.Ident)
		literal, explicit := assignment.Rhs[0].(*ast.CompositeLit)
		if !named || name.Name == "_" || name.Name == walk.elementType || locals[name.Name] != nil || !explicit || !namedRegistryType(literal.Type, walk.elementType) {
			return nil, fmt.Errorf("native registry local row is ambiguous or unsupported")
		}
		locals[name.Name] = literal
	}
	for _, literal := range locals {
		for _, element := range literal.Elts {
			if field, ok := element.(*ast.KeyValueExpr); ok {
				if value, ok := field.Value.(*ast.Ident); ok && locals[value.Name] != nil {
					return nil, fmt.Errorf("native registry local row shadows a field value")
				}
			}
		}
	}
	returned, ok := body[len(body)-1].(*ast.ReturnStmt)
	if !ok || len(returned.Results) != 1 {
		return nil, fmt.Errorf("native registry rows require one complete return")
	}
	list, ok := returned.Results[0].(*ast.CompositeLit)
	if !ok || !walk.sliceType(list.Type) || len(list.Elts) < len(locals) {
		return nil, fmt.Errorf("native registry row return differs from its declarations")
	}
	seen := map[string]bool{}
	for _, element := range list.Elts {
		if literal, ok := element.(*ast.CompositeLit); ok {
			if literal.Type != nil && !namedRegistryType(literal.Type, walk.elementType) {
				return nil, fmt.Errorf("native registry inline row has an unknown type")
			}
			continue
		}
		name, ok := element.(*ast.Ident)
		if !ok || seen[name.Name] || locals[name.Name] == nil {
			return nil, fmt.Errorf("native registry returned row is missing or repeated")
		}
		seen[name.Name] = true
		walk.locals[element] = locals[name.Name]
	}
	if len(seen) != len(locals) {
		return nil, fmt.Errorf("native registry contains an unreturned local row")
	}
	return walk.expression(list)
}
