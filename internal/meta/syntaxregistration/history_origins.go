package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Freeze the prior source-selected target before the next corpus grows. The
// historical observation is exact; it does not inherit today's denominator.
func pinPreviousSyntaxTarget(source *goSource, version, capability int) error {
	function, err := source.function("expectedBoundarySpecsForVersion")
	if err != nil {
		return err
	}
	var selection *ast.SwitchStmt
	count := 0
	ast.Inspect(function.Body, func(node ast.Node) bool {
		candidate, ok := node.(*ast.SwitchStmt)
		if ok && candidate.Init == nil && candidate.Tag != nil && source.text(candidate.Tag) == "version" {
			selection = candidate
			count++
			return false
		}
		return true
	})
	if count != 1 {
		return fmt.Errorf("historical syntax target switch is missing or ambiguous")
	}
	seen := 0
	for _, statement := range selection.Body.List {
		clause, ok := statement.(*ast.CaseClause)
		if !ok {
			return fmt.Errorf("historical syntax target has an unsupported clause")
		}
		for _, expression := range clause.List {
			value, literal := integer(expression)
			if !literal {
				return fmt.Errorf("historical syntax version is not a literal")
			}
			if value != version {
				continue
			}
			seen++
			if len(clause.Body) != 1 {
				return fmt.Errorf("historical syntax target is not one source assignment")
			}
			assignment, ok := clause.Body[0].(*ast.AssignStmt)
			if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 ||
				source.text(assignment.Lhs[0]) != "expected[0].Target" {
				return fmt.Errorf("historical syntax target ownership differs")
			}
			target, literal := integer(assignment.Rhs[0])
			if !literal || target != capability {
				return fmt.Errorf("historical syntax target differs from its pinned source corpus")
			}
		}
	}
	if seen > 1 {
		return fmt.Errorf("historical syntax version is repeated")
	}
	if seen == 0 {
		source.insert(selection.Body.Rbrace, fmt.Sprintf("\ncase %d: expected[0].Target = %d\n", version, capability))
	}
	return nil
}
