package analyzer

import (
	"go/ast"
)

type factCollector struct {
	resolver    *resolver
	delta       *SemanticDelta
	subject     Identity
	file        parsedFile
	typeNodes   map[ast.Node]bool
	callTargets map[ast.Expr]bool
	varTypes    map[string]typeReference
	blocked     map[string]bool
	parents     []ast.Node
}

func newFactCollector(resolver *resolver, delta *SemanticDelta, subject Identity, file parsedFile, function *ast.FuncDecl) *factCollector {
	return &factCollector{
		resolver:    resolver,
		delta:       delta,
		subject:     subject,
		file:        file,
		typeNodes:   make(map[ast.Node]bool),
		callTargets: make(map[ast.Expr]bool),
		varTypes:    variableTypes(function, file, resolver.imports[file.file]),
		blocked:     localBindings(function),
	}
}
func (c *factCollector) collectSignature(function *ast.FuncDecl) {
	if function.Recv != nil {
		for _, field := range function.Recv.List {
			c.collectTypeRefs(field.Type, RelationUses, false, OriginSignature)
		}
	}
	if function.Type.Params != nil {
		for _, field := range function.Type.Params.List {
			c.collectTypeRefs(field.Type, RelationUses, false, OriginSignature)
		}
	}
	if function.Type.Results != nil {
		for _, field := range function.Type.Results.List {
			c.collectTypeRefs(field.Type, RelationGenerates, false, OriginSignature)
		}
	}
}
func (c *factCollector) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		if len(c.parents) > 0 {
			c.parents = c.parents[:len(c.parents)-1]
		}
		return nil
	}
	var parent ast.Node
	if len(c.parents) > 0 {
		parent = c.parents[len(c.parents)-1]
	}
	c.handle(node, parent)
	c.parents = append(c.parents, node)
	return c
}
