package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func generateSyntaxTests(source *goSource, corpus []byte, promoteMetaSource bool) error {
	total, valid, capability, err := corpusTotals(corpus)
	if err != nil {
		return err
	}
	expected := map[string]int{
		"report.Summary.Satisfied": total, "report.Summary.ValidCases": valid,
		"report.Summary.Unresolved": total, "report.Summary.Total": total,
		"languagesyntax.FixedTotal": total, "languagesyntax.FixedCapabilityTotal": capability,
	}
	seen := map[string]int{}
	var failure error
	ast.Inspect(source.file, func(node ast.Node) bool {
		if failure != nil {
			return false
		}
		comparison, ok := node.(*ast.BinaryExpr)
		if !ok || (comparison.Op != token.NEQ && comparison.Op != token.EQL) {
			return true
		}
		name := strings.ReplaceAll(source.text(comparison.X), " ", "")
		if name == "len(report.Source.GoooFiles)" {
			failure = reviseInventoryComparison(source, comparison, promoteMetaSource)
			if failure == nil {
				seen[name]++
			}
			return true
		}
		actual, literal := integer(comparison.Y)
		if !literal {
			return true
		}
		if old, wanted := expected[name]; wanted && actual == old {
			source.replace(comparison.Y, strconv.Itoa(old+1))
			seen[name]++
		}
		return true
	})
	if failure != nil {
		return failure
	}
	for _, name := range []string{"report.Summary.Satisfied", "report.Summary.ValidCases",
		"report.Summary.Unresolved", "languagesyntax.FixedTotal", "languagesyntax.FixedCapabilityTotal",
		"len(report.Source.GoooFiles)"} {
		if seen[name] == 0 {
			return fmt.Errorf("syntax conformance obligation not found: %s", name)
		}
	}
	return nil
}

// Declared inventory follows registry membership automatically. Historical
// literal fixtures retain their existing append/promotion transformation.
func reviseInventoryComparison(source *goSource, comparison *ast.BinaryExpr, promoteMetaSource bool) error {
	if actual, literal := integer(comparison.Y); literal {
		if actual <= 0 {
			return fmt.Errorf("source inventory baseline is not positive")
		}
		if !promoteMetaSource {
			actual++
		}
		source.replace(comparison.Y, strconv.Itoa(actual))
		return nil
	}
	call, ok := comparison.Y.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 || call.Ellipsis.IsValid() ||
		source.text(call.Fun) != "languagesyntax.DeclaredSourceCount" {
		return fmt.Errorf("source inventory requires a declared boundary independent of observations")
	}
	return source.requireImport(comparison,
		"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax", "languagesyntax")
}
