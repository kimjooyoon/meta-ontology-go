package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// Keep the assignment intact until field observation owns its read/write window.
func (p *arithmeticProjection) collectFieldDivision(assign *ast.AssignStmt, ids [2]string, info *types.Info) {
	if assign.Tok != token.QUO_ASSIGN || len(assign.Rhs) != 1 {
		return
	}
	target, ok := singleAssignmentTarget(assign).(*ast.SelectorExpr)
	if !ok || !types.Identical(info.TypeOf(target), types.Typ[types.Int64]) {
		return
	}
	start, end := p.fset.Position(assign.Pos()).Offset, p.fset.Position(assign.End()).Offset
	site := CompositionFaultSite{RootActivityID: ids[0], ActivityID: ids[1], Operator: "/", Assignment: "/=",
		Expression: p.source[start:end], ProjectionSHA256: digest([]byte(p.source)), Start: start, End: end}
	site.ExpressionID = compositionDigest(site)
	p.sites = append(p.sites, arithmeticProjectionSite{site, assign})
}

// Each original field division retains its order within its source activity.
// Binary divisions in the operand have already received their own checked calls.
func (p *compositionFieldProjection) bindFieldFaults(faults []CompositionFaultSite) error {
	byActivity := map[[2]string][]int{}
	for index, site := range faults {
		if site.Assignment == "/=" && site.Variable == "" {
			key := [2]string{site.RootActivityID, site.ActivityID}
			byActivity[key] = append(byActivity[key], index)
		}
	}
	for index := range p.nodes {
		node := &p.nodes[index]
		if node.kind != "compound_write" || p.sites[node.first].Operation != "/=" {
			continue
		}
		binding := p.inventory.bindings[node.binding]
		key := [2]string{binding.RootActivityID, binding.ActivityID}
		queue := byActivity[key]
		if len(queue) == 0 {
			return fmt.Errorf("compound field division has no source fault site")
		}
		node.fault = queue[0]
		byActivity[key] = queue[1:]
	}
	for _, remaining := range byActivity {
		if len(remaining) != 0 {
			return fmt.Errorf("source field division has no observation binding")
		}
	}
	return nil
}
