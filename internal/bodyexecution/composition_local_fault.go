package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func (p *arithmeticProjection) collectLocalDivision(assign *ast.AssignStmt, ids [2]string, info *types.Info) {
	if assign.Tok != token.QUO_ASSIGN || len(assign.Rhs) != 1 {
		return
	}
	target, ok := singleAssignmentTarget(assign).(*ast.Ident)
	if !ok || !types.Identical(info.TypeOf(target), types.Typ[types.Int64]) {
		return
	}
	start, end := p.fset.Position(assign.Pos()).Offset, p.fset.Position(assign.End()).Offset
	site := CompositionFaultSite{RootActivityID: ids[0], ActivityID: ids[1], Operator: "/", Assignment: "/=", Variable: target.Name,
		Expression: p.source[start:end], ProjectionSHA256: digest([]byte(p.source)), Start: start, End: end}
	site.ExpressionID = compositionDigest(site)
	p.sites = append(p.sites, arithmeticProjectionSite{site, assign})
}

// Bind retained source operations to local writes before rendering observation.
// The AST walk preserves assignment order, including repeated writes to a name.
func (p *readProjection) bindLocalFaults(faults []CompositionFaultSite) error {
	byVariable := map[[3]string][]int{}
	for index, site := range faults {
		if site.Assignment == "/=" && site.Variable != "" {
			key := [3]string{site.RootActivityID, site.ActivityID, site.Variable}
			byVariable[key] = append(byVariable[key], index)
		}
	}
	for index := range p.nodes {
		node := &p.nodes[index]
		if node.localKind != "compound_write" {
			continue
		}
		local := p.localSites[node.localWrite]
		if local.Operation != "/=" {
			continue
		}
		key := [3]string{local.RootActivityID, local.ActivityID, local.Name}
		queue := byVariable[key]
		if len(queue) == 0 {
			return fmt.Errorf("compound local division has no source fault site")
		}
		node.fault, byVariable[key] = queue[0], queue[1:]
	}
	for _, remaining := range byVariable {
		if len(remaining) != 0 {
			return fmt.Errorf("source local division has no observation binding")
		}
	}
	return nil
}
