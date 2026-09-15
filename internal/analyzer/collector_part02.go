package analyzer

import (
	"go/ast"
)

func (c *factCollector) handle(node ast.Node, parent ast.Node) {
	switch current := node.(type) {
	case *ast.CallExpr:
		c.callTargets[current.Fun] = true
		result := c.resolve(current.Fun)
		c.recordResolution(result, relationForCall(result.entries), current.Fun, "call target", OriginImplementation)
	case *ast.Field:
		c.collectTypeRefs(current.Type, RelationUses, true, OriginImplementation)
	case *ast.CompositeLit:
		c.collectTypeRefs(current.Type, RelationUses, true, OriginImplementation)
	case *ast.TypeAssertExpr:
		c.collectTypeRefs(current.Type, RelationUses, true, OriginImplementation)
	case *ast.SelectorExpr:
		if c.typeNodes[current] || c.callTargets[current] {
			return
		}
		result := c.resolve(current)
		c.recordResolution(result, relationForReference(result.entries), current, "symbol reference", OriginImplementation)
	case *ast.Ident:
		if c.typeNodes[current] || isSelectorChild(parent) || isCallTarget(parent, current) || isDeclarationName(parent, current) {
			return
		}
		result := c.resolve(current)
		c.recordResolution(result, relationForReference(result.entries), current, "symbol reference", OriginImplementation)
	}
}
func (c *factCollector) collectTypeRefs(
	expr ast.Expr, relation Relation, respectBlocks bool, origin ObservationOrigin,
) {
	if expr == nil {
		return
	}
	ast.Inspect(expr, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		c.typeNodes[node] = true
		expression, ok := node.(ast.Expr)
		if !ok {
			return true
		}
		result := c.resolveWithBlocks(expression, respectBlocks)
		c.recordResolution(result, relation, expression, "type reference", origin)
		return true
	})
}
func (c *factCollector) resolve(expr ast.Expr) resolution {
	return c.resolveWithBlocks(expr, true)
}
func (c *factCollector) resolveWithBlocks(expr ast.Expr, respectBlocks bool) resolution {
	if respectBlocks && c.blocksSemanticLookup(expr) {
		return resolution{state: unresolved}
	}
	return c.resolver.resolveExpression(expr, c.file, c.varTypes)
}
func (c *factCollector) blocksSemanticLookup(expr ast.Expr) bool {
	base := unwrapExpr(expr)
	if selector, ok := base.(*ast.SelectorExpr); ok {
		base = unwrapExpr(selector.X)
	}
	ident, ok := base.(*ast.Ident)
	if !ok || !c.blocked[ident.Name] {
		return false
	}
	_, hasType := c.varTypes[ident.Name]
	return !hasType
}
