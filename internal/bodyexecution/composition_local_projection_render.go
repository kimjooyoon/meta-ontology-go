package bodyexecution

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func (p readProjection) localSlot(variable int) string {
	return p.prefix + "LocalSlot" + strconv.Itoa(variable)
}

func (p readProjection) localAssignment(index int, value string) string {
	site := p.localSites[index]
	return p.prefix + "LocalAssign(" + strconv.Itoa(index) + "," + p.prefix + "LocalScope,&" +
		p.localSlot(site.VariableOrdinal) + ",func() " + site.ScalarType + " {return " + value + "})"
}

func (p readProjection) renderLocal(item readProjectionNode) string {
	switch item.localKind {
	case "scope":
		body := item.node.(*ast.BlockStmt)
		var out strings.Builder
		out.WriteString("{" + p.prefix + "LocalScope:=" + p.prefix + "LocalFrame();")
		for _, variable := range item.variables {
			out.WriteString("var " + p.localSlot(variable) + " int = -1;")
		}
		out.WriteString(p.render(p.offset(body.Lbrace)+1, p.offset(body.Rbrace)))
		out.WriteString("}")
		return out.String()
	case "write":
		if update, ok := item.node.(*ast.IncDecStmt); ok {
			value := p.render(p.offset(update.X.Pos()), p.offset(update.X.End()))
			return p.localSites[item.localWrite].Name + " = " + p.localAssignment(item.localWrite, value+" + 1")
		}
		return p.localAssignment(item.localWrite, p.renderExcept(item.start, item.end, item.node))
	case "use":
		site := p.localUseSites[item.site]
		return p.prefix + "LocalUse(" + strconv.Itoa(item.site) + "," + p.prefix + "LocalScope," +
			p.localSlot(site.VariableOrdinal) + "," + site.Name + ")"
	case "compound_write":
		assignment := item.node.(*ast.AssignStmt)
		operation, _ := bodycodegen.CompoundAssignmentOperator(assignment.Tok)
		previous := p.render(p.offset(assignment.Lhs[0].Pos()), p.offset(assignment.Lhs[0].End()))
		operand := p.render(p.offset(assignment.Rhs[0].Pos()), p.offset(assignment.Rhs[0].End()))
		expression := previous + " " + operation.String() + " (" + operand + ")"
		if operation == token.QUO {
			expression = p.faultPrefix + "Quotient(" + previous + "," + operand + "," + strconv.Itoa(item.fault) + ")"
		}
		return p.localSites[item.localWrite].Name + " = " +
			p.localAssignment(item.localWrite, expression)
	}
	return p.source[item.start:item.end]
}
