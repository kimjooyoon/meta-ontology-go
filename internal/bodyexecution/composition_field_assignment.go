package bodyexecution

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func singleAssignmentTarget(assign *ast.AssignStmt) ast.Expr {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return nil
	}
	return assign.Lhs[0]
}

func (p *compositionFieldProjection) collectFieldAssignment(assign *ast.AssignStmt, skipped map[*ast.Ident]bool, objects map[*ast.Object]int, fset *token.FileSet, source string) {
	selector, ok := singleAssignmentTarget(assign).(*ast.SelectorExpr)
	if !ok {
		return
	}
	name, ok := ast.Unparen(selector.X).(*ast.Ident)
	if !ok {
		return
	}
	_, compound := bodycodegen.CompoundAssignmentOperator(assign.Tok)
	skipped[name] = !compound
	binding, found := objects[name.Obj]
	if !found {
		return
	}
	for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
		if descriptor.GoName != selector.Sel.Name {
			continue
		}
		start, end := fset.Position(assign.Rhs[0].Pos()).Offset, fset.Position(assign.Rhs[0].End()).Offset
		site := CompositionFieldSite{Binding: binding, Field: field, Kind: "field_assignment",
			DependencyMode: "scalar_rhs", Start: start, End: end, Expression: source[start:end]}
		kind, value := "write", ast.Node(assign.Rhs[0])
		if compound {
			site.DependencyMode, site.Operation = "compound_update", assign.Tok.String()
			kind, value = "compound_write", assign
		}
		first := len(p.sites)
		p.sites = append(p.sites, site)
		p.addNode(assign, value, kind, binding, first, field, fset)
		return
	}
}

func (r compositionFieldRenderer) compoundWrite(n compositionFieldNode) string {
	assign := n.node.(*ast.AssignStmt)
	operation, _ := bodycodegen.CompoundAssignmentOperator(assign.Tok)
	start, end := r.sourceOffsets(assign.Lhs[0])
	target := r.source[start:end]
	previous := r.render(start, end, nil)
	start, end = r.sourceOffsets(assign.Rhs[0])
	operand := r.render(start, end, nil)
	expression := previous + " " + operation.String() + " (" + operand + ")"
	if operation == token.QUO {
		expression = r.plan.faultPrefix + "Quotient(" + previous + "," + operand + "," + strconv.Itoa(n.fault) + ")"
	}
	value := r.scalarValue(n, expression)
	return target + " = " + r.call("FieldAssign", strconv.Itoa(n.first), r.prefix+"FieldScope",
		"&"+r.slot(n.binding)+"["+strconv.Itoa(n.field)+"]", value)
}
