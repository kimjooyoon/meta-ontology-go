package analyzer

import (
	"go/ast"
	"go/token"
	"sort"
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
			c.collectTypeRefs(field.Type, RelationUses, false)
		}
	}
	if function.Type.Params != nil {
		for _, field := range function.Type.Params.List {
			c.collectTypeRefs(field.Type, RelationUses, false)
		}
	}
	if function.Type.Results != nil {
		for _, field := range function.Type.Results.List {
			c.collectTypeRefs(field.Type, RelationGenerates, false)
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

func (c *factCollector) handle(node ast.Node, parent ast.Node) {
	switch current := node.(type) {
	case *ast.CallExpr:
		c.callTargets[current.Fun] = true
		result := c.resolve(current.Fun)
		c.recordResolution(result, relationForCall(result.entries), current.Fun, "call target")
	case *ast.Field:
		c.collectTypeRefs(current.Type, RelationUses, true)
	case *ast.CompositeLit:
		c.collectTypeRefs(current.Type, RelationUses, true)
	case *ast.TypeAssertExpr:
		c.collectTypeRefs(current.Type, RelationUses, true)
	case *ast.SelectorExpr:
		if c.typeNodes[current] || c.callTargets[current] {
			return
		}
		result := c.resolve(current)
		c.recordResolution(result, relationForReference(result.entries), current, "symbol reference")
	case *ast.Ident:
		if c.typeNodes[current] || isSelectorChild(parent) || isCallTarget(parent, current) || isDeclarationName(parent, current) {
			return
		}
		result := c.resolve(current)
		c.recordResolution(result, relationForReference(result.entries), current, "symbol reference")
	}
}

func (c *factCollector) collectTypeRefs(expr ast.Expr, relation Relation, respectBlocks bool) {
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
		if result.state != unresolved {
			c.recordResolution(result, relation, expression, "type reference")
		}
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

func (c *factCollector) recordResolution(result resolution, relation Relation, expression ast.Expr, reason string) {
	if result.state == unresolved {
		if reason == "call target" {
			c.delta.ImplementationDetails = append(c.delta.ImplementationDetails, ImplementationDetail{
				Reference: expressionName(expression),
				Span:      c.span(expression),
				Reason:    "unregistered semantic symbol",
			})
		}
		return
	}
	if result.state == ambiguous {
		options := uniqueIdentities(result.entries)
		if len(options) == 0 {
			return
		}
		c.delta.Candidates = append(c.delta.Candidates, Candidate{
			Subject:   c.subject,
			Relation:  relation,
			Reference: expressionName(expression),
			Options:   options,
			Span:      c.span(expression),
			Reason:    "multiple registered semantic symbols match",
		})
		return
	}
	if len(result.entries) != 1 || !result.entries[0].Identity.Valid() {
		return
	}
	c.delta.Added = append(c.delta.Added, Fact{
		Subject:  c.subject,
		Relation: relation,
		Object:   result.entries[0].Identity,
		Span:     c.span(expression),
	})
}

func (c *factCollector) span(node ast.Node) Span {
	return spanFor(c.resolver.fileSet, node)
}

func relationForCall(entries []Registration) Relation {
	if allOfKind(entries, KindActivity) {
		return RelationInvokes
	}
	if allOfKind(entries, KindEntity) {
		return RelationUses
	}
	return RelationReferences
}

func relationForReference(entries []Registration) Relation {
	if allOfKind(entries, KindEntity) {
		return RelationUses
	}
	return RelationReferences
}

func allOfKind(entries []Registration, kind SymbolKind) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if entry.Kind != kind {
			return false
		}
	}
	return true
}

func uniqueIdentities(entries []Registration) []Identity {
	seen := make(map[Identity]bool, len(entries))
	options := make([]Identity, 0, len(entries))
	for _, entry := range entries {
		if entry.Identity.Valid() && !seen[entry.Identity] {
			seen[entry.Identity] = true
			options = append(options, entry.Identity)
		}
	}
	sort.Slice(options, func(i, j int) bool { return identityLess(options[i], options[j]) })
	return options
}

func localBindings(function *ast.FuncDecl) map[string]bool {
	blocked := make(map[string]bool)
	collectFieldNames(blocked, function.Recv)
	collectFieldNames(blocked, function.Type.Params)
	collectFieldNames(blocked, function.Type.Results)
	if function.Body == nil {
		return blocked
	}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.ValueSpec:
			addNames(blocked, current.Names)
		case *ast.AssignStmt:
			if current.Tok == token.DEFINE {
				addExprNames(blocked, current.Lhs)
			}
		case *ast.RangeStmt:
			addExprName(blocked, current.Key)
			addExprName(blocked, current.Value)
		case *ast.TypeSpec:
			blocked[current.Name.Name] = true
		case *ast.FuncLit:
			collectFieldNames(blocked, current.Type.Params)
			collectFieldNames(blocked, current.Type.Results)
		}
		return true
	})
	return blocked
}

func collectFieldNames(blocked map[string]bool, fields *ast.FieldList) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		addNames(blocked, field.Names)
	}
}

func addNames(blocked map[string]bool, names []*ast.Ident) {
	for _, name := range names {
		if name.Name != "_" {
			blocked[name.Name] = true
		}
	}
}

func addExprNames(blocked map[string]bool, expressions []ast.Expr) {
	for _, expression := range expressions {
		addExprName(blocked, expression)
	}
}

func addExprName(blocked map[string]bool, expression ast.Expr) {
	ident, ok := expression.(*ast.Ident)
	if ok && ident.Name != "_" {
		blocked[ident.Name] = true
	}
}
