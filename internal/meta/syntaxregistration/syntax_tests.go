package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func generateSyntaxTests(raw, corpus []byte) ([]byte, error) {
	total, valid, capability, err := corpusTotals(corpus)
	if err != nil {
		return nil, err
	}
	source, err := parseGo(raw)
	if err != nil {
		return nil, err
	}
	expected := map[string]int{
		"report.Summary.Satisfied": total, "report.Summary.ValidCases": valid,
		"report.Summary.Unresolved": total, "report.Summary.Total": total,
		"languagesyntax.FixedTotal": total, "languagesyntax.FixedCapabilityTotal": capability,
	}
	seen := map[string]int{}
	var failure error
	ast.Inspect(source.file, func(node ast.Node) bool {
		comparison, ok := node.(*ast.BinaryExpr)
		if !ok || (comparison.Op != token.NEQ && comparison.Op != token.EQL) {
			return true
		}
		name := strings.ReplaceAll(source.text(comparison.X), " ", "")
		actual, literal := integer(comparison.Y)
		if !literal {
			return true
		}
		if old, wanted := expected[name]; wanted && actual == old {
			source.replace(comparison.Y, strconv.Itoa(old+1))
			seen[name]++
		}
		if name == "len(report.Source.GoooFiles)" {
			if actual <= 0 {
				failure = fmt.Errorf("source inventory baseline is not positive")
			}
			source.replace(comparison.Y, strconv.Itoa(actual+1))
			seen[name]++
		}
		return true
	})
	if failure != nil {
		return nil, failure
	}
	for _, name := range []string{"report.Summary.Satisfied", "report.Summary.ValidCases",
		"report.Summary.Unresolved", "languagesyntax.FixedTotal", "languagesyntax.FixedCapabilityTotal",
		"len(report.Source.GoooFiles)"} {
		if seen[name] == 0 {
			return nil, fmt.Errorf("syntax conformance obligation not found: %s", name)
		}
	}
	return source.finish()
}
