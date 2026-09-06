package syntaxregistration

import (
	"fmt"
	"go/ast"
)

func generateSelection(raw []byte, version, capability int) ([]byte, error) {
	source, err := parseGo(raw)
	if err != nil {
		return nil, err
	}
	for _, item := range []struct{ function, result string }{
		{"activeDenominator", evidenceName(version)},
		{"activeDenominatorDigest", digestName(version)},
	} {
		function, err := source.function(item.function)
		if err != nil {
			return nil, err
		}
		count := 0
		ast.Inspect(function, func(node ast.Node) bool {
			selection, ok := node.(*ast.SwitchStmt)
			if !ok || selection.Tag == nil || source.text(selection.Tag) != "languagesyntax.FixedCapabilityTotal" {
				return true
			}
			matched := false
			for _, statement := range selection.Body.List {
				clause, ok := statement.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expression := range clause.List {
					value, literal := integer(expression)
					if literal && value == capability+1 {
						count = -100
						return false
					}
					if literal && value == capability {
						matched = true
					}
				}
			}
			if matched {
				source.insert(selection.Body.Rbrace, fmt.Sprintf("\ncase %d:\nreturn %s\n", capability+1, item.result))
				count++
			}
			return false
		})
		if count != 1 {
			return nil, fmt.Errorf("denominator selection %s lacks one exact baseline", item.function)
		}
	}
	source.edits = append(source.edits, sourceEdit{len(raw), len(raw),
		fmt.Sprintf("\n//go:embed evidence/denominator-v%d.json\nvar %s []byte\n", version, evidenceName(version))})
	return source.finish()
}
