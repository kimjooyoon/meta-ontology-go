package bodycodegen

import "go/ast"

type valueAccessOperand struct {
	kind string
	expr ast.Expr
}

// access retains the source occurrence of each selector and explicit bound.
// Operand nodes distinguish the collection from the expressions choosing a
// position. Omitted bounds have no invented source range or runtime value.
// This view does not resolve individual elements of constructed collections.
func (r *valueOriginReader) access(kind, name string, at, base ast.Expr, env valueOriginEnv, operands ...valueAccessOperand) []string {
	inputs := r.inputs(base, env)
	if kind != "field_access" {
		inputs = joinValueOriginIDs([]string{r.node("access_base", "", base, inputs, nil)})
	}
	for _, operand := range operands {
		if operand.expr == nil {
			continue
		}
		dependencies := r.inputs(operand.expr, env)
		id := r.node(operand.kind, "", operand.expr, dependencies, nil)
		inputs = joinValueOriginIDs(inputs, []string{id})
	}
	return joinValueOriginIDs([]string{r.node(kind, name, at, inputs, nil)})
}
